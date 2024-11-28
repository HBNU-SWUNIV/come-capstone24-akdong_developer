package cmd

import (
	"fmt"
	"os"
	"os/exec"
    "strings"

	"github.com/spf13/cobra"
)

var psListCmd = &cobra.Command{
	Use:   "list_c",
	Short: "List all containers with their running status",
	RunE: func(cmd *cobra.Command, args []string) error {
		return listRunningContainers("/CarteDaemon/container")
	},
}

func init() {
	rootCmd.AddCommand(psListCmd)
}

// 컨테이너 목록을 확인하고 실행 여부를 표시하는 함수
func listRunningContainers(containerDir string) error {
    // 컨테이너 디렉토리에서 컨테이너 목록 가져오기
    files, err := os.ReadDir(containerDir)
    if err != nil {
        return fmt.Errorf("failed to read container directory: %v", err)
    }

    // fmt.Println("---- Running Containers ----")
    fmt.Println("  Container List [running]")
    fmt.Println("===========================")

    for _, file := range files {
        if file.IsDir() {
            containerName := file.Name()
            isRunning, err := isContainerRunning(containerName)
            if err != nil {
                fmt.Printf("%s   Error: %v\n", containerName, err)
                continue
            }

            // 실행 중이면 T, 그렇지 않으면 F로 표시
            runningStatus := "F"
            if isRunning {
                runningStatus = "T"
            }

            fmt.Printf("[%s] %s\n", containerName, runningStatus)
        }
    }

    return nil
}


func isContainerRunning(containerName string) (bool, error) {
    // PID 파일 경로 설정
    pidFilePath := "/CarteDaemon/container/" + containerName + "/pid"
    // fmt.Printf("Checking container %s at %s\n", containerName, pidFilePath)

    // PID 파일이 있는지 확인
    if _, err := os.Stat(pidFilePath); os.IsNotExist(err) {
        // PID 파일이 존재하지 않으면 실행되지 않음으로 간주
        fmt.Printf("PID file does not exist for container %s\n", containerName)
        return false, nil
    } else if err != nil {
        // 다른 오류 발생 시 에러 반환
        return false, fmt.Errorf("error checking PID file: %v", err)
    }

    // PID 파일에서 PID 읽기
    pidData, err := os.ReadFile(pidFilePath)
    if err != nil {
        return false, fmt.Errorf("error reading PID file: %v", err)
    }

    // PID를 문자열에서 정수로 변환
    pid := strings.TrimSpace(string(pidData))
    // fmt.Printf("Found PID for container %s: %s\n", containerName, pid)

    // ps 명령을 통해 해당 PID가 실행 중인지 확인
    cmd := exec.Command("ps", "-p", pid)
    //output, err := cmd.CombinedOutput()
    if err := cmd.Run(); err != nil {
        // fmt.Printf("ps command failed for PID %s: %v\nOutput: %s\n", pid, err, string(output))
        return false, nil
    }

    // fmt.Printf("ps command succeeded for PID %s: Output: %s\n", pid, string(output))
    return true, nil
}
