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
    "github.com/google/uuid"
    "carte/pkg/subsystem"
    "carte/pkg/utils"
    "golang.org/x/sys/unix"
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

// nextAvailableIP는 새로운 컨테이너에 할당할 IP 주소를 관리합니다.
var nextAvailableIP = 100 // 초기 IP 설정

// getNextAvailableIP 함수는 다음 사용 가능한 IP 주소를 반환합니다.
func getNextAvailableIP() string {
    ipAddr := fmt.Sprintf("192.168.1.%d", nextAvailableIP)
    nextAvailableIP++
    if nextAvailableIP > 254 { // 서브넷 제한
        nextAvailableIP = 100 // IP 주소를 재설정
    }
    return ipAddr
}

// RunContainer 함수는 컨테이너를 실행합니다.
func RunContainer(name, imageID, cpuLimit, memoryLimit string) {
    imagesDir := "/var/run/carte/images/"
    imagePath := filepath.Join(imagesDir, imageID)

    if _, err := os.Stat(imagePath); os.IsNotExist(err) {
        fmt.Printf("이미지 %s를 찾을 수 없습니다.\n", imageID)
        return
    }

    // 브리지 설정 자동화
    if err := subsystem.SetupBridge(); err != nil {
        fmt.Println("브리지 설정 실패:", err)
        return
    }

    // 컨테이너 실행 설정
    cmd := exec.Command("/bin/sh") // 기본으로 /bin/sh를 실행 (추후 필요시 변경 가능)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
        AmbientCaps: []uintptr{unix.CAP_NET_ADMIN},
    }

    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Start(); err != nil {
        fmt.Println("컨테이너 실행 실패:", err)
        return
    }

    fmt.Printf("컨테이너가 PID %d에서 실행 중입니다.\n", cmd.Process.Pid)

    containerID := fmt.Sprintf("%d", cmd.Process.Pid)
    vethHost := "veth_" + containerID
    vethContainer := "eth0_" + containerID

    // 새로운 IP 주소 할당
    ipAddr := getNextAvailableIP()
    fmt.Printf("컨테이너에 할당된 IP: %s\n", ipAddr)

    // veth 페어 설정
    if err := subsystem.SetupVethPair(containerID, vethHost, vethContainer); err != nil {
        fmt.Println("veth 설정 실패:", err)
        return
    }

    // 네임스페이스에서 veth 활성화 및 IP 할당
    if err := subsystem.ActivateVethInContainer(containerID, vethContainer, ipAddr); err != nil {
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

    saveContainerInfo(cmd.Process.Pid, name, imageID, networkInfo, resourcesInfo)

    if err := cmd.Wait(); err != nil {
        fmt.Println("컨테이너 종료 실패:", err)
    } else {
        fmt.Println("컨테이너 실행 성공")
    }
}

// 컨테이너 정보를 저장하는 함수
func saveContainerInfo(pid int, name, image string, networkInfo, resourcesInfo map[string]string) {
    containersDir := "/var/run/carte/containers/"
    utils.CreateDirIfNotExists(containersDir) // 중복된 함수 대신 utils 사용

    if name == "" {
        name = fmt.Sprintf("container-%d", pid)
    }

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




