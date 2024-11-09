package cmd

import (
	"fmt"
	"os/exec"
	"os"
	"github.com/spf13/cobra"
)

// 컨테이너 제거 명령어 정의
var removeCmd = &cobra.Command{
	Use:   "remove [ContainerName]",
	Short: "Remove Container",
	Args:  cobra.ExactArgs(1), // 1개의 인자를 받도록 설정
	RunE: func(cmd *cobra.Command, args []string) error {
		containerName := args[0] // 첫 번째 인자가 컨테이너 이름
		return CtRemove(containerName) // CtRemove 함수 호출
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}

// 컨테이너 제거 함수
func CtRemove(containerName string) error {
	// 컨테이너 경로 설정
	containerPath := "/CarteTest/container/" + containerName

	// 컨테이너 경로 확인
	if _, err := os.Stat(containerPath); os.IsNotExist(err) {
		return fmt.Errorf("Container %s does not exist", containerName)
	}

	// rm -rf 명령어를 사용하여 컨테이너 삭제
	cmd := exec.Command("rm", "-rf", containerPath)
	output, err := cmd.CombinedOutput() // 명령 실행 후 출력 및 에러를 함께 캡처

	if err != nil {
		// 실패 시 명령어 출력과 에러를 모두 표시
		return fmt.Errorf("Failed to remove container: %s\nOutput: %s", err, string(output))
	}

	fmt.Printf("Container %s removed successfully\n", containerName)
	return nil
}
