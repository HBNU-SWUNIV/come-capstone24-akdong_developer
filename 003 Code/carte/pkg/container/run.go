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

func RunContainer(name, imageID, cpuLimit, memoryLimit string) {
    imagesDir := "/var/run/carte/images/"
    imagePath := filepath.Join(imagesDir, imageID)

    if _, err := os.Stat(imagePath); os.IsNotExist(err) {
        fmt.Printf("이미지 %s를 찾을 수 없습니다.\n", imageID)
        return
    }

    cmd := exec.Command("/bin/sh")
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

    // SetupCNINetwork 호출에서 IP 주소와 에러를 처리합니다.
    ipAddr, err := subsystem.SetupCNINetwork(containerID)
    if err != nil {
        fmt.Println("CNI 네트워크 설정 실패:", err)
        return
    }
    fmt.Printf("컨테이너에 할당된 IP: %s\n", ipAddr)

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
        "ip_address": ipAddr, // 실제 할당된 IP 주소를 저장합니다.
    }

    saveContainerInfo(cmd.Process.Pid, name, imageID, networkInfo, resourcesInfo)

    if err := cmd.Wait(); err != nil {
        fmt.Println("컨테이너 종료 실패:", err)
    } else {
        fmt.Println("컨테이너 실행 성공")
    }
}


func saveContainerInfo(pid int, name, image string, networkInfo, resourcesInfo map[string]string) {
    containersDir := "/var/run/carte/containers/"
    utils.CreateDirIfNotExists(containersDir)

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
