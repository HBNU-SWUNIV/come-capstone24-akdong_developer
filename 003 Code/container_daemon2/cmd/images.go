package cmd

import (
    "fmt"
    "os"

    "github.com/spf13/cobra"
)

// 이미지 목록을 확인하는 커맨드
var imageListCmd = &cobra.Command{
    Use:   "list_i",
    Short: "List images",
    RunE: func(cmd *cobra.Command, args []string) error {
        return listImages("/CarteTest/image")
    },
}

func init() {
    rootCmd.AddCommand(imageListCmd)
}

// 디렉토리 내 파일 목록을 읽어서 출력하는 함수
func listImages(imageDir string) error {
    files, err := os.ReadDir(imageDir)
    if err != nil {
        return fmt.Errorf("디렉토리 읽기 실패: %v", err)
    }

    fmt.Println("---- image list ----")
    for _, file := range files {
        if !file.IsDir() { // 파일만 출력, 하위 디렉토리 제외
            fmt.Println(file.Name())
        }
    }

    return nil
}
