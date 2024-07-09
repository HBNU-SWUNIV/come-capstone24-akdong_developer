package daemon

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
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
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET | syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: syscall.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: syscall.Getgid(), Size: 1},
		},
	}

	return cmd
}
