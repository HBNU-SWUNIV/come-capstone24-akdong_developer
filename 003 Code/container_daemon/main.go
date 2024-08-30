package main

import(
	"archive/tar"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"strconv"
)

/* 구현 된 것 
daemon 실행하고 client 실행 하면, 연결 되는 로직 존재(Carte)
Carte CLI 생성 로직 존재(권한 설정, 이미지 생성을 위한 자동 경로 설정)
*/

/* 우선 필요한 것(되도록 이번주까지 구현할 것)
1. 기존 이미지의 container 생성, 삭제, 재시작 처리 로직
2. 플러그인 없이 네트워크 구성하기(오픈소스 활용한 라인 분석 : 커스터마이징 하도록)
*/

// 기존 이미지로 container 생성하기
func CtCreate(imageName string, containerName string) {

	// 컨테이너 생성(Carte run <이미지 이름>)

	// 이미지 경로 확인
	imagePath := "/CarteTest/image/" + imageName
	imageTarPath := imagePath + ".tar"
	containerPath := "/CarteTest/container/" + containerName

	// 이미지가 tar 파일인지 확인하고 해제
	if _, err := os.Stat(imageTarPath); err == nil{
		fmt.Println("Found tar file...")

		// /CarteTest/image/testimage __ tar인경우 폴더가 없으면 생성
		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			if err := os.MkdirAll(imagePath, 0755); err != nil {
				log.Fatalf("Failed to create image directory: %v", err)
			}
		}

		err := extractTar(imageTarPath, imagePath)
		if err != nil{
			log.Fatalf("Failed to extract tar file: %v", err)
		}
	} else if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		log.Fatalf("Image directory not found: %s", imagePath)
	} else if err != nil{
		log.Fatalf("Failed to check image directory: %v", err)
	}

	// 컨테이너 경로 확인
	if _, err := os.Stat(containerPath); os.IsNotExist(err) {
		if err := os.Mkdir(containerPath, 0755); err != nil && !os.IsExist(err) {
			log.Fatalf("Failed to create container directory: %v", err)
		}
	} else {
		fmt.Printf("Container %s already exists", containerName)
		StartContainer(containerPath)
		return
	}

	// ----- pivot_root

	// 루트 파일 시스템 구성 (Chroot) [pivot root 사용하기]
	// 마운트 네임 스페이스 격리 및 Cgroup 추가
	// cgroups 설정을 위한 로직을 추가합니다.
	// 프로세스 실행
	// 리소스 할당 및 관리
	// 컨테이너 실행 유지

	pivotRoot(imagePath, containerPath)

	// 마운트 네임 스페이스 격리 및 Cgroup 추가
	if err := setupCgroups(); err != nil {
		log.Fatalf("Failed to set up cgroups: %v", err)
	}

	// 컨테이너 내에서 프로세스 실행
	cmd := exec.Command("/bin/sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start the command: %v", err)
	}

	// 리소스 할당 및 관리
	go func() {
		cmd.Wait() // 프로세스가 종료될 때까지 대기
		fmt.Println("Process has exited.")
	}()

	fmt.Printf("Container %s created and started successfully!\n", containerName)
	select {} // 컨테이너를 계속 실행 상태로 유지

}

// 이미지 압축 해제 함수
func extractTar(tarFile, destDir string) error {
	file, err := os.Open(tarFile)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %v", err)
	}
	defer file.Close()

	tarReader := tar.NewReader(file)
	for {
		header, err := tarReader.Next()
		if err == io.EOF{
			break
		}
		if err != nil{
			return fmt.Errorf("failed to read tar file: %v", err)
		}

		// 경로 설정
		targetPath := filepath.Join(destDir, header.Name) 
		switch header.Typeflag {
		case tar.TypeDir:
			// 디렉토리 생성
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil{
				return fmt.Errorf("failed to create directory: %v", err)
			}
		case tar.TypeReg:
			// 파일 생성
			outFile, err := os.Create(targetPath)
			if err != nil{
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

func pivotRoot(newRoot, containerPath string){
	// pivot_root 작업을 위한 old_root 디렉토리 생성
	putOld := filepath.Join(containerPath, ".pivot_root_old")
	if err := os.MkdirAll(putOld, 0700); err != nil {
		log.Fatalf("Failed to create old root directory: %v", err)
	}

	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		log.Fatalf("Failed to mount new root: %v", err)
	}

	if err := syscall.PivotRoot(newRoot, putOld); err != nil {
		log.Fatalf("Failed to pivot root: %v", err)
	}

	// 새로운 루트로 이동
	if err := os.Chdir("/"); err != nil {
		log.Fatalf("Failed to change directory to new root: %v", err)
	}

	// 사용이 끝난 old_root 마운트 해제 및 삭제
	putOld = "/.pivot_root_old"
	if err := syscall.Unmount(putOld, syscall.MNT_DETACH); err != nil {
		log.Fatalf("Failed to unmount old root: %v", err)
	}
	if err := os.RemoveAll(putOld); err != nil {
		log.Fatalf("Failed to remove old root directory: %v", err)
	}
}

func setupCgroups() error {
	cgroups := "/sys/fs/cgroup/"
	pid := os.Getpid()

	// Memory limit
	if err := os.MkdirAll(filepath.Join(cgroups, "memory", "carte"), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cgroups, "memory", "carte", "memory.limit_in_bytes"), []byte("104857600"), 0700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cgroups, "memory", "carte", "cgroup.procs"), []byte(strconv.Itoa(pid)), 0700); err != nil {
		return err
	}

	// CPU limit
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


func StartContainer(containerPath string) {
	// 컨테이너에서 실행할 명령어
	cmd := exec.Command("/bin/sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	cmd.Stdout = os.Stdout // ------> 어떤 의미?
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start();
	err != nil {
		log.Fatalf("Failed to start the command in existing container: %v", err)
	}

	// 리소스 할당 및 관리
	go func() {
		cmd.Wait() // 프로세스 종료까지 대기
		fmt.Println("Process in existing container has exited")
	}()

	fmt.Println("Container started succesfully")
	select {}
}



func main() {
	// 테스트를 위해 "hello-world"라는 이미지를 "test-container" 이름으로 컨테이너 생성
	CtCreate("testimage", "testcontainer")
}

// Carte_Daemon 실행(서버, 컨테이너 생성 구현), Carte_Client 실행(이미지 전달)
// 시스템 호출, 네임 스페이스,, fork 부모 자식 프로세스 필요
