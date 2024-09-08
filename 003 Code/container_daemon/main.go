package main

import (
    "archive/tar"
    "fmt"
    "io"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "bufio"
    "strings"
    "golang.org/x/sys/unix"
    "syscall" 
    "time"
)

// 1. 지금 CtCreate 넘어와서 StartContainer 하면 안되고(no such file or directory) 바로 StartContainer 실행하면 됨(원인 파악 필요)
// --> 컨테이너 이미 존재하는 경우 또 생성하면서 오류 발생(if 위치가 잘못 명시되어 있었음), if문제가 아닌, 컨테이너 생성 후 바로 실행 불가에 따른 문제 였음

// 컨테이너 생성(이미지 압축 해제)
func CtCreate(imageName string, containerName string) error {
    imageTarPath := "/CarteTest/image/" + imageName + ".tar"
    containerPath := "/CarteTest/container/" + containerName

    // 이미지 압축 해제
    if _, err := os.Stat(imageTarPath); err == nil {
        fmt.Println("Found tar file...")
        if _, err := os.Stat(containerPath); os.IsNotExist(err) {
            if err := os.MkdirAll(containerPath, 0755); err != nil {
                return fmt.Errorf("failed to create container directory: %v", err)
            }
            if err := extractTar(imageTarPath, containerPath); err != nil {
                return fmt.Errorf("failed to extract tar file: %v", err)
            }
            if err := prepareContainer(containerPath); err != nil {
                return fmt.Errorf("error preparing container: %v", err)
            }
            fmt.Println("Container created successfully")
        } else {
            fmt.Println("Container already prepared")
        }
    } else {
        return fmt.Errorf("image file not found: %s", imageTarPath)
    }

    return nil
}

// tar 파일 압축 해제
func extractTar(tarFile, destDir string) error {
    file, err := os.Open(tarFile)
    if err != nil {
        return fmt.Errorf("failed to open tar file: %v", err)
    }
    defer file.Close()

    tarReader := tar.NewReader(file)
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("failed to read tar file: %v", err)
        }

        targetPath := filepath.Join(destDir, header.Name)
        switch header.Typeflag {
        case tar.TypeDir:
            if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
                return fmt.Errorf("failed to create directory: %v", err)
            }
        case tar.TypeReg:
            outFile, err := os.Create(targetPath)
            if err != nil {
                return fmt.Errorf("failed to create file: %v", err)
            }
            if _, err := io.Copy(outFile, tarReader); err != nil {
                outFile.Close()
                return fmt.Errorf("failed to write file: %v", err)
            }
            outFile.Close()

            // tar 헤더에서 파일 권한 설정
            if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
                return fmt.Errorf("failed to set file permissions: %v", err)
            }
        default:
            log.Printf("Unknown type: %v in %s", header.Typeflag, header.Name)
        }
    }
    return nil
}

// 루트 파일 시스템 변경 함수
func prepareContainer(newRoot string) error {
    // 새로운 루트로 변경 (chroot)

    fmt.Printf("checking path: %s\n", newRoot)
    if _, err := os.Stat(newRoot); err != nil {
        return fmt.Errorf("command not found or inaccessible: %v", err) // 오류 메시지 수정 및 오류 반환
    }

    if err := syscall.Chroot(newRoot); err != nil {
        return fmt.Errorf("failed to chroot to new root: %v", err)
    }

    // 루트 디렉토리로 이동
    if err := os.Chdir("/"); err != nil {
        return fmt.Errorf("failed to change directory to new root: %v", err)
    }

    // 시스템 디렉토리 설정 (proc, sys, dev 마운트)
    if err := setupSystemDirs(); err != nil {
        return err
    }

    // /dev/null 및 기타 장치 파일 생성
    if err := createDevFiles(); err != nil {
        return err
    }

    return nil
}

func setupSystemDirs() error {
    // /proc 마운트 확인 후 마운트
    if !isMounted("/proc") {
        if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
            return fmt.Errorf("failed to mount /proc: %v", err)
        }
    }

    // /sys 마운트 확인 후 마운트
    if !isMounted("/sys") {
        if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
            return fmt.Errorf("failed to mount /sys: %v", err)
        }
    }

    // /dev 마운트 확인 후 마운트
    if !isMounted("/dev") {
        if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
            return fmt.Errorf("failed to mount /dev: %v", err)
        }
    }
    return nil
}

// 마운트 여부 확인 함수
func isMounted(target string) bool {
    file, err := os.Open("/proc/mounts")
    if err != nil {
        return false
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        if strings.Contains(scanner.Text(), target) {
            return true
        }
    }
    return false
}

func createDevFiles() error {
    // /dev/null 생성
    if err := syscall.Mknod("/dev/null", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 3))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/null: %v", err)
    }

    // /dev/zero 생성
    if err := syscall.Mknod("/dev/zero", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 5))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/zero: %v", err)
    }

    // /dev/random 생성
    if err := syscall.Mknod("/dev/random", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 8))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/random: %v", err)
    }

    // /dev/urandom 생성
    if err := syscall.Mknod("/dev/urandom", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 9))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/urandom: %v", err)
    }

    // /dev/tty 생성
    if err := syscall.Mknod("/dev/tty", syscall.S_IFCHR|0666, int(unix.Mkdev(5, 0))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/tty: %v", err)
    }

    // /dev/console 생성
    if err := syscall.Mknod("/dev/console", syscall.S_IFCHR|0622, int(unix.Mkdev(5, 1))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/console: %v", err)
    }

    return nil
}

// 새로운 네임스페이스에서 명령어 실행
func runInNewNamespace(containerPath, path string, args []string) error {
    // chroot 전 경로 확인
    fullPath := filepath.Join(containerPath, path)
    fmt.Printf("Before chroot, checking path: %s\n", fullPath)
    if _, err := os.Stat(fullPath); err != nil {
        return fmt.Errorf("before chroot: command not found: %v", err)
    }

    // chroot 적용
    if err := syscall.Chroot(containerPath); err != nil {
        return fmt.Errorf("failed to apply chroot to container path: %v", err)
    }
    fmt.Println("Chroot applied successfully")

    // chroot 적용 후에는 더 이상 fullPath를 사용할 수 없음
    if _, err := os.Stat(path); err != nil {
        return fmt.Errorf("after chroot: command not found: %v", err)
    }

    // 루트 디렉토리로 이동
    if err := os.Chdir("/"); err != nil {
        return fmt.Errorf("failed to change to new root directory: %v", err)
    }

    // 명령 실행
    cmd := exec.Command(path, args...)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
    }
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 명령 실행
    fmt.Println("Starting command execution")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to run command in new namespace: %v", err)
    }
    fmt.Println("Command execution finished")

    return nil
}

func startContainer(containerPath string) error {
    return runInNewNamespace(containerPath, "/bin/busybox", []string{"httpd", "-f", "-p", "8080"})
}

func setupCgroups(containerPath string) error {
    cgroupRoot := "/CarteTest/cgroup"
    pid := os.Getpid()

    if err := os.MkdirAll(cgroupRoot, 0755); err != nil {
        return fmt.Errorf("failed to create cgroup path: %v", err)
    }

    // CPU 리소스 설정
    cpuLimitPath := filepath.Join(cgroupRoot, "cpu", "myContainerGroup")
    if err := os.MkdirAll(cpuLimitPath, 0755); err != nil {
        return fmt.Errorf("failed to create cgroup for cpu: %v", err)
    }

    // CPU 쿼터 설정 (100000us = 100ms every 100ms)
    if err := os.WriteFile(filepath.Join(cpuLimitPath, "cpu.cfs_quota_us"), []byte("100000"), 0644); err != nil {
        return fmt.Errorf("failed to set cpu quota: %v", err)
    }

    // 현재 프로세스(컨테이너 프로세스)를 새 cgroup에 추가
    if err := os.WriteFile(filepath.Join(cpuLimitPath, "tasks"), []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
        return fmt.Errorf("failed to add process to cgroup: %v", err)
    }

    return nil
}

func main() {
    containerName := "testcontainer"
    imageName := "busybox"
    containerPath := "/CarteTest/container/" + containerName

    // 컨테이너 환경 설정
    if err := CtCreate(imageName, containerName); err != nil {
        log.Fatal("Failed to create container:", err)
    }

    // cgroups 설정
    if err := setupCgroups(containerPath); err != nil {
        log.Fatal("Failed to setup cgroups:", err)
    }

    // 파일 시스템 동기화 강화
    syscall.Sync()  // 동기화 호출
    time.Sleep(2 * time.Second)  // 동기화가 시스템에 반영될 시간을 기다림 // --> 컨테이너 생성후, 실행까지 바로 적용이 안됨(CLI 제작을 통해서 함수 나눈거대로 생성, 실행 바꾸기)

    // 컨테이너 실행
    fmt.Println("Attempting to start container...")
    if err := startContainer(containerPath); err != nil {
        log.Fatal("Failed to start container:", err)
    } else {
        fmt.Println("Container started successfully. Checking port binding...")
        time.Sleep(5 * time.Second)  // 시간 지연 후 포트 체크
        cmd := exec.Command("sh", "-c", "netstat -tulnp | grep :8080 || ss -tulnp | grep :8080")
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        if err := cmd.Run(); err != nil {
            fmt.Println("Failed to check port binding:", err)
        }
    }
}


