package container

import (
    "fmt"
    "path/filepath"
    "syscall"
    "io/ioutil"
    "encoding/json"
    "os/exec"
)

// 컨테이너 시작 함수
func StartContainer(name string) {
    containersDir := "/var/run/carte/containers/"
    containerFile := filepath.Join(containersDir, name+".json")

    // 컨테이너 정보 로드
    containerInfo, err := loadContainerInfo(containerFile)
    if err != nil {
        fmt.Printf("컨테이너 정보 로드 실패: %v\n", err)
        return
    }

    if containerInfo.Status != "stopped" {
        fmt.Println("중지된 상태의 컨테이너만 시작할 수 있습니다.")
        return
    }

    // 컨테이너 다시 시작
    cmd := exec.Command("/bin/sh")
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
    }

    if err := cmd.Start(); err != nil {
        fmt.Println("컨테이너 시작 실패:", err)
        return
    }

    containerInfo.PID = cmd.Process.Pid
    containerInfo.Status = "running"

    // 컨테이너 정보 저장
    if err := saveContainerInfoToFile(containerInfo, containerFile); err != nil {
        fmt.Printf("컨테이너 정보 저장 실패: %v\n", err)
        return
    }

    fmt.Printf("컨테이너 %s가 시작되었습니다.\n", name)
}

// 컨테이너 정보 저장 함수
func saveContainerInfoToFile(containerInfo ContainerInfo, filePath string) error {
    data, err := json.MarshalIndent(containerInfo, "", "    ")
    if err != nil {
        return err
    }
    return ioutil.WriteFile(filePath, data, 0644)
}

