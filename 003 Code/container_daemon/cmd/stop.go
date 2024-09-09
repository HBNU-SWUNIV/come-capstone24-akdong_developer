package cmd

import (
    "fmt"
    "os/exec"
    "strings"

    "github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
    Use:   "stop [containerName]",
    Short: "Stop a running container",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        containerName := args[0]
        return stopContainer(containerName)
    },
}

func init() {
    rootCmd.AddCommand(stopCmd)
}

func stopContainer(containerName string) error {
    // 컨테이너와 관련된 PID 찾기 (httpd 프로세스를 찾음)
    pid, err := getContainerPID("/bin/busybox")
    if err != nil {
        return fmt.Errorf("failed to get PID for container %s: %v", containerName, err)
    }

    // PID로 프로세스 종료
    if err := exec.Command("kill", "-9", pid).Run(); err != nil {
        return fmt.Errorf("failed to stop container %s: %v", containerName, err)
    }

    fmt.Printf("Container %s stopped successfully\n", containerName)
    return nil
}

// /bin/busybox와 연관된 PID 찾기
func getContainerPID(processName string) (string, error) {
    // 'pgrep' 명령을 사용하여 프로세스 이름으로 PID 찾기
    cmd := exec.Command("pgrep", "-f", processName)
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to find process: %v", err)
    }

    // 출력된 PID 리스트 중 첫 번째 PID 반환
    pid := strings.TrimSpace(string(output))
    if pid == "" {
        return "", fmt.Errorf("no process found for %s", processName)
    }

    return pid, nil
}
