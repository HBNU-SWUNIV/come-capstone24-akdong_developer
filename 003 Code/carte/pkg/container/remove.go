package container

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
)

// removeContainer 함수: 주어진 컨테이너 이름으로 컨테이너를 삭제
func RemoveContainer(name string) {
    containersDir := "/var/run/carte/containers/"

    // 해당 컨테이너의 정보를 읽기 위해 파일 확인
    containerFile := filepath.Join(containersDir, name + ".json")
    if _, err := os.Stat(containerFile); os.IsNotExist(err) {
        fmt.Printf("컨테이너 %s를 찾을 수 없습니다.\n", name)
        return
    }

    // 컨테이너 정보 파일 읽기
    containerData, err := ioutil.ReadFile(containerFile)
    if err != nil {
        log.Fatalf("컨테이너 정보 파일 읽기 실패: %v", err)
    }

    var containerInfo ContainerInfo
    if err := json.Unmarshal(containerData, &containerInfo); err != nil {
        log.Fatalf("JSON 파싱 실패: %v", err)
    }

    // 컨테이너 상태 확인 (실행 중인 경우 삭제 불가)
    if containerInfo.Status == "running" {
        fmt.Printf("실행 중인 컨테이너 %s는 삭제할 수 없습니다. 먼저 정지시켜야 합니다.\n", name)
        return
    }

    // 컨테이너 정보 파일 삭제
    if err := os.Remove(containerFile); err != nil {
        fmt.Printf("컨테이너 정보 파일 삭제 실패: %v\n", err)
        return
    }

    fmt.Printf("컨테이너 %s가 삭제되었습니다.\n", name)
}
