package container

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "syscall"
    "time"
    "path/filepath"
    "strconv"
    "github.com/google/uuid"
    "carte/pkg/subsystem"
)

// 컨테이너 정보 구조체
type ContainerInfo struct {
    ID        string            `json:"id"`
    Name      string            `json:"name"`
    Image     string            `json:"image"`
    PID       int               `json:"pid"`
    Status    string            `json:"status"`
    Network   map[string]string `json:"network"`
    Resources map[string]string `json:"resources"`
    CreatedAt string            `json:"created_at"`
    ExitCode  int               `json:"exit_code,omitempty"`
}

// RunContainer 실행 시 파라미터로 이름과 이미지 지정
func RunContainer(name, image, cpuLimit, memoryLimit string) {
    // 이미지에 따른 실행 프로세스 설정
    cmd := exec.Command(image)  // 예시로 이미지 이름을 실행 파일로 처리
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
    }

    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        fmt.Println("컨테이너 실행 실패:", err)
        return
    }

    // 네트워크 정보 동적 할당
    containerID := fmt.Sprintf("%d", cmd.Process.Pid)
    ipAddr := "192.168.1." + containerID // 예시 IP 주소

    if err := subsystem.SetupNetworkNamespace(containerID, ipAddr); err != nil {
        fmt.Println("네트워크 설정 실패:", err)
        return
    }

    // cgroups 자원 제한 설정
    if err := subsystem.SetCgroupLimits(containerID, cpuLimit, memoryLimit); err != nil {
        fmt.Println("cgroups 설정 실패:", err)
        return
    }

    resourcesInfo := map[string]string{
        "cpu":    cpuLimit,
        "memory": memoryLimit,
    }

    networkInfo := map[string]string{
        "ip_address": ipAddr,
    }

    saveContainerInfo(cmd.Process.Pid, name, image, networkInfo, resourcesInfo)

    if err := cmd.Wait(); err != nil {
        fmt.Println("컨테이너 종료 실패:", err)
    } else {
        fmt.Println("컨테이너 실행 성공")
    }
}

// 컨테이너 정보를 저장하는 함수
func saveContainerInfo(pid int, name, image string, networkInfo, resourcesInfo map[string]string) {
    containersDir := "/var/run/carte/containers/"
    createDirIfNotExists(containersDir)

    // 이름이 없으면 기본 이름 생성
    if name == "" {
        name = fmt.Sprintf("container-%d", pid)
    }

    // 컨테이너 정보 생성
    containerInfo := ContainerInfo{
        ID:        uuid.New().String(),
        Name:      name,
        Image:     image,
        PID:       pid,
        Status:    "running",
        Network:   networkInfo,
        Resources: resourcesInfo,
        CreatedAt: time.Now().Format(time.RFC3339),
    }

    // 컨테이너 정보를 JSON으로 변환하여 파일로 저장
    containerData, err := json.MarshalIndent(containerInfo, "", "    ")
    if err != nil {
        fmt.Println("컨테이너 정보 저장 실패:", err)
        return
    }

    containerFile := containersDir + containerInfo.ID + ".json"
    if err := ioutil.WriteFile(containerFile, containerData, 0644); err != nil {
        fmt.Println("컨테이너 정보 파일 저장 실패:", err)
    }
}

// 중복되는 디렉토리 생성 로직을 함수로 분리
func createDirIfNotExists(dir string) {
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        os.MkdirAll(dir, 0755)
    }
}


