package cmd
// 컨테이너 격리가 제대로 되지 않는 문제!!!!!!

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
var startCmd = &cobra.Command{
    Use:   "start [containerName]",
    Short: "Container start",
    RunE: func(cmd *cobra.Command, args []string) error {
        containerName := args[0]
        containerPath := "/CarteTest/container/" + containerName

        // cgroups 설정
        if err := setupCgroups(containerPath); err != nil {
            log.Fatal("Failed to setup cgroups:", err)
        }

        // 파일 시스템 동기화 강화
        syscall.Sync()
        time.Sleep(2 * time.Second)

        // 컨테이너 실행
        fmt.Println("Attempting to start container...")
        if err := startContainer(containerPath, containerName); err != nil {
            return fmt.Errorf("error starting container: %v", err)
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(startCmd)
}

func runInNewNamespace(containerPath, path string, args []string, containerName string) (*exec.Cmd, error) {
    fullPath := filepath.Join(containerPath, path)
    fmt.Printf("Before chroot, checking path: %s\n", fullPath)
    if _, err := os.Stat(fullPath); err != nil {
        return nil, fmt.Errorf("before chroot: command not found: %v", err)
    }

    // Chroot 설정
    if err := syscall.Chroot(containerPath); err != nil {
        return nil, fmt.Errorf("failed to apply chroot to container path: %v", err)
    }
    fmt.Println("Chroot applied successfully")

    if err := os.Chdir("/"); err != nil {
        return nil, fmt.Errorf("failed to change to new root directory: %v", err)
    }

    // 명령 실행
    cmd := exec.Command(path, args...)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER,
    }
    
    // 사용자 UID/GID 매핑 설정
    cmd.SysProcAttr.UidMappings = []syscall.SysProcIDMap{
        {ContainerID: 0, HostID: os.Getuid(), Size: 1}, // 컨테이너의 root (0)와 호스트의 UID를 매핑
    }
    cmd.SysProcAttr.GidMappings = []syscall.SysProcIDMap{
        {ContainerID: 0, HostID: os.Getgid(), Size: 1}, // 컨테이너의 root (0)와 호스트의 GID를 매핑
    }

    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    fmt.Println("Starting command execution")
    if err := cmd.Start(); err != nil {
        return nil, fmt.Errorf("failed to run command in new namespace: %v, cmd: %+v", err, cmd)
    }
    fmt.Println("Command execution started")

    return cmd, nil
}



func startContainer(containerPath, containerName string) error {
    //cmd, err := runInNewNamespace(containerPath, "/bin/busybox", []string{"sh"}, containerName)
    cmd, err := runInNewNamespace(containerPath, "/bin/busybox", []string{"ip", "a"}, containerName)

    
    if err != nil {
        return fmt.Errorf("failed to start container in new namespace: %v", err)
    }

    // PID 기록
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
    pidFilePath := "pid"
    currentDir, err := os.Getwd()
    if err != nil {
        fmt.Printf("Error getting current directory: %v\n", err)
        return err
    }
    fmt.Printf("Current working directory: %s\n", currentDir)

    if _, err := os.Stat(pidFilePath); os.IsNotExist(err) {
        fmt.Printf("PID file does not exist, creating new one for container %s\n", containerName)

        pidFile, err := os.Create(pidFilePath)
        if err != nil {
            return fmt.Errorf("failed to create PID file: %v", err)
        }
        defer pidFile.Close()

        _, err = pidFile.WriteString(fmt.Sprintf("%d", pid))
        if err != nil {
            return fmt.Errorf("failed to write PID to file: %v", err)
        }

        fmt.Printf("PID %d recorded for container %s\n", pid, containerName)
    } else if err != nil {
        return fmt.Errorf("error checking PID file: %v", err)
    } else {
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

    cpuLimitPath := filepath.Join(cgroupRoot, "cpu", "myContainerGroup")
    if err := os.MkdirAll(cpuLimitPath, 0755); err != nil {
        return fmt.Errorf("failed to create cgroup for cpu: %v", err)
    }

    if err := os.WriteFile(filepath.Join(cpuLimitPath, "cpu.cfs_quota_us"), []byte("100000"), 0644); err != nil {
        return fmt.Errorf("failed to set cpu quota: %v", err)
    }

    if err := os.WriteFile(filepath.Join(cpuLimitPath, "tasks"), []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
        return fmt.Errorf("failed to add process to cgroup: %v", err)
    }

    return nil
}
