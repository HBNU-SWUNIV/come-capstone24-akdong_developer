package daemon

import (
    "os/exec"
    "syscall"
)

func SetNamespace() *exec.Cmd {
    cmd := exec.Command("sh")
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
    }
    return cmd
}
