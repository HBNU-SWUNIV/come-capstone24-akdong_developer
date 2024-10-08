package cmd

import (
    "fmt"
    "os"
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

// 컨테이너 중지 함수 ///////////////////////////////////// bin/busybox말고 다른 이미지도 사용가능하도록 ///////
func stopContainer(containerName string) error {
    // 컨테이너와 관련된 PID 찾기
    pid, err := getContainerPID("/bin/busybox")
    if err != nil {
        return fmt.Errorf("failed to get PID for container %s: %v", containerName, err)
    }

    // PID로 프로세스 종료
    if err := exec.Command("kill", "-9", pid).Run(); err != nil {
        return fmt.Errorf("failed to stop container %s: %v", containerName, err)
    }

    fmt.Printf("Container %s stopped successfully\n", containerName)

    // PID 파일 삭제
    if err := removeContainerPID(containerName); err != nil {
        return fmt.Errorf("failed to remove PID file for container %s: %v", containerName, err)
    }

    fmt.Printf("Remove PID")

    // 마운트 해제 (umount)
    if err := unmountContainer(containerName); err != nil {
        return fmt.Errorf("failed to unmount container %s: %v", containerName, err)
    }

    fmt.Printf("Container %s unmounted successfully\n", containerName)
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

// 컨테이너 마운트 해제 함수
func unmountContainer(containerName string) error {
    // /proc, /sys, /dev 등을 umount 처리
    mountPoints := []string{"/CarteTest/container/" + containerName + "/proc", 
                            "/CarteTest/container/" + containerName + "/sys", 
                            "/CarteTest/container/" + containerName + "/dev"}

    for _, mountPoint := range mountPoints {
        if err := exec.Command("umount", mountPoint).Run(); err != nil {
            return fmt.Errorf("failed to unmount %s: %v", mountPoint, err)
        }
    }

    return nil
}

// 컨테이너의 PID 파일을 삭제하는 함수
func removeContainerPID(containerName string) error {
    pidFilePath := "/CarteTest/container/" + containerName + "/pid"
    
    // PID 파일이 존재하는지 확인 후 삭제
    if err := os.Remove(pidFilePath); err != nil {
        return fmt.Errorf("failed to remove PID file: %v", err)
    }

    return nil
}