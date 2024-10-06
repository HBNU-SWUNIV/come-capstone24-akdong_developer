package container

import (
    "fmt"
    "io/ioutil"
)

// 이미지 목록을 출력하는 함수
func ListImage() {
    imagesDir := "/var/run/carte/images/"
    files, err := ioutil.ReadDir(imagesDir)
    if err != nil {
        fmt.Printf("이미지 목록을 불러올 수 없습니다: %v\n", err)
        return
    }

    fmt.Println("사용 가능한 이미지 목록:")
    for _, file := range files {
        fmt.Println(file.Name())
    }
}
