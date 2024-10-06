package container

import (
    "fmt"
    "io/ioutil"
)

// 컨테이너 목록을 출력하는 함수
func ListContainer() {
    containersDir := "/var/run/carte/containers/"
    files, err := ioutil.ReadDir(containersDir)
    if err != nil {
        fmt.Printf("컨테이너 목록을 불러올 수 없습니다: %v\n", err)
        return
    }

    fmt.Println("실행 중인 컨테이너 목록:")
    for _, file := range files {
        fmt.Println(file.Name())
    }
}


