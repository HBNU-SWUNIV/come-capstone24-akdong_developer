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
    imagePath := "/CarteTest/image/" + imageName
    imageTarPath := imagePath + ".tar"
    containerPath := "/CarteTest/container/" + containerName

    // 이미지 압축 해제
    if _, err := os.Stat(imageTarPath); err == nil {
        fmt.Println("Found tar file...")
        if _, err := os.Stat(imagePath); os.IsNotExist(err) {
            if err := os.MkdirAll(imagePath, 0755); err != nil {
                log.Fatalf("Failed to create image directory: %v", err)
            }
        }
        if err := extractTar(imageTarPath, imagePath); err != nil {
            log.Fatalf("Failed to extract tar file: %v", err)
        }
    } else if _, err := os.Stat(imagePath); os.IsNotExist(err) {
        log.Fatalf("Image directory not found: %s", imagePath)
    } else if err != nil {
        log.Fatalf("Failed to check image directory: %v", err)
    }

    // 이건 그냥 컨테이너 경로 확인 - 컨테이너 경로를 이렇게 설정해도 되는가?
    if _, err := os.Stat(containerPath); os.IsNotExist(err) {
        if err := os.Mkdir(containerPath, 0755); err != nil && !os.IsExist(err) {
            log.Fatalf("Failed to create container directory: %v", err)
        }
    } else {
        fmt.Printf("Container %s already exists\n", containerName)
        StartContainer(containerPath)
        return
    }

    if err := pivotRoot(containerPath, imagePath); err != nil {
    // if err := pivotRoot(imagePath, containerPath); err != nil {
        log.Fatalf("Failed to pivot root: %v", err)
    }

    if err := setupCgroups(); err != nil {
        log.Fatalf("Failed to set up cgroups: %v", err)
    }

    runInNewNamespace(containerPath, "/usr/sbin/nginx", []string{"-g", "daemon off;"})
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
func pivotRoot(newRoot, oldRoot string) error {
    // 새로운 루트 디렉토리와 기존 루트 디렉토리 생성


    // Bind 마운트: 새로운 루트 디렉토리를 마운트
    if err := mountBind(newRoot, newRoot); err != nil {
        return fmt.Errorf("failed to bind mount new root: %v", err)
    }

    if err := os.MkdirAll(oldRoot, 0700); err != nil {
        return fmt.Errorf("failed to create old root directory: %v", err)
    }

    // pivot_root 시스템 호출로 루트 파일 시스템 변경
    if err := syscall.PivotRoot(newRoot, oldRoot); err != nil {
        return fmt.Errorf("failed to pivot root: %v", err)
    }

    // 문제 : PivotRoot전에 newRoot생성을 하는 바람에 컨테이너 생성으로 넘어가서 실행이 안됐었음(무한 대기), 순서 변경으로 해결
    // 생성 문제 아님 : failed to pivot root: invalid argument 문제
    if err := os.MkdirAll(newRoot, 0700); err != nil {
        return fmt.Errorf("failed to create new root directory: %v", err)
    }

    // 루트 변경 후 새로운 루트로 작업 디렉토리 이동
    if err := os.Chdir("/"); err != nil {
        return fmt.Errorf("failed to change directory to new root: %v", err)
    }

    // 기존 루트 파일 시스템을 언마운트
    if err := syscall.Unmount(oldRoot, syscall.MNT_DETACH); err != nil {
        return fmt.Errorf("failed to unmount old root: %v", err)
    }

    // 기존 루트 디렉토리 삭제
    if err := os.RemoveAll(oldRoot); err != nil {
        return fmt.Errorf("failed to remove old root directory: %v", err)
    }

    return nil
}

// Bind 마운트 함수
func mountBind(target, mountPoint string) error {
    // 마운트 시스템 호출 사용
    if err := syscall.Mount(target, mountPoint, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
        return fmt.Errorf("failed to bind mount: %v", err)
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
