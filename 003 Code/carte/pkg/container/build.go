package container

import (
    "bufio"
    "crypto/sha256"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    //"archive/tar"
    //"compress/gzip"
)

// BuildContainer는 주어진 Cartefile과 이미지 이름을 기반으로 새로운 이미지를 빌드하는 함수입니다.
func BuildContainer(cartefilePath, imageName string) error {
    fmt.Println("Cartefile을 읽고 있습니다...")

    // 1. Cartefile 읽기
    file, err := os.Open(cartefilePath)
    if err != nil {
        return fmt.Errorf("Cartefile을 열 수 없습니다: %v", err)
    }
    defer file.Close()

    // 임시 작업 디렉토리 생성
    tempDir, err := ioutil.TempDir("", "carte_build_")
    if err != nil {
        return fmt.Errorf("임시 디렉토리를 만들 수 없습니다: %v", err)
    }
    defer os.RemoveAll(tempDir) // 빌드가 끝나면 삭제

    scanner := bufio.NewScanner(file)
    var baseImage string
    var commands []string

    // 2. Cartefile의 각 줄을 처리
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        if strings.HasPrefix(line, "FROM") {
            baseImage = strings.TrimSpace(strings.TrimPrefix(line, "FROM"))
            fmt.Println("베이스 이미지:", baseImage)
        } else if strings.HasPrefix(line, "RUN") || strings.HasPrefix(line, "COPY") || strings.HasPrefix(line, "CMD") {
            commands = append(commands, line)
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("Cartefile 읽기 중 오류 발생: %v", err)
    }

    // 3. 베이스 이미지 확인 및 준비
    baseImagePath := filepath.Join("/var/run/carte/images", baseImage)
    if _, err := os.Stat(baseImagePath); os.IsNotExist(err) {
        return fmt.Errorf("베이스 이미지가 로컬에 없습니다: %s", baseImage)
    }
    fmt.Println("베이스 이미지가 준비되었습니다:", baseImage)

    // 4. 베이스 이미지 복사
    imagePath := filepath.Join("/var/run/carte/images", imageName)
    rootFsPath := filepath.Join(imagePath, "rootfs")
    os.MkdirAll(rootFsPath, 0755) // rootfs 디렉토리 생성

    fmt.Println("베이스 이미지 복사 중:", baseImagePath, "->", rootFsPath)
    if err := copyDirectory(baseImagePath, rootFsPath); err != nil {
        return fmt.Errorf("베이스 이미지 복사 중 오류 발생: %v", err)
    }

    // 5. 명령어 처리 (RUN, COPY, CMD)
    for i, cmd := range commands {
        if strings.HasPrefix(cmd, "RUN") {
            runCommand := strings.TrimSpace(strings.TrimPrefix(cmd, "RUN"))
            if err := handleLayer(runCommand, rootFsPath, "", imageName, i); err != nil {
                return fmt.Errorf("RUN 명령어 실행 중 오류 발생: %v", err)
            }
        } else if strings.HasPrefix(cmd, "COPY") {
            copyParts := strings.Fields(strings.TrimPrefix(cmd, "COPY"))
            if len(copyParts) != 2 {
                return fmt.Errorf("COPY 명령어 형식이 잘못되었습니다: %s", cmd)
            }
            src := copyParts[0]
            dest := filepath.Join(rootFsPath, copyParts[1]) // rootfs로 직접 복사
            if err := copyFile(src, dest); err != nil {
                return fmt.Errorf("COPY 중 오류 발생: %v", err)
            }
        } else if strings.HasPrefix(cmd, "CMD") {
            // CMD 명령어 처리 (이미지 메타데이터에 저장)
            cmdStr := strings.TrimSpace(strings.TrimPrefix(cmd, "CMD"))
            cmdStr = strings.Trim(cmdStr, "[]") // ["python3", "/app.py"] 형태 제거
            fmt.Println("CMD 명령어 설정:", cmdStr)
        }
    }

    fmt.Println("이미지 빌드 완료:", imagePath)
    return nil
}



// handleLayer는 명령어 실행 전 캐시를 확인하고, 캐시가 없으면 새로운 레이어를 생성합니다.
func handleLayer(command, rootFsPath, layerDir, imageName string, layerIndex int) error {
    // 1. 명령어 해시를 생성하여 캐시가 존재하는지 확인
    layerHash := hashCommand(command)
    cachedLayerPath := filepath.Join("/var/run/carte/cache", layerHash)

    if _, err := os.Stat(cachedLayerPath); err == nil {
        fmt.Println("캐시된 레이어 사용:", cachedLayerPath)
        return nil
    }

    // 2. 캐시된 레이어가 없으면 새로운 레이어 생성
    fmt.Println("새로운 레이어 생성:", command)
    
    // 여기서 layerDir 제거
    if err := runInChroot(command, rootFsPath, layerIndex); err != nil {
        return fmt.Errorf("명령어 실행 중 오류: %v", err)
    }

    // 3. 레이어 저장
    newLayerPath := filepath.Join(layerDir, fmt.Sprintf("layer_%d", layerIndex))
    if err := saveLayerDiff(rootFsPath, newLayerPath); err != nil {
        return fmt.Errorf("레이어 저장 중 오류 발생: %v", err)
    }

    return nil
}


// 레이어 병합 함수
func mergeLayers(layerDir, imagePath string) error {
    fmt.Println("레이어 병합 중:", layerDir)

    // 이미지의 최종 루트 파일 시스템을 생성
    finalFsPath := filepath.Join(imagePath, "rootfs")
    os.MkdirAll(finalFsPath, 0755)

    // 레이어 디렉토리 내 모든 레이어 병합
    err := filepath.Walk(layerDir, func(file string, fi os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if fi.IsDir() {
            return nil
        }

        relPath, err := filepath.Rel(layerDir, file)
        if err != nil {
            return fmt.Errorf("파일 경로 변환 중 오류 발생: %v", err)
        }

        destPath := filepath.Join(finalFsPath, relPath)
        fmt.Printf("병합 중: %s -> %s\n", file, destPath)

        if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
            return fmt.Errorf("상위 디렉토리 생성 중 오류 발생: %v", err)
        }

        return copyFile(file, destPath)
    })

    if err != nil {
        return fmt.Errorf("레이어 병합 중 오류 발생: %v", err)
    }

    fmt.Println("레이어 병합 완료")
    return nil
}


// runInChroot는 chroot 환경에서 명령어를 실행합니다.
func runInChroot(command, rootFsPath string, layerIndex int) error {
    fmt.Println("RUN 명령어 실행 중 (chroot):", command)
    cmd := exec.Command("chroot", rootFsPath, "sh", "-c", command)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    fmt.Println("명령어 실행 시작")
    err := cmd.Run()
    fmt.Println("명령어 실행 완료")
    return err
}

// copyDirectory는 베이스 이미지를 임시 작업 디렉토리로 복사하는 함수입니다.
func copyDirectory(src, dest string) error {
    fmt.Printf("exec.Command로 디렉토리 복사: %s -> %s\n", src, dest)
    
    // rsync를 사용하여 디렉토리를 복사
    cmd := exec.Command("rsync", "-a", "--delete", "--verbose", "--exclude", "/proc", "--exclude", "/sys", "--exclude", "/dev", src+"/", dest+"/")
    
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("rsync 명령어 실행 중 오류 발생: %v", err)
    }
    
    fmt.Println("rsync를 사용한 디렉토리 복사 완료")
    return nil
}



// copyFile은 src 파일을 dest로 복사합니다.
func copyFile(src, dest string) error {
    srcFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    destFile, err := os.Create(dest)
    if err != nil {
        return err
    }
    defer destFile.Close()

    if _, err = io.Copy(destFile, srcFile); err != nil {
        return err
    }

    srcInfo, err := os.Stat(src)
    if err != nil {
        return err
    }
    return os.Chmod(dest, srcInfo.Mode())
}



// saveLayerDiff는 파일 시스템의 변화를 레이어로 저장합니다.
func saveLayerDiff(rootFsPath, layerDir string) error {
    fmt.Println("rsync 명령어 실행 시작")
    
    // rsync에서 /proc, /sys, /dev 등의 가상 파일 시스템 디렉토리를 제외
    cmd := exec.Command("rsync", "-a", "--delete", "--partial", "--verbose",
        "--exclude", "/proc", "--exclude", "/sys", "--exclude", "/dev",
        rootFsPath+"/", layerDir+"/")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    err := cmd.Run()
    if err != nil {
        fmt.Printf("rsync 명령어 실행 중 오류 발생: %v\n", err)
        return err
    }
    
    fmt.Println("rsync 명령어 실행 완료")
    return nil
}

// hashCommand는 명령어를 해싱하여 캐시 키로 사용합니다.
func hashCommand(command string) string {
    hash := sha256.New()
    hash.Write([]byte(command))
    return fmt.Sprintf("%x", hash.Sum(nil))
}
