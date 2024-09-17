package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

// 이미지 목록을 확인하는 커맨드
var containerListCmd = &cobra.Command{
    Use:   "container",
    Short: "List containers",
    RunE: func(cmd *cobra.Command, args []string) error {
        return listContainers("/CarteTest/container")
    },
}

func init() {
    rootCmd.AddCommand(containerListCmd)
}

// 디렉토리 내 파일 목록을 읽어서 출력하는 함수
func listContainers(containerDir string) error {
    files, err := os.ReadDir(containerDir)
    if err != nil {
        return fmt.Errorf("디렉토리 읽기 실패: %v", err)
    }

    fmt.Println("---- container list ----")
    for _, file := range files {
        if file.IsDir() { // 파일만 출력, 하위 디렉토리 제외
            fmt.Println(file.Name())
        }
    }

    return nil
}
