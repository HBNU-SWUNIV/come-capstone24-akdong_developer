package cmd

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "syscall" 
    "time"
	
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Container start",
	RunE: func(cmd *cobra.Command, args []string) error {  // err 타입을 error로 변경
		containerName := "testcontainer"
		containerPath := "/CarteTest/container/" + containerName

		// cgroups 설정
		if err := setupCgroups(containerPath); err != nil {
			log.Fatal("Failed to setup cgroups:", err)
		}

		// 파일 시스템 동기화 강화
		syscall.Sync()  // 동기화 호출
		time.Sleep(2 * time.Second)  // 동기화가 시스템에 반영될 시간을 기다림
	
		// 컨테이너 실행
		fmt.Println("Attempting to start container...")

		// 컨테이너 시작 (명령 실행)
		if err := startContainer(containerPath, containerName); err != nil {
			return fmt.Errorf("error starting container: %v", err)
		}

		return nil  // 함수가 정상적으로 끝나면 nil 반환
	},
}

func init(){
    rootCmd.AddCommand(startCmd)
}

// 새로운 네임스페이스에서 명령어 실행
func runInNewNamespace(containerPath, path string, args []string, containerName string) error {
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
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
        // CLONE_NEWNET 플래그를 제거했습니다.
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

func startContainer(containerPath, containerName string) error {
	// /CarteTest/container/testcontainer/www(생성)/index.html(생성)
	return runInNewNamespace(containerPath, "/bin/busybox", []string{"httpd", "-f", "-p", "8080", "-h", "/www"}, containerName)
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
