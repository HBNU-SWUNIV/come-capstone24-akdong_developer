package main

import (
    "archive/tar"
    "fmt"
    "io"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"
    "bufio"
    "strings"
    "golang.org/x/sys/unix"
)

func CtCreate(imageName string, containerName string) {
    imageTarPath := "/CarteTest/image/" + imageName + ".tar" // tar 이미지 경로
    containerPath := "/CarteTest/container/" + containerName

    // 이미지 압축 해제
    if _, err := os.Stat(imageTarPath); err == nil {
        fmt.Println("Found tar file...")
        if _, err := os.Stat(containerPath); os.IsNotExist(err) {
            if err := os.MkdirAll(containerPath, 0755); err != nil {
                log.Fatalf("Failed to create container directory: %v", err)
            }
        }
        if err := extractTar(imageTarPath, containerPath); err != nil {
            log.Fatalf("Failed to extract tar file: %v", err)
        }
        if err := prepareContainer(containerPath); err != nil {
            fmt.Println("Error:", err)
        } else {
            fmt.Println("Container prepared successfully")
        }
    } else {
        log.Fatalf("Image file not found: %s", imageTarPath)
    }
    if err := StartContainer("/CarteTest/container/testcontainer"); err != nil {
        log.Fatal(err)
    }
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

func mountBind(source, target string) error {
    return syscall.Mount(source, target, "", syscall.MS_BIND|syscall.MS_REC, "")
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

// cgroups 설정
func setupCgroups() error {
    cgroupsPath := "/CarteTest/cgroup" // 사용자 지정 경로
    if err := os.MkdirAll(cgroupsPath, 0755); err != nil {
        return fmt.Errorf("failed to create cgroup path: %v", err)
    }

    // CPU 제한 설정 예시
    cpuPath := filepath.Join(cgroupsPath, "cpu/testcontainer")
    if err := os.MkdirAll(cpuPath, 0755); err != nil {
        return fmt.Errorf("failed to create CPU cgroup path: %v", err)
    }
    if err := os.WriteFile(filepath.Join(cpuPath, "cpu.shares"), []byte("512"), 0644); err != nil {
        return fmt.Errorf("failed to set CPU shares: %v", err)
    }

    return nil
}

// 새로운 네임스페이스에서 명령어 실행
func runInNewNamespace(containerPath, path string, args []string) error {
    fullPath := filepath.Join(containerPath, path)
    fmt.Printf("Before chroot, checking fullPath: %s\n", fullPath)
    if _, err := os.Stat(fullPath); err != nil {
        return fmt.Errorf("before chroot: fullPath not found: %v", err)
    }

    if err := syscall.Chroot(containerPath); err != nil {
        return fmt.Errorf("failed to apply chroot to container path: %v", err)
    }
    fmt.Println("Chroot applied successfully")

    if _, err := os.Stat(path); err != nil {
        return fmt.Errorf("after chroot: path not found: %v", err)
    }

    if err := os.Chdir("/"); err != nil {
        return fmt.Errorf("failed to change to new root directory: %v", err)
    }

    cmd := exec.Command(path, args...)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
    }
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    fmt.Println("Starting command execution")
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to run command in new namespace: %v", err)
    }
    fmt.Println("Command execution finished")

    return nil
}


func StartContainer(containerPath string) error {
    return runInNewNamespace(containerPath, "/bin/busybox", []string{"httpd", "-f", "-p", "8080"})
}

// func main() {
//     if err := StartContainer("/CarteTest/container/testcontainer"); err != nil {
//         log.Fatal(err)
//     }
// }

func main() {
    CtCreate("busybox", "testcontainer")
}
