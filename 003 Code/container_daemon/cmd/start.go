package cmd

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"
    "golang.org/x/sys/unix"
    "github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
    Use:   "start [containerName]",
    Short: "Container start",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        containerName := args[0]
        containerPath := "/CarteTest/container/" + containerName

        // Cgroups 설정
        if err := setupCgroups(containerPath); err != nil {
            log.Fatal("Failed to setup cgroups:", err)
        }

        // 컨테이너 실행
        fmt.Println("Attempting to start container...")
        if err := startContainer(containerPath, containerName); err != nil {
            return fmt.Errorf("error starting container: %v", err)
        }

        return nil
    },
}

func init() {
    rootCmd.AddCommand(startCmd)
}

func startContainer(containerPath, containerName string) error {
    cmd, err := runInNewNamespace(containerPath, "/bin/busybox", []string{"sh"}, containerName)
    if err != nil {
        return fmt.Errorf("failed to start container in new namespace: %v", err)
    }

    // 네트워크 네임스페이스 설정을 cmd.Start() 이후로 이동
    if err := setupNetworkNamespace(cmd); err != nil {
        return fmt.Errorf("failed to setup network namespace: %v", err)
    }

    pid := cmd.Process.Pid
    if err := recordContainerPID(containerName, pid); err != nil {
        return fmt.Errorf("failed to record PID: %v", err)
    }

    fmt.Printf("Container %s started with PID %d\n", containerName, pid)
    if err := cmd.Wait(); err != nil {
        return fmt.Errorf("process finished with error: %v", err)
    }

    return nil
}   

func runInNewNamespace(containerPath, path string, args []string, containerName string) (*exec.Cmd, error) {
    fmt.Println("[DEBUG] Starting runInNewNamespace function")
    fmt.Printf("[DEBUG] Container path: %s\n", containerPath)
    fmt.Printf("[DEBUG] Command path: %s\n", path)

    // 필수 디렉토리 생성
    os.MkdirAll(filepath.Join(containerPath, "proc"), 0755)
    os.MkdirAll(filepath.Join(containerPath, "sys"), 0755)
    os.MkdirAll(filepath.Join(containerPath, "dev"), 0755)
    fmt.Println("[DEBUG] Created necessary directories in container path")

    // /dev 디렉토리에 장치 파일 생성
    if err := syscall.Mknod(filepath.Join(containerPath, "dev/null"), syscall.S_IFCHR|0666, int(unix.Mkdev(1, 3))); err != nil && !os.IsExist(err) {
        return nil, fmt.Errorf("failed to create /dev/null: %v", err)
    }
    if err := syscall.Mknod(filepath.Join(containerPath, "dev/zero"), syscall.S_IFCHR|0666, int(unix.Mkdev(1, 5))); err != nil && !os.IsExist(err) {
        return nil, fmt.Errorf("failed to create /dev/zero: %v", err)
    }
    if err := syscall.Mknod(filepath.Join(containerPath, "dev/random"), syscall.S_IFCHR|0666, int(unix.Mkdev(1, 8))); err != nil && !os.IsExist(err) {
        return nil, fmt.Errorf("failed to create /dev/random: %v", err)
    }
    if err := syscall.Mknod(filepath.Join(containerPath, "dev/urandom"), syscall.S_IFCHR|0666, int(unix.Mkdev(1, 9))); err != nil && !os.IsExist(err) {
        return nil, fmt.Errorf("failed to create /dev/urandom: %v", err)
    }
    fmt.Println("[DEBUG] Created device files in /dev")

    // containerPath를 자신에게 bind mount
    if err := syscall.Mount(containerPath, containerPath, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
        return nil, fmt.Errorf("failed to bind mount container path: %v", err)
    }
    fmt.Printf("[DEBUG] Successfully bind-mounted container path %s\n", containerPath)

    // containerPath를 private mount로 설정
    if err := syscall.Mount("", containerPath, "", syscall.MS_PRIVATE, ""); err != nil {
        return nil, fmt.Errorf("failed to make container path private: %v", err)
    }
    fmt.Println("[DEBUG] Set container path as private mount")

    // /proc, /sys 마운트
    if err := syscall.Mount("proc", filepath.Join(containerPath, "proc"), "proc", 0, ""); err != nil {
        return nil, fmt.Errorf("failed to mount /proc: %v", err)
    }
    if err := syscall.Mount("sysfs", filepath.Join(containerPath, "sys"), "sysfs", 0, ""); err != nil {
        return nil, fmt.Errorf("failed to mount /sys: %v", err)
    }
    fmt.Println("[DEBUG] Mounted /proc and /sys in container path")

    // 현재 작업 디렉토리를 containerPath로 변경
    if err := os.Chdir(containerPath); err != nil {
        return nil, fmt.Errorf("failed to change directory to container path: %v", err)
    }
    fmt.Printf("[DEBUG] Changed working directory to container path %s\n", containerPath)

    cmd := exec.Command(path, args...)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
        AmbientCaps: []uintptr{unix.CAP_NET_ADMIN}, // CAP_NET_ADMIN 권한 추가
    }
    cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

    // chroot를 사용하여 루트 디렉토리를 변경
    fmt.Println("[DEBUG] Attempting chroot...")
    if err := syscall.Chroot(containerPath); err != nil {
        return nil, fmt.Errorf("failed to chroot to container path: %v", err)
    }
    fmt.Println("[DEBUG] chroot succeeded")

    // 새로운 루트로 이동
    if err := os.Chdir("/"); err != nil {
        return nil, fmt.Errorf("failed to change to new root directory after chroot: %v", err)
    }
    fmt.Println("[DEBUG] Changed directory to new root")

    // 프로세스를 시작하여 cmd.Process가 nil이 아닌 상태로 변경
    if err := cmd.Start(); err != nil {
        return nil, fmt.Errorf("failed to run command in new namespace: %v", err)
    }
    fmt.Println("[DEBUG] Command started successfully")

    return cmd, nil
}

// 컨테이너는 들어가지지만 네트워크 안됨
func setupNetworkNamespace(cmd *exec.Cmd) error {
    pid := cmd.Process.Pid
    netnsName := fmt.Sprintf("netns_%d", pid)
    vethHost := fmt.Sprintf("veth_host_%d", pid)
    vethContainer := fmt.Sprintf("veth_cont_%d", pid)

    // 기존 네임스페이스와 veth 인터페이스 제거
    exec.Command("ip", "netns", "del", netnsName).Run()
    exec.Command("ip", "link", "del", vethHost).Run()

    // 네임스페이스 생성
    if output, err := exec.Command("ip", "netns", "add", netnsName).CombinedOutput(); err != nil {
        return fmt.Errorf("[ERROR] Failed to add netns: %v\nOutput: %s", err, output)
    }
    fmt.Println("[DEBUG] Network namespace created")

    // veth 페어 생성
    if output, err := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethContainer).CombinedOutput(); err != nil {
        return fmt.Errorf("[ERROR] Failed to create veth pair: %v\nOutput: %s", err, output)
    }
    fmt.Println("[DEBUG] veth pair created successfully")

    // vethContainer를 네임스페이스로 이동
    if output, err := exec.Command("ip", "link", "set", vethContainer, "netns", netnsName).CombinedOutput(); err != nil {
        return fmt.Errorf("[ERROR] Failed to move vethContainer to netns: %v\nOutput: %s", err, output)
    }
    fmt.Println("[DEBUG] vethContainer moved to network namespace")

    // 호스트 쪽 vethHost에 IP 주소 할당 및 활성화
    if output, err := exec.Command("ip", "addr", "add", "192.168.1.1/24", "dev", vethHost).CombinedOutput(); err != nil {
        return fmt.Errorf("[ERROR] Failed to assign IP to vethHost: %v\nOutput: %s", err, output)
    }
    if output, err := exec.Command("ip", "link", "set", vethHost, "up").CombinedOutput(); err != nil {
        return fmt.Errorf("[ERROR] Failed to bring up vethHost: %v\nOutput: %s", err, output)
    }
    fmt.Println("[DEBUG] vethHost IP assigned and interface brought up")

    // 네임스페이스 내에서 vethContainer에 IP 주소 할당 및 인터페이스 활성화
    if output, err := exec.Command("ip", "netns", "exec", netnsName, "ip", "addr", "add", "192.168.1.2/24", "dev", vethContainer).CombinedOutput(); err != nil {
        return fmt.Errorf("[ERROR] Failed to assign IP to vethContainer: %v\nOutput: %s", err, output)
    }
    if output, err := exec.Command("ip", "netns", "exec", netnsName, "ip", "link", "set", vethContainer, "up").CombinedOutput(); err != nil {
        return fmt.Errorf("[ERROR] Failed to bring up vethContainer: %v\nOutput: %s", err, output)
    }
    fmt.Println("[DEBUG] vethContainer IP assigned and interface brought up")

    // 네임스페이스 내 루프백 인터페이스 활성화
    if output, err := exec.Command("ip", "netns", "exec", netnsName, "ip", "link", "set", "lo", "up").CombinedOutput(); err != nil {
        return fmt.Errorf("[ERROR] Failed to bring up loopback in netns: %v\nOutput: %s", err, output)
    }
    fmt.Println("[DEBUG] Loopback interface brought up in network namespace")

    // 네임스페이스 내 기본 게이트웨이 설정
    if output, err := exec.Command("ip", "netns", "exec", netnsName, "ip", "route", "add", "default", "via", "192.168.1.1").CombinedOutput(); err != nil {
        return fmt.Errorf("[ERROR] Failed to set default route in netns: %v\nOutput: %s", err, output)
    }
    fmt.Println("[DEBUG] Default route set in network namespace")

    return nil
}

func recordContainerPID(containerName string, pid int) error {
    pidFilePath := fmt.Sprintf("pid")
    // pidFilePath := fmt.Sprintf("/CarteTest/container/%s/pid", containerName)
    pidDir := filepath.Dir(pidFilePath)

    // 디렉토리 생성
    if err := os.MkdirAll(pidDir, 0755); err != nil {
        return fmt.Errorf("failed to create PID directory: %v", err)
    }

    pidFile, err := os.Create(pidFilePath)
    if err != nil {
        return fmt.Errorf("failed to create PID file: %v", err)
    }
    defer pidFile.Close()

    _, err = pidFile.WriteString(fmt.Sprintf("%d", pid))
    return err
}


func setupCgroups(containerPath string) error {
    cgroupRoot := "/CarteTest/cgroup"
    pid := os.Getpid()

    if err := os.MkdirAll(cgroupRoot, 0755); err != nil {
        return fmt.Errorf("failed to create cgroup path: %v", err)
    }

    // CPU와 메모리 리소스 제한 설정
    cpuLimitPath := filepath.Join(cgroupRoot, "cpu", "myContainerGroup")
    memoryLimitPath := filepath.Join(cgroupRoot, "memory", "myContainerGroup")

    os.MkdirAll(cpuLimitPath, 0755)
    os.MkdirAll(memoryLimitPath, 0755)

    // CPU 쿼터 설정
    os.WriteFile(filepath.Join(cpuLimitPath, "cpu.cfs_quota_us"), []byte("100000"), 0644)

    // 메모리 제한 설정 (50MB)
    os.WriteFile(filepath.Join(memoryLimitPath, "memory.limit_in_bytes"), []byte("52428800"), 0644)

    // 현재 프로세스를 cgroup에 추가
    os.WriteFile(filepath.Join(cpuLimitPath, "tasks"), []byte(fmt.Sprintf("%d", pid)), 0644)
    os.WriteFile(filepath.Join(memoryLimitPath, "tasks"), []byte(fmt.Sprintf("%d", pid)), 0644)

    return nil
}
