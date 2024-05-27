package models

import (
	"os/exec"
	"syscall"
)

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

