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

// 컨테이너 실행 커맨드
// Carte start [컨테이너 명]
var startCmd = &cobra.Command{
	Use:   "start [containerName]",
	Short: "Container start",
	RunE: func(cmd *cobra.Command, args []string) error { 
		containerName := args[0]
        //containerName := "testcontainer"
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
func runInNewNamespace(containerPath, path string, args []string, containerName string) (*exec.Cmd, error) {
    // chroot 전 경로 확인
    fullPath := filepath.Join(containerPath, path)
    fmt.Printf("Before chroot, checking path: %s\n", fullPath)
    if _, err := os.Stat(fullPath); err != nil {
        return nil, fmt.Errorf("before chroot: command not found: %v", err)
    }

    // chroot 적용
    if err := syscall.Chroot(containerPath); err != nil {
        return nil, fmt.Errorf("failed to apply chroot to container path: %v", err)
    }
    fmt.Println("Chroot applied successfully")


    // chroot 적용 후에는 더 이상 fullPath를 사용할 수 없음
    if _, err := os.Stat(path); err != nil {
        return nil, fmt.Errorf("after chroot: command not found: %v", err)
    }

    // 루트 디렉토리로 이동
    if err := os.Chdir("/"); err != nil {
        return nil, fmt.Errorf("failed to change to new root directory: %v", err)
    }

    // 명령 실행
    cmd := exec.Command(path, args...)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        
        // Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS,
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET, //syscall.CLONE_NEWUSER,
        // UidMappings: []syscall.SysProcIDMap{
        //     {ContainerID: 0, HostID: os.Getuid(), Size: 1},
        // },
        // GidMappings: []syscall.SysProcIDMap{
        //     {ContainerID: 0, HostID: os.Getgid(), Size: 1},
        // },
        // Credential: &syscall.Credential{
        //     Uid: uint32(os.Getuid()),
        //     Gid: uint32(os.Getgid()),
        // },
    }
    
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    fmt.Println("Starting command execution")
    if err := cmd.Start(); err != nil {
        return nil, fmt.Errorf("failed to run command in new namespace: %v", err)
    }
    fmt.Println("Command execution started")

    return cmd, nil
}


//////////////////////////////////////////////////////////////////
// 이 부분에 대한 스크립트 필요함(사용자가 입력할 수 있도록) --> 일일히 칠 수 있도록 하기
func startContainer(containerPath, containerName string) error {
    // /CarteTest/container/testcontainer/www(생성)/index.html(생성)

    cmd, err := runInNewNamespace(containerPath, "/bin/busybox", []string{"sh"}, containerName)
    // cmd, err := runInNewNamespace(containerPath, "/bin/busybox", []string{"httpd", "-f", "-p", "8080", "-h", "/www"}, containerName)
    if err != nil {
        return fmt.Errorf("failed to start container in new namespace: %v", err)
    }

    // PID 기록 (컨테이너 이름과 연결)
    pid := cmd.Process.Pid
    if err := recordContainerPID(containerName, pid); err != nil {
        return fmt.Errorf("failed to record PID: %v", err)
    }

    fmt.Printf("Container %s started with PID %d\n", containerName, pid)

    // 프로세스 실행 종료 대기
    if err := cmd.Wait(); err != nil {
        return fmt.Errorf("process finished with error: %v", err)
    }

    return nil
}

func recordContainerPID(containerName string, pid int) error {
    // 컨테이너 이름과 PID를 기록하는 파일 경로
    pidFilePath := "pid"
    //pidFilePath := fmt.Sprintf("/CarteTest/container/%s/pid", containerName)

    // 현재 작업 디렉토리 확인
    currentDir, err := os.Getwd()
    if err != nil {
        fmt.Printf("Error getting current directory: %v\n", err)
        return err
    }
    fmt.Printf("Current working directory: %s\n", currentDir)

    // PID 파일이 존재하는지 확인
    if _, err := os.Stat(pidFilePath); os.IsNotExist(err) {
        // PID 파일이 존재하지 않으면 새로 생성
        fmt.Printf("PID file does not exist, creating new one for container %s\n", containerName)

        // PID 파일 생성
        pidFile, err := os.Create(pidFilePath)
        if err != nil {
            return fmt.Errorf("failed to create PID file: %v", err)
        }
        defer pidFile.Close()

        // PID 값을 파일에 기록
        _, err = pidFile.WriteString(fmt.Sprintf("%d", pid))
        if err != nil {
            return fmt.Errorf("failed to write PID to file: %v", err)
        }

        fmt.Printf("PID %d recorded for container %s\n", pid, containerName)
    } else if err != nil {
        // 다른 오류 발생 시 에러 처리
        return fmt.Errorf("error checking PID file: %v", err)
    } else {
        // PID 파일이 존재하면 아무 작업도 하지 않음
        fmt.Printf("PID file already exists for container %s, skipping creation\n", containerName)
    }

    return nil
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
