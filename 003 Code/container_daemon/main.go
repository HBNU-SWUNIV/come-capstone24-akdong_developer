package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)


// 2024/09/02 22:15:31 Unknown type: 50 in var/run
// 2024/09/02 22:15:31 Unknown type: 50 in var/spool/mail
// Container nginxtest already exists
// Container started successfully
// 오류 해결 필요

func CtCreate(imageName string, containerName string) {
	// 이미지와 컨테이너 경로 설정
	imagePath := "/CarteTest/image/" + imageName
	imageTarPath := imagePath + ".tar.gz"
	containerPath := "/CarteTest/container/" + containerName

	// 이미지 압축 해제
	if _, err := os.Stat(imageTarPath); err == nil {
		fmt.Println("Found tar.gz file...")
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			if err := os.MkdirAll(imagePath, 0755); err != nil {
				log.Fatalf("Failed to create image directory: %v", err)
			}
		}
		if err := extractTarGz(imageTarPath, imagePath); err != nil {
			log.Fatalf("Failed to extract tar file: %v", err)
		}
	} else if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Fatalf("Image directory not found: %s", imagePath)
	} else if err != nil {
		log.Fatalf("Failed to check image directory: %v", err)
	}

	// 컨테이너 경로 설정
	if _, err := os.Stat(containerPath); os.IsNotExist(err) {
		if err := os.Mkdir(containerPath, 0755); err != nil && !os.IsExist(err) {
			log.Fatalf("Failed to create container directory: %v", err)
		}
	} else {
		fmt.Printf("Container %s already exists\n", containerName)
		StartContainer(containerPath)
		return
	}

	// 루트 파일 시스템 구성
	pivotRoot(imagePath, containerPath)

	// cgroups 설정
	if err := setupCgroups(); err != nil {
		log.Fatalf("Failed to set up cgroups: %v", err)
	}

	// 컨테이너 내에서 프로세스 실행
	cmd := &exec.Cmd{
		Path: "/usr/sbin/nginx",
		Args: []string{"/usr/sbin/nginx", "-g", "daemon off;"},
		SysProcAttr: &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
		Dir:    "/",
	}

	// 로그 파일 생성
	logFile, err := os.Create(filepath.Join(containerPath, "container_output.log"))
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start the command: %v", err)
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Fatalf("Command execution failed: %v", err)
		}
		fmt.Println("Process has exited.")
	}()

	fmt.Printf("Container %s created and started successfully!\n", containerName)
}

// tar.gz 파일 압축 해제
func extractTarGz(tarGzFile, destDir string) error {
	file, err := os.Open(tarGzFile)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %v", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
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

func pivotRoot(newRoot, containerPath string) {
	oldRoot := filepath.Join(containerPath, ".pivot_root_old")

	// 이전 루트 디렉토리 생성
	if err := os.MkdirAll(oldRoot, 0700); err != nil {
		log.Fatalf("Failed to create old root directory: %v", err)
	}

	// 새로운 루트 디렉토리 생성
	if err := os.MkdirAll("/new-root", 0700); err != nil {
		log.Fatalf("Failed to create new root directory: %v", err)
	}

	// 새로운 루트 디렉토리 마운트
	if err := mountBind("/new-root", newRoot); err != nil {
		log.Fatalf("Failed to bind mount new root: %v", err)
	}

	// 루트 변경
	if err := syscall.PivotRoot(newRoot, oldRoot); err != nil {
		log.Fatalf("Failed to pivot root: %v", err)
	}

	// 새로운 루트로 이동
	if err := os.Chdir("/"); err != nil {
		log.Fatalf("Failed to change directory to new root: %v", err)
	}

	// 이전 루트 언마운트 및 삭제
	if err := syscall.Unmount(oldRoot, syscall.MNT_DETACH); err != nil {
		log.Fatalf("Failed to unmount old root: %v", err)
	}
	if err := os.RemoveAll(oldRoot); err != nil {
		log.Fatalf("Failed to remove old root directory: %v", err)
	}
}

// mountBind 함수를 추가하여 mount 명령어를 호출
func mountBind(mountPoint, target string) error {
	cmd := exec.Command("mount", "--bind", target, mountPoint)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to execute mount command: %v", err)
	}
	return nil
}

// cgroups 설정
func setupCgroups() error {
	cgroups := "/sys/fs/cgroup/"
	pid := os.Getpid()

	// 메모리 cgroup 설정
	if err := os.MkdirAll(filepath.Join(cgroups, "memory", "carte"), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cgroups, "memory", "carte", "memory.limit_in_bytes"), []byte("104857600"), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cgroups, "memory", "carte", "cgroup.procs"), []byte(strconv.Itoa(pid)), 0700); err != nil {
		return err
	}

	// CPU cgroup 설정
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

// 컨테이너 실행
func StartContainer(containerPath string) {
	cmd := &exec.Cmd{
		Path: "/usr/sbin/nginx",
		Args: []string{"/usr/sbin/nginx", "-g", "daemon off;"},
		SysProcAttr: &syscall.SysProcAttr{
			Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
		},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
		Dir:    "/",
	}

	logFile, err := os.Create(filepath.Join(containerPath, "container_output.log"))
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start the command in existing container: %v", err)
	}

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Fatalf("Command execution failed: %v", err)
		}
		fmt.Println("Process in existing container has exited")
	}()

	fmt.Println("Container started successfully")
}

func main() {
	CtCreate("nginx", "nginxtest")
}
