package daemon

import (
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "strconv"
    "syscall"
    "log"
)

func setupCgroups() error {
    cgroups := "/sys/fs/cgroup/"
    pid := os.Getpid()

    // Memory limit
    if err := os.MkdirAll(filepath.Join(cgroups, "memory", "carte"), 0755); err != nil {
        return err
    }
    if err := ioutil.WriteFile(filepath.Join(cgroups, "memory", "carte", "memory.limit_in_bytes"), []byte("104857600"), 0700); err != nil {
        return err
    }
    if err := ioutil.WriteFile(filepath.Join(cgroups, "memory", "carte", "cgroup.procs"), []byte(strconv.Itoa(pid)), 0700); err != nil {
        return err
    }

    // CPU limit
    if err := os.MkdirAll(filepath.Join(cgroups, "cpu", "carte"), 0755); err != nil {
        return err
    }
    if err := ioutil.WriteFile(filepath.Join(cgroups, "cpu", "carte", "cpu.cfs_quota_us"), []byte("50000"), 0700); err != nil {
        return err
    }
    if err := ioutil.WriteFile(filepath.Join(cgroups, "cpu", "carte", "cgroup.procs"), []byte(strconv.Itoa(pid)), 0700); err != nil {
        return err
    }

    return nil
}

func setNamespacesAndCgroups(command string) *exec.Cmd {
    cmd := exec.Command("sh", "-c", command)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET, // 필요한 네임스페이스만 설정
    }
    cmd.Env = append(os.Environ(), "PATH=/usr/sbin:/usr/bin:/sbin:/bin")

    return cmd
}

func StartContainer(image, command string) error {
    cmd := setNamespacesAndCgroups(command)

    if err := setupCgroups(); err != nil {
        log.Printf("Error setting up cgroups: %v", err)
        return err
    }

    stdout, err := cmd.StdoutPipe()
    if err != nil {
        log.Printf("Error creating stdout pipe: %v", err)
        return err
    }
    stderr, err := cmd.StderrPipe()
    if err != nil {
        log.Printf("Error creating stderr pipe: %v", err)
        return err
    }

    if err := cmd.Start(); err != nil {
        log.Printf("Error starting command: %v", err)
        return err
    }

    go func() {
        out, _ := ioutil.ReadAll(stdout)
        log.Printf("Command output: %s", out)
    }()

    go func() {
        errOut, _ := ioutil.ReadAll(stderr)
        log.Printf("Command error output: %s", errOut)
    }()

    if err := cmd.Wait(); err != nil {
        log.Printf("Error waiting for command: %v", err)
        return err
    }

    return nil
}


