package container

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
)

func ListContainer() {
    containersDir := "/var/run/carte/containers/"
    if _, err := os.Stat(containersDir); os.IsNotExist(err) {
        fmt.Println("실행 중인 컨테이너가 없습니다.")
        return
    }

    files, err := ioutil.ReadDir(containersDir)
    if err != nil {
        log.Fatalf("디렉토리 읽기 실패: %v", err)
    }

    if len(files) == 0 {
        fmt.Println("실행 중인 컨테이너가 없습니다.")
        return
    }

    fmt.Println("실행 중인 컨테이너 목록:")
    for _, file := range files {
        if file.IsDir() {
            continue
        }

        // JSON 파일 읽기
        containerData, err := ioutil.ReadFile(filepath.Join(containersDir, file.Name()))
        if err != nil {
            log.Fatalf("파일 읽기 실패: %v", err)
        }

        // JSON을 ContainerInfo 구조체로 변환
        var containerInfo ContainerInfo
        if err := json.Unmarshal(containerData, &containerInfo); err != nil {
            log.Fatalf("JSON 파싱 실패: %v", err)
        }

        fmt.Printf("ID: %s, 이름: %s, 이미지: %s, 상태: %s, PID: %d, 생성 시각: %s\n",
            containerInfo.ID, containerInfo.Name, containerInfo.Image, containerInfo.Status, containerInfo.PID, containerInfo.CreatedAt)
    }
}


