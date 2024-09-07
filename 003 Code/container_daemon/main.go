package main

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

// 컨테이너 생성
func CtCreate(imageName string, containerName string) {
    // imagePath := "/CarteTest/image/" + imageName  --> 압축 해제 위치 잘못됨
    imageTarPath := "/CarteTest/image/" + imageName + ".tar" // tar 이미지 경로
    containerPath := "/CarteTest/container/" + containerName
    oldRootPath := containerPath + "/old-root"

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
        if err := prepareContainer(containerPath, oldRootPath); err != nil {
            fmt.Println("Error:", err)
        } else {
            fmt.Println("Container prepared successfully")
        }
    } else if os.IsNotExist(err) {
        log.Fatalf("Image file not found: %s", imageTarPath)
    } else {
        log.Fatalf("Error checking image file: %v", err)
    }

    // runInNewNamespace(containerPath, "/usr/sbin/nginx", []string{"-g", "daemon off;"})
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
		default:
			log.Printf("Unknown type: %v in %s", header.Typeflag, header.Name)
		}
	}
	return nil
}

// 루트 파일 시스템 변경 함수
func prepareContainer(newRoot, oldRoot string) error {
    if err := os.MkdirAll(oldRoot, 0755); err != nil {
        return fmt.Errorf("failed to create old root directory: %v", err)
    }
    if err := mountBind(newRoot, newRoot); err != nil {
        return fmt.Errorf("failed to bind mount new root: %v", err)
    }
    if err := syscall.PivotRoot(newRoot, oldRoot); err != nil {
        return fmt.Errorf("failed to pivot root: %v", err)
    }
    if err := os.Chdir("/"); err != nil {
        return fmt.Errorf("failed to change directory to new root: %v", err)
    }
    if err := setupSystemDirs(); err != nil {
        return err
    }
    return nil
}

func mountBind(source, target string) error {
    return syscall.Mount(source, target, "", syscall.MS_BIND|syscall.MS_REC, "")
}

func setupSystemDirs() error {
    if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
        return fmt.Errorf("failed to mount /proc: %v", err)
    }
    if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
        return fmt.Errorf("failed to mount /sys: %v", err)
    }
    if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
        return fmt.Errorf("failed to mount /dev: %v", err)
    }
    return nil
}

// cgroups 설정
func setupCgroups() error {
	cgroups := "/sys/fs/cgroup/"
	pid := os.Getpid()

	if err := os.MkdirAll(filepath.Join(cgroups, "memory", "carte"), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cgroups, "memory", "carte", "memory.limit_in_bytes"), []byte("104857600"), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cgroups, "memory", "carte", "cgroup.procs"), []byte(strconv.Itoa(pid)), 0700); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Join(cgroups, "cpu", "carte"), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cgroups, "cpu", "carte", "cpu.cfs_quota_us"), []byte("50000"), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cgroups, "cpu", "carte", "cgroup.procs"), []byte(strconv.Itoa(pid)), 0700); err != nil {
		return err
	}

	return nil
}

// 새로운 네임스페이스에서 명령어 실행
func runInNewNamespace(containerPath, path string, args []string) {
    cmd := exec.Command(path, args...)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
    }
    cmd.Dir = containerPath

    logFile, err := os.Create(filepath.Join(containerPath, "container_output.log"))
    if err != nil {
        log.Fatalf("Failed to create log file: %v", err)
    }
    defer logFile.Close()

    cmd.Stdout = logFile
    cmd.Stderr = logFile

    if err := cmd.Start(); err != nil {
        log.Fatalf("Failed to start command: %v", err)
    }

    if err := cmd.Wait(); err != nil {
        log.Fatalf("Command execution failed: %v", err)
    }

    fmt.Println("Process has exited.")
}

// 컨테이너 실행
func StartContainer(containerPath string) {
	runInNewNamespace(containerPath, "/usr/sbin/nginx", []string{"-g", "daemon off;"})
}

func main() {
	CtCreate("nginx", "nginxtest")
}
