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
)

func CreateContainer(c *gin.Context) {

	// cgroups를 사용하여 메모리 제한 설정
	cgroups()

	// 격리된 환경에서 새로운 컨테이너 실행
	cmd := exec.Command("/bin/sh")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET,
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Container created successfully"})
}

func Cgroups() {
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
