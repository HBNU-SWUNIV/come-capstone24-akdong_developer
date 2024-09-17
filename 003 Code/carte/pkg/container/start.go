package container

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"
    "time"
    "github.com/google/uuid"
    "carte/pkg/subsystem"
)

// startContainer 함수: 주어진 컨테이너 이름으로 컨테이너를 다시 시작
func StartContainer(name string) {
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

    // 컨테이너 상태가 'stopped'인지 확인
    if containerInfo.Status != "stopped" {
        fmt.Printf("컨테이너 %s는 이미 실행 중이거나 정지 상태가 아닙니다.\n", name)
        return
    }

    // 새로운 프로세스로 컨테이너 실행
    cmd := exec.Command(containerInfo.Image) // 이미지에 맞는 프로세스를 재시작
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
    }

    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        fmt.Println("컨테이너 재시작 실패:", err)
        return
    }

    // 네트워크 정보 동적 할당
    containerID := fmt.Sprintf("%d", cmd.Process.Pid)
    ipAddr := containerInfo.Network["ip_address"] // 기존 IP 주소 유지

    if err := subsystem.SetupNetworkNamespace(containerID, ipAddr); err != nil {
        fmt.Println("네트워크 설정 실패:", err)
        return
    }

    // cgroups 자원 제한 설정
    if err := subsystem.SetCgroupLimits(containerID, containerInfo.Resources["cpu"], containerInfo.Resources["memory"]); err != nil {
        fmt.Println("cgroups 설정 실패:", err)
        return
    }

    // 컨테이너 상태 및 PID 업데이트
    containerInfo.PID = cmd.Process.Pid
    containerInfo.Status = "running"
    containerInfo.CreatedAt = time.Now().Format(time.RFC3339)

    // 컨테이너 정보 파일 업데이트
    containerData, err = json.MarshalIndent(containerInfo, "", "    ")
    if err != nil {
        fmt.Println("컨테이너 정보 업데이트 실패:", err)
        return
    }

    if err := ioutil.WriteFile(containerFile, containerData, 0644); err != nil {
        fmt.Println("컨테이너 정보 파일 저장 실패:", err)
    }

    fmt.Printf("컨테이너 %s가 다시 시작되었습니다.\n", containerInfo.Name)

    if err := cmd.Wait(); err != nil {
        fmt.Println("컨테이너 실행 중 오류 발생:", err)
    }
}
