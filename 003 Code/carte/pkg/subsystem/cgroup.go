package subsystem

import (
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
    "strconv"
    "strings"
)

// cgroups 경로 설정 (보통 /sys/fs/cgroup 아래)
const cgroupRoot = "/sys/fs/cgroup"

// cgroups 설정 함수 (CPU 및 메모리 제한)
func SetCgroupLimits(containerID string, cpuLimit string, memoryLimit string) error {
    cpuLimit = strings.TrimSuffix(cpuLimit, "%")
    cpuQuota, err := strconv.Atoi(cpuLimit)
    if err != nil {
        return fmt.Errorf("CPU 제한 설정 실패: %v", err)
    }

    cpuCgroupPath := filepath.Join(cgroupRoot, "cpu", "carte", containerID)
    memoryCgroupPath := filepath.Join(cgroupRoot, "memory", "carte", containerID)

    // cgroup 디렉토리 생성
    if err := os.MkdirAll(cpuCgroupPath, 0755); err != nil {
        return fmt.Errorf("CPU cgroup 디렉토리 생성 실패: %v", err)
    }
    if err := os.MkdirAll(memoryCgroupPath, 0755); err != nil {
        return fmt.Errorf("메모리 cgroup 디렉토리 생성 실패: %v", err)
    }

    // CPU 제한 설정
    if err := ioutil.WriteFile(filepath.Join(cpuCgroupPath, "cpu.cfs_quota_us"), []byte(strconv.Itoa(cpuQuota*1000)), 0644); err != nil {
        return fmt.Errorf("CPU 제한 설정 실패: %v", err)
    }

    // 메모리 제한 설정
    if err := ioutil.WriteFile(filepath.Join(memoryCgroupPath, "memory.limit_in_bytes"), []byte(memoryLimit), 0644); err != nil {
        return fmt.Errorf("메모리 제한 설정 실패: %v", err)
    }

    // 프로세스 PID를 cgroups에 추가
    pid := strconv.Itoa(os.Getpid())
    if err := ioutil.WriteFile(filepath.Join(cpuCgroupPath, "tasks"), []byte(pid), 0644); err != nil {
        return fmt.Errorf("PID 추가 실패: %v", err)
    }
    if err := ioutil.WriteFile(filepath.Join(memoryCgroupPath, "tasks"), []byte(pid), 0644); err != nil {
        return fmt.Errorf("PID 추가 실패: %v", err)
    }

    return nil
}

