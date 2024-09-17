package container

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strconv"
    "syscall"
)

// stopContainer 함수: 주어진 컨테이너 이름으로 컨테이너를 종료
func StopContainer(name string) {
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

    // 컨테이너의 PID로 프로세스 종료
    pid := containerInfo.PID
    fmt.Printf("컨테이너 %s 종료 중 (PID: %d)...\n", containerInfo.Name, pid)

    // SIGTERM으로 프로세스 종료 시도
    if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
        fmt.Printf("SIGTERM 전송 실패: %v\n", err)
        return
    }

    // 프로세스가 정상적으로 종료될 때까지 대기
    if err := waitForProcessToExit(pid); err != nil {
        fmt.Println("SIGTERM으로 프로세스 종료에 실패, SIGKILL을 시도합니다...")
        if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
            fmt.Printf("SIGKILL 전송 실패: %v\n", err)
            return
        }
    }

    // 종료 후 상태 업데이트
    containerInfo.Status = "stopped"
    containerInfo.ExitCode = 0 // SIGTERM 성공 시, 정상 종료로 가정

    // 컨테이너 정보 업데이트 (상태 변경)
    containerData, err = json.MarshalIndent(containerInfo, "", "    ")
    if err != nil {
        fmt.Println("컨테이너 정보 업데이트 실패:", err)
        return
    }

    if err := ioutil.WriteFile(containerFile, containerData, 0644); err != nil {
        fmt.Println("컨테이너 정보 파일 저장 실패:", err)
    }

    fmt.Printf("컨테이너 %s가 종료되었습니다.\n", containerInfo.Name)
}

// 프로세스가 정상적으로 종료될 때까지 대기하는 함수
func waitForProcessToExit(pid int) error {
    for i := 0; i < 5; i++ { // 5초간 대기
        if !isProcessRunning(pid) {
            return nil
        }
        syscall.Sleep(1)
    }
    return fmt.Errorf("프로세스가 종료되지 않았습니다")
}

// PID로 프로세스가 실행 중인지 확인하는 함수
func isProcessRunning(pid int) bool {
    // SIGCONT 신호를 사용하여 프로세스 상태 확인
    err := syscall.Kill(pid, 0)
    return err == nil
}
