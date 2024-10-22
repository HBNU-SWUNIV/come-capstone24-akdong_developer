package subsystem

import (
    //"fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strconv"
)

// cgroups 설정 함수 (CPU 및 메모리 제한)
func SetCgroupLimits(containerID string, cpuLimit string, memoryLimit string) error {
    cpuLimit = cpuLimit[:len(cpuLimit)-1] // "100%"와 같은 형식에서 "%" 제거
    cpuQuota, _ := strconv.Atoi(cpuLimit)

    cpuCgroupPath := filepath.Join("/sys/fs/cgroup/cpu", containerID)
    memoryCgroupPath := filepath.Join("/sys/fs/cgroup/memory", containerID)

    os.MkdirAll(cpuCgroupPath, 0755)
    os.MkdirAll(memoryCgroupPath, 0755)

    ioutil.WriteFile(filepath.Join(cpuCgroupPath, "cpu.cfs_quota_us"), []byte(strconv.Itoa(cpuQuota*1000)), 0644)
    ioutil.WriteFile(filepath.Join(memoryCgroupPath, "memory.limit_in_bytes"), []byte(memoryLimit), 0644)

    return nil
}
