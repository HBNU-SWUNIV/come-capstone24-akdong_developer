package container

import (
    "fmt"
    "io/ioutil"
    "path/filepath"
    "syscall"
    "encoding/json"
)

// 컨테이너 중지 함수
func StopContainer(name string) {
    containersDir := "/var/run/carte/containers/"
    containerFile := filepath.Join(containersDir, name+".json")

    // 컨테이너 정보 읽기
    containerData, err := ioutil.ReadFile(containerFile)
    if err != nil {
        fmt.Printf("컨테이너 정보 읽기 실패: %v\n", err)
        return
    }

    var containerInfo ContainerInfo
    if err := json.Unmarshal(containerData, &containerInfo); err != nil {
        fmt.Printf("컨테이너 정보 파싱 실패: %v\n", err)
        return
    }

    // PID를 기반으로 프로세스 종료
    if err := syscall.Kill(containerInfo.PID, syscall.SIGTERM); err != nil {
        fmt.Printf("컨테이너 중지 실패: %v\n", err)
        return
    }

    containerInfo.Status = "stopped"

    // 컨테이너 정보를 다시 저장
    updatedData, err := json.MarshalIndent(containerInfo, "", "    ")
    if err != nil {
        fmt.Printf("컨테이너 정보 업데이트 실패: %v\n", err)
        return
    }

    if err := ioutil.WriteFile(containerFile, updatedData, 0644); err != nil {
        fmt.Printf("컨테이너 정보 저장 실패: %v\n", err)
    }

    fmt.Printf("컨테이너 %s가 중지되었습니다.\n", name)
}

