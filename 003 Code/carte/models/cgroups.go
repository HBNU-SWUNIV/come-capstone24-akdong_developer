package models

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func cgroups() error {
	// cgroups 경로
	cgroup := "/sys/fs/cgroup/"
	pid := os.Getpid()
	memLimit := "100000000" // 예: 100MB

	// 메모리 cgroup 설정
	memCgroupPath := filepath.Join(cgroup, "memory", "mycontainer")
	err := os.Mkdir(memCgroupPath, 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(memCgroupPath, "memory.limit_in_bytes"), []byte(memLimit), 0644)
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(memCgroupPath, "cgroup.procs"), []byte(strconv.Itoa(pid)), 0644)
	if err != nil {
		return err
	}

	fmt.Println("Container with limited memory running...")
	return nil
}


