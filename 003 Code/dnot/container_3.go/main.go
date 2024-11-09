package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Mount가 진행되지 않아서 다시 개념 학습겸 작성
// 현재 이미지는 우선 예시파일로 하나씩 진행해보기(현재 진행하고 있는 login 서버)

// 목표 : docker build 명령어처럼 carte build image 진행하면 이미지 생성, carte container run 진행시 컨테이너 생성


// 1) 명령어로 이미지 생성하기[carte build image]
// 1_1) 이미지 압축하기 
// 		인터넷이 연결되지 않은 경우나, 내부망 사용으로 외부 인터넷에 접속할 수 없는 경우 자신이 원하는 이미지를 다운로드할 수 있음
//		+ DockerHub와의 연결 없이도 이미지를 다운받을 수 있음

func buildImage() error {
	
	fmt.Println("----------------------- Start Build -----------------------")
	err := os.MkdirAll("/Carte/rootfs", 0755)
	if err != nil{
		return err
	}

	// bin/bash, bin/ls 파일로 입력 진행 --> 추후에 login server 파일 넣는거(경로 바꿔서) 해보기, golang 연관성 및 의존성 생각해봐야됨
	var imagepath string

	ct_build_iput := 1
	for {
		if ct_build_iput == 0 {
			break
		}
		fmt.Print("input your filepath: ")
		_, err := fmt.Scanf("%s", &imagepath)
		if err != nil{
			return err
		}

		// only file copy --> 
		// Carte(run), need exe? , no set main.go , auto input carte ,
		err = copyFile(imagepath, "/Carte/rootfs/")
		if err != nil{
			return err
		}

		fmt.Print("(if you continue please input 1, else input 0), input : ")
		_, err = fmt.Scanf("%d", &ct_build_iput)
		if err != nil {
			return err
		}
	}
	
	fmt.Println("----------------------- Start image create -----------------------")
	err = createImage("/Carte/rootfs", "/Carte/image.tar")
	if err != nil {
		return err
	}

	fmt.Println("---------------------- Image Build Complete ----------------------")
	return nil
}

// 파일 복사 함수
func copyFile(src, dst string) error {
	cmd := exec.Command("cp -r -p", src, dst)
	return cmd.Run()
}

// 이미지 파일 생성 함수
func createImage(srcDir, dstFile string) error {
	cmd := exec.Command("tar", "-C", srcDir, "-cvf", dstFile, ".")
	return cmd.Run()
}

// 2) 명령어로 container 생성하기[carte container run]
func runContainer() error {
	fmt.Println("----------------------- Start Create Container -----------------------")
	return nil
}

// 2_1) 실행 container 확인하기(carte cotainer ps)


// 3) 컨테이너 2개로 1:1 통신 진행해보기

func main() {

	var procedure string
	fmt.Scanf("%s", &procedure)

	if procedure == "Carte_Image_build" {
		err := buildImage()
		if err != nil {
			fmt.Println("Error building image:", err)
			return
		}
	}

	if procedure == "Carte_Create_Container" {
		err := runContainer()
		if err != nil {
			fmt.Println("Error create container:", err)
			return
		}
	}
	
}
