package cmd

import (
    "fmt"
    "os"
    "os/exec"
    "github.com/spf13/cobra"
)

// image_create 명령 정의
var imageCreateCmd = &cobra.Command{
    Use:   "image_create [scriptFile]",
    Short: "Create an image using the specified script file",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        scriptFile := args[0]

        // 스크립트 파일 실행
        if err := executeScript(scriptFile); err != nil {
            return fmt.Errorf("failed to execute script: %v", err)
        }

        fmt.Printf("Image created successfully using script %s\n", scriptFile)
        return nil
    },
}

func init() {
    rootCmd.AddCommand(imageCreateCmd)
}

// 스크립트 실행 함수
func executeScript(scriptFile string) error {
    // 스크립트 파일 존재 여부 확인
    if _, err := os.Stat(scriptFile); os.IsNotExist(err) {
        return fmt.Errorf("script file %s does not exist", scriptFile)
    }

    // 실행 권한이 있는지 확인
    if err := os.Chmod(scriptFile, 0755); err != nil {
        return fmt.Errorf("failed to set execute permissions on script file: %v", err)
    }

    // 스크립트 실행
    cmd := exec.Command("/bin/bash", scriptFile)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 스크립트 실행 오류 처리
    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error running script: %v", err)
    }

    return nil
}
