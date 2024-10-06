package subsystem

import (
    "fmt"
    "os/exec"
)

// 브리지 설정 함수
func SetupBridge() error {
    // 브리지가 존재하지 않을 경우 생성
    if err := exec.Command("ip", "link", "show", "br0").Run(); err != nil {
        fmt.Println("브리지가 존재하지 않음, 생성 중...")
        if err := exec.Command("ip", "link", "add", "br0", "type", "bridge").Run(); err != nil {
            return fmt.Errorf("브리지 생성 실패: %v", err)
        }
    }

    // 브리지 활성화
    if err := exec.Command("ip", "link", "set", "br0", "up").Run(); err != nil {
        return fmt.Errorf("브리지 활성화 실패: %v", err)
    }

    fmt.Println("브리지 br0 활성화 완료")
    return nil
}

// veth 페어 생성 및 설정
func SetupVethPair(containerID, vethHost, vethContainer string) error {
    // veth 페어 생성
    if err := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethContainer).Run(); err != nil {
        return fmt.Errorf("veth 페어 생성 실패: %v", err)
    }

    // 호스트 측 veth를 브리지에 연결
    if err := exec.Command("ip", "link", "set", vethHost, "master", "br0").Run(); err != nil {
        return fmt.Errorf("브리지에 veth 연결 실패: %v", err)
    }

    // veth 활성화
    if err := exec.Command("ip", "link", "set", vethHost, "up").Run(); err != nil {
        return fmt.Errorf("호스트 측 veth 활성화 실패: %v", err)
    }

    fmt.Printf("veth 페어 %s <-> %s 생성 및 연결 완료\n", vethHost, vethContainer)

    return nil
}

// 네임스페이스에서 veth 활성화 및 IP 할당
func ActivateVethInContainer(containerID, vethContainer, ipAddr string) error {
    // 네트워크 네임스페이스에 veth 활성화
    if err := exec.Command("ip", "netns", "exec", "/proc/"+containerID+"/ns/net", "ip", "link", "set", vethContainer, "up").Run(); err != nil {
        return fmt.Errorf("컨테이너 측 veth 활성화 실패: %v", err)
    }

    // 컨테이너에 IP 할당
    if err := exec.Command("ip", "netns", "exec", "/proc/"+containerID+"/ns/net", "ip", "addr", "add", ipAddr+"/24", "dev", vethContainer).Run(); err != nil {
        return fmt.Errorf("IP 할당 실패: %v", err)
    }

    fmt.Printf("컨테이너 측 veth %s 활성화 및 IP 할당 완료\n", vethContainer)
    return nil
}


