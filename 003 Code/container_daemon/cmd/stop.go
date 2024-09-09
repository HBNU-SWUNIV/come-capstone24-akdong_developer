package cmd

import (
	"fmt"
	"os/exec"
	"github.com/spf13/cobra"
)

// 컨테이너 정지 함수 정의
func stopContainer(containerName string) error {
    // 임시로 컨테이너의 프로세스 ID를 찾는 로직을 추가합니다.
    // 실제 구현에서는 컨테이너 관리 시스템에 따라 PID를 얻는 방식을 사용해야 합니다.
    cmd := exec.Command("pgrep", "-f", containerName)
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to find container process: %v", err)
    }

    pid := string(output)
    killCmd := exec.Command("kill", "-9", pid)
    if err := killCmd.Run(); err != nil {
        return fmt.Errorf("failed to kill container process: %v", err)
    }

    fmt.Println("Container process stopped successfully.")
    return nil
}

// 실행 중인 컨테이너 정지 커맨드 정의
var stopCmd = &cobra.Command{
    Use:   "stop [container name]",
    Short: "Stop a running container",
    Args:  cobra.ExactArgs(1), // 정확히 하나의 인자를 요구합니다.
    RunE: func(cmd *cobra.Command, args []string) error {
        containerName := args[0]
        return stopContainer(containerName)
    },
}

func init() {
    rootCmd.AddCommand(stopCmd)
}
