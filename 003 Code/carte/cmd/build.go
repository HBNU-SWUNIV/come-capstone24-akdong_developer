package cmd

import (
	"carte/models"
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var imageName string

// cobra 라이브러리, Use : 명령어 정의, Short : 명령어에 대한 간단한 설명, Run : 명령어가 실행될 때 호출되는 함수.
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a container image",
	Run: func(cmd *cobra.Command, args []string) {
		// 현재 디렉토리에 Cartefile이 있는지 확인
		cartefilePath := "Cartefile"
		if _, err := os.Stat(cartefilePath); os.IsNotExist(err) {
			fmt.Println("Cartefile not found in the current directory")
			return
		}

		//현재 작업 디렉토리 가져오기
		workingDir, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error getting current directory: %s\n", err)
			return
		}

		// 기본 이미지 이름 설정
		if imageName == "" {
			imageName = "image.tar.gz" 
		}

		// 설정한 이미지 이름 
		fmt.Printf("Building container image with name: %s...\n", imageName)

		err = models.BuildImage(imageName, workingDir)
		if err != nil {
			fmt.Printf("Error building image: %s\n", err)
			return
		}

		fmt.Println("Image built successfully.")
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVarP(&imageName, "name", "n", "", "Name of the output image file (default is 'image.tar.gz')")
}
