package cmd

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// 실행 중인 컨테이너 목록 확인
var psCmd = &cobra.Command{
	Use:   "ps",
	Short: "List running containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listRunningContainers()
	},
}

func init() {
	rootCmd.AddCommand(psCmd)
}

func listRunningContainers() error {
	// 'pgrep -a busybox'를 사용하여 실행 중인 모든 busybox 프로세스를 찾음
	cmd := exec.Command("pgrep", "-a", "busybox")
	output, err := cmd.CombinedOutput() // CombinedOutput 사용
	if err != nil {
		return fmt.Errorf("failed to list running containers: %v\nDetails: %s", err, output)
	}

	// 출력 결과를 처리하여 실행 중인 컨테이너 목록을 출력
	procs := strings.Split(string(output), "\n")
	if len(procs) == 0 || (len(procs) == 1 && procs[0] == "") {
		fmt.Println("No running containers found.")
		return nil
	}

	fmt.Println("Running containers:")
	for _, proc := range procs {
		if proc != "" {
			fmt.Println(proc) // PID 및 명령어 출력
		}
	}

	return nil
}
