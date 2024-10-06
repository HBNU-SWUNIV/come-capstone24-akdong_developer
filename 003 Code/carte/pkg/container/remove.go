package container

import (
    "fmt"
    "os"
    "path/filepath"
    "io/ioutil"
    "encoding/json"
)

// 컨테이너 제거 함수
func RemoveContainer(name string) {
    containersDir := "/var/run/carte/containers/"
    containerFile := filepath.Join(containersDir, name+".json")

    // 컨테이너 정보 로드
    containerInfo, err := loadContainerInfo(containerFile)
    if err != nil {
        fmt.Printf("컨테이너 정보 로드 실패: %v\n", err)
        return
    }

    // 컨테이너가 실행 중인지 확인
    if containerInfo.Status == "running" {
        fmt.Println("실행 중인 컨테이너를 제거할 수 없습니다. 먼저 중지하세요.")
        return
    }

    // 컨테이너 파일 삭제
    if err := os.Remove(containerFile); err != nil {
        fmt.Printf("컨테이너 파일 삭제 실패: %v\n", err)
        return
    }

    fmt.Printf("컨테이너 %s가 제거되었습니다.\n", name)
}

// 컨테이너 정보 로드 함수
func loadContainerInfo(filePath string) (ContainerInfo, error) {
    var containerInfo ContainerInfo
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return containerInfo, err
    }
    err = json.Unmarshal(data, &containerInfo)
    return containerInfo, err
}


