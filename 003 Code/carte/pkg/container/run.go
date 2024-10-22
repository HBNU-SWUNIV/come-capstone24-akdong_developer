package container

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"
    "time"
    "github.com/google/uuid"
    "carte/pkg/subsystem"
    "carte/pkg/utils"
    // "golang.org/x/sys/unix"
    
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

    // 쉘 명령 실행
    cmd := exec.Command("/bin/sh")
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
    }

    // 표준 입출력 연결
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 새로운 루트 내부에 put_old 디렉토리 생성
    putOld := filepath.Join(imagePath, "put_old")
    if err := os.MkdirAll(putOld, 0700); err != nil {
        fmt.Println("put_old 디렉토리 생성 실패:", err)
        return
    }

    // 새로운 루트를 바인드 마운트
    if err := syscall.Mount(imagePath, imagePath, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
        fmt.Println("파일 시스템 바인드 마운트 실패:", err)
        return
    }

    // pivot_root 실행
    if err := syscall.PivotRoot(imagePath, putOld); err != nil {
        fmt.Println("pivot_root 실패:", err)
        return
    }

    // 루트 디렉토리 변경
    if err := syscall.Chdir("/"); err != nil {
        fmt.Println("chdir 실패:", err)
        return
    }

    // put_old 마운트 해제
    if err := syscall.Unmount("/put_old", syscall.MNT_DETACH); err != nil {
        fmt.Println("put_old 마운트 해제 실패:", err)
        return
    }

    // put_old 디렉토리 삭제
    if err := os.Remove("/put_old"); err != nil {
        fmt.Println("put_old 삭제 실패:", err)
        return
    }

    // 컨테이너 실행
    if err := cmd.Start(); err != nil {
        fmt.Println("컨테이너 실행 실패:", err)
        return
    }

    fmt.Printf("컨테이너가 PID %d에서 실행 중입니다.\n", cmd.Process.Pid)

    containerID := fmt.Sprintf("%d", cmd.Process.Pid)

    // CNI 네트워크 설정
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

    if err := cmd.Wait(); err != nil {
        fmt.Println("컨테이너 종료 실패:", err)
    } else {
        fmt.Println("컨테이너 실행 성공")
    }
}

// 컨테이너 정보 저장 함수
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

