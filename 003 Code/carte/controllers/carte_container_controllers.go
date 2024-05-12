package controllers

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
	// "github.com/containerd/cgroups"
)

func CreateContainer(c *gin.Context) {
	// cgroups를 사용하여 메모리 제한 설정
	cgroups()

	// 각 네임스페이스 설정 함수 호출
    setupUTSNamespace()
    setupPIDNamespace()
    setupNetworkNamespace()
    setupIPCNamespace()
    setupMountNamespace()

	// 컨테이너 내부에서 호스트 이름, PID, IP 주소, IPC 등을 확인하기 위해 명령어 실행
    cmd := exec.Command("/bin/sh", "-c", "hostname; ps aux; ip a; ipcs")
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Container created successfully"})
}

func cgroups() {
	// cgroups 경로
	cgroup := "/sys/fs/cgroup/"
	pid := os.Getpid()
	memLimit := "100000000" // 예: 100MB

	// 메모리 cgroup 설정
	memCgroupPath := filepath.Join(cgroup, "memory", "mycontainer")
	os.Mkdir(memCgroupPath, 0755)
	os.WriteFile(filepath.Join(memCgroupPath, "memory.limit_in_bytes"), []byte(memLimit), 0644)
	os.WriteFile(filepath.Join(memCgroupPath, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644)

	// 여기에 컨테이너 실행 로직 추가
	fmt.Println("Container with limited memory running...")
}

func setupUTSNamespace() {
    // 호스트 이름 변경
    cmd := exec.Command("/bin/hostname", "container1")
    cmd.Run()
}

func setupPIDNamespace() {
    // 프로세스 ID 변경
    syscall.Sethostname([]byte("container1"))

    // PID 변경
    cmd := exec.Command("/bin/sh", "-c", "echo 1 > /proc/self/ns/pid")
    cmd.Run()
}

func setupNetworkNamespace() {
    // 네트워크 설정 (생략)
    cmd := exec.Command("/bin/sh", "-c", "ip link add veth0 type veth peer name veth1")
    cmd.Run()
}

func setupIPCNamespace() {
    // IPC 설정 (생략)
    cmd := exec.Command("/bin/sh", "-c", "ipcmk -M 1024")
    cmd.Run()
}

func setupMountNamespace() {
    // 파일 시스템 설정 (생략)
    cmd := exec.Command("/bin/sh", "-c", "mkdir /mnt/containerroot; mount -t tmpfs none /mnt/containerroot")
    cmd.Run()
}













// func BuildImage() error {
//     // 이미지 디렉토리 생성
// 	fmt.Println("Start Mkdir")
// 	err := os.MkdirAll("/my_container/rootfs", 0755)
// 	if err != nil {
// 		return err
// 	}

// 	// 필요한 파일 복사
// 	fmt.Println("Start CP")
// 	err = copyFile("/bin/bash", "/my_container/rootfs/")
// 	if err != nil {
// 		return err
// 	}
// 	err = copyFile("/bin/ls", "/my_container/rootfs/")
// 	if err != nil {
// 		return err
// 	}

// 	// 이미지 파일 생성
// 	fmt.Println("Start image create")
// 	err = createImage("/my_container/rootfs", "/my_container/image.tar")
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println("Image build complete.")
// 	return nil

//     // 압축 -- 대기 시간이 너무 오래걸림 ++ 파이프 설정 필요
//     // tarCmd := exec.Command("tar", "-C", "/my_container/rootfs", "-c", ".")
//     // gzipCmd := exec.Command("gzip")
// }

// // 파일 복사 함수
// func copyFile(src, dst string) error {
// 	cmd := exec.Command("cp", src, dst)
// 	return cmd.Run()
// }

// // 이미지 파일 생성 함수
// func createImage(srcDir, dstFile string) error {
// 	cmd := exec.Command("tar", "-C", srcDir, "-cvf", dstFile, ".")
// 	return cmd.Run()
// }



