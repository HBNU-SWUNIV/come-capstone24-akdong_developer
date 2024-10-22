package container

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"
    "carte/pkg/subsystem"
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

// MountNamespaceContainer는 마운트 네임스페이스를 사용하여 컨테이너 파일 시스템을 분리합니다.
func MountNamespaceContainer(rootFsPath string) error {
    // 마운트 네임스페이스에서 파일 시스템을 분리
    if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
        return fmt.Errorf("파일 시스템 분리 실패: %v", err)
    }

    // rootfs를 새로운 루트로 마운트 (read/write로 마운트)
    if err := syscall.Mount(rootFsPath, rootFsPath, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
        return fmt.Errorf("루트 파일 시스템 마운트 실패: %v", err)
    }

    // rootfs 내부에 proc 디렉토리가 없으면 생성
    procPath := filepath.Join(rootFsPath, "proc")
    if _, err := os.Stat(procPath); os.IsNotExist(err) {
        if err := os.Mkdir(procPath, 0755); err != nil {
            return fmt.Errorf("proc 디렉토리 생성 실패: %v", err)
        }
    }

    // rootfs 내부에 proc 파일 시스템을 마운트
    if err := syscall.Mount("proc", procPath, "proc", 0, ""); err != nil {
        return fmt.Errorf("proc 마운트 실패: %v", err)
    }

    return nil
}


func RunContainer(name, imageID, cpuLimit, memoryLimit string) {
    imagesDir := "/var/run/carte/images/"
    rootFsPath := filepath.Join(imagesDir, imageID, "rootfs")  // rootfs 디렉토리 확인

    if _, err := os.Stat(rootFsPath); os.IsNotExist(err) {
        fmt.Printf("이미지 %s를 찾을 수 없습니다.\n", imageID)
        return
    }

    // 새로운 네임스페이스에서 프로세스를 실행
    cmd := exec.Command("/bin/sh")
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
    }

    // 표준 입출력 연결
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // 마운트 네임스페이스를 설정
    if err := MountNamespaceContainer(rootFsPath); err != nil {
        fmt.Printf("마운트 네임스페이스 설정 실패: %v\n", err)
        return
    }

    // chroot 사용 대신에 파일 시스템 마운트를 설정한 후 진행
    if err := syscall.Chdir(rootFsPath); err != nil {
        fmt.Printf("루트 파일 시스템 변경 실패: %v\n", err)
        return
    }

    // 컨테이너 실행
    if err := cmd.Start(); err != nil {
        fmt.Printf("컨테이너 실행 실패: %v\n", err)
        return
    }

    fmt.Printf("컨테이너가 PID %d에서 실행 중입니다.\n", cmd.Process.Pid)

    containerID := fmt.Sprintf("%d", cmd.Process.Pid)  // PID를 containerID로 사용

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
