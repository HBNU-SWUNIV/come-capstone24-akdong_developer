package cmd

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
    Use:   "stop [containerName]",
    Short: "Stop a running container",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        containerName := args[0]
        
        fmt.Println("     Container Stop    ")
        fmt.Println("=========================")
        
        return stopContainer(containerName)
    },
}

func init() {
    rootCmd.AddCommand(stopCmd)
}

func stopContainer(containerName string) error {
    containerPath := "/CarteDaemon/container/" + containerName
    pidFilePath := filepath.Join(containerPath, "pid")

    // PID 파일에서 컨테이너의 PID 읽기
    pid, err := os.ReadFile(pidFilePath)
    if err != nil {
        return fmt.Errorf("failed to read PID file for container %s: %v", containerName, err)
    }
    pidStr := strings.TrimSpace(string(pid))
    fmt.Printf("Stopping container %s with PID %s\n", containerName, pidStr)

    // 마운트 해제
    if err := unmountContainer(containerPath); err != nil {
        fmt.Printf("Warning: failed to unmount container %s: %v\n", containerName, err)
    } else {
        fmt.Printf("Unmounted all mounts for container %s\n", containerName)
    }

    // veth 인터페이스 삭제
    vethHost := fmt.Sprintf("vh_%s", pidStr)
    if err := deleteVethInterface(vethHost); err != nil {
        fmt.Printf("Warning: failed to delete veth interface %s: %v\n", vethHost, err)
    } else {
        fmt.Printf("Deleted veth interface %s\n", vethHost)
    }

    // PID 파일 삭제
    if err := os.Remove(pidFilePath); err != nil {
        fmt.Printf("Warning: failed to remove PID file for container %s: %v\n", containerName, err)
    } else {
        fmt.Printf("Removed PID file for container %s\n", containerName)
    }

    return nil
}

// veth 인터페이스 삭제 함수
func deleteVethInterface(vethHost string) error {
    if err := exec.Command("ip", "link", "del", vethHost).Run(); err != nil {
        return fmt.Errorf("failed to delete veth interface %s: %v", vethHost, err)
    }
    return nil
}

// 마운트 해제 함수
func unmountContainer(containerPath string) error {
    // 컨테이너의 루트 및 시스템 디렉토리 포함
    mountPoints := []string{
        filepath.Join(containerPath, "proc"),
        filepath.Join(containerPath, "sys"),
        filepath.Join(containerPath, "dev"),
        containerPath, // 루트 컨테이너 경로
    }

    for _, mountPoint := range mountPoints {
        if err := exec.Command("umount", "-l", mountPoint).Run(); err != nil {
            fmt.Printf("Warning: failed to unmount %s: %v\n", mountPoint, err)
            continue // 오류가 발생해도 다음 마운트 포인트로 넘어감
        }
        fmt.Printf("Unmounted %s\n", mountPoint)
    }

    return nil
}
