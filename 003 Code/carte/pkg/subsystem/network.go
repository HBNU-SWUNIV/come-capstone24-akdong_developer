package subsystem

import (
    "fmt"
    "os/exec"
)

// 브리지 설정 함수 (carte0)
func SetupBridge() error {
    // 브리지 carte0가 존재하지 않을 경우 생성
    err := exec.Command("ip", "link", "show", "carte0").Run()
    if err != nil {
        fmt.Println("브리지 carte0가 존재하지 않음, 생성 중...")
        if err := exec.Command("ip", "link", "add", "carte0", "type", "bridge").Run(); err != nil {
            return fmt.Errorf("브리지 생성 실패: %v", err)
        }
    }

    // 브리지 활성화
    if err := exec.Command("ip", "link", "set", "carte0", "up").Run(); err != nil {
        return fmt.Errorf("브리지 활성화 실패: %v", err)
    }

    fmt.Println("브리지 carte0 활성화 완료")

    // 브리지에 IP가 이미 할당되었는지 확인한 후 IP 할당
    out, err := exec.Command("ip", "addr", "show", "carte0").Output()
    if err != nil || !containsIP(out) {
        fmt.Println("브리지에 IP가 할당되지 않음, IP 할당 중...")
        if err := exec.Command("ip", "addr", "add", "192.168.1.1/24", "dev", "carte0").Run(); err != nil {
            return fmt.Errorf("브리지에 IP 할당 실패: %v", err)
        }
    } else {
        fmt.Println("브리지에 이미 IP가 할당되어 있습니다.")
    }

    // IP 포워딩 활성화
    if err := EnableIPForwarding(); err != nil {
        return fmt.Errorf("IP 포워딩 설정 실패: %v", err)
    }

    // NAT 설정 추가 (외부 인터페이스를 맞게 지정해야 함)
    externalInterface := "enp6s0" // 실제 네트워크 인터페이스로 변경 필요
    if err := SetupNAT("carte0", externalInterface); err != nil {
        return fmt.Errorf("NAT 설정 실패: %v", err)
    }

    return nil
}

// 브리지에 IP가 이미 할당되었는지 확인하는 함수
func containsIP(output []byte) bool {
    return string(output) != ""
}

// veth 페어 생성 및 설정
func SetupVethPair(containerID, vethHost, vethContainer string) error {
    // veth 페어 생성
    if err := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethContainer).Run(); err != nil {
        return fmt.Errorf("veth 페어 생성 실패: %v", err)
    }

    // 호스트 측 veth를 carte0 브리지에 연결
    if err := exec.Command("ip", "link", "set", vethHost, "master", "carte0").Run(); err != nil {
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
    netnsName := "ns_" + containerID  // 네임스페이스 이름 지정

    // 네임스페이스 추가
    if err := exec.Command("ip", "netns", "add", netnsName).Run(); err != nil {
        return fmt.Errorf("네트워크 네임스페이스 추가 실패: %v", err)
    }

    // veth 네트워크 인터페이스를 네임스페이스로 이동
    if err := exec.Command("ip", "link", "set", vethContainer, "netns", netnsName).Run(); err != nil {
        return fmt.Errorf("veth 네임스페이스 이동 실패: %v", err)
    }

    // 네임스페이스에서 veth 활성화
    fmt.Printf("네임스페이스 경로: /proc/%s/ns/net\n", containerID)
    if err := exec.Command("ip", "netns", "exec", netnsName, "ip", "link", "set", vethContainer, "up").Run(); err != nil {
        return fmt.Errorf("컨테이너 측 veth 활성화 실패: %v", err)
    }

    // 네임스페이스 내에서 IP 할당
    if err := exec.Command("ip", "netns", "exec", netnsName, "ip", "addr", "add", ipAddr+"/24", "dev", vethContainer).Run(); err != nil {
        return fmt.Errorf("IP 할당 실패: %v", err)
    }

    // 기본 게이트웨이 설정 (네임스페이스 내)
    if err := exec.Command("ip", "netns", "exec", netnsName, "ip", "route", "add", "default", "via", "192.168.1.1").Run(); err != nil {
        return fmt.Errorf("기본 경로 설정 실패: %v", err)
    }

    fmt.Printf("veth %s 활성화 및 IP 할당 완료\n", vethContainer)

    // 네트워크 상태 확인을 위한 추가 코드
    out, err := exec.Command("ip", "netns", "exec", netnsName, "ip", "link", "show", vethContainer).Output()
    if err != nil {
        return fmt.Errorf("veth 상태 확인 실패: %v", err)
    }
    fmt.Printf("veth 상태 확인 결과: %s\n", string(out))

    return nil
}




// IP 포워딩 활성화 함수
func EnableIPForwarding() error {
    // IP 포워딩 활성화
    if err := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1").Run(); err != nil {
        return fmt.Errorf("IP 포워딩 활성화 실패: %v", err)
    }
    fmt.Println("IP 포워딩 활성화 완료")
    return nil
}

// NAT 설정 추가 함수
func SetupNAT(bridgeName, externalInterface string) error {
    // NAT 설정 추가
    if err := exec.Command("iptables", "-t", "nat", "-A", "POSTROUTING", "-s", "192.168.1.0/24", "-o", externalInterface, "-j", "MASQUERADE").Run(); err != nil {
        return fmt.Errorf("NAT 설정 실패: %v", err)
    }
    fmt.Println("NAT 설정 완료")
    return nil
}

