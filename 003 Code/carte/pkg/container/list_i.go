package container

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
)

// listImage 함수: 빌드된 이미지 목록을 출력
func ListImage() {
    imagesDir := "/var/run/carte/images/"

    // 이미지 디렉토리가 존재하는지 확인
    if _, err := os.Stat(imagesDir); os.IsNotExist(err) {
        fmt.Println("사용 가능한 이미지가 없습니다.")
        return
    }

    // 이미지 디렉토리의 파일 목록 읽기
    files, err := ioutil.ReadDir(imagesDir)
    if err != nil {
        fmt.Printf("이미지 목록을 읽는 중 오류 발생: %v\n", err)
        return
    }

    if len(files) == 0 {
        fmt.Println("사용 가능한 이미지가 없습니다.")
        return
    }

    fmt.Println("사용 가능한 이미지 목록:")
    for _, file := range files {
        if file.IsDir() {
            // 이미지 디렉토리 이름(이미지 ID) 출력
            fmt.Printf("- %s\n", file.Name())
        }
    }
}
