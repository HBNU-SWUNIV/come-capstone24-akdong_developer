package subsystem

import (
    "fmt"
    "os/exec"
)

// 네트워크 네임스페이스 생성 및 veth 페어를 설정하여 컨테이너에 IP 할당
func SetupNetworkNamespace(containerID string, ipAddr string) error {
    // veth 페어 이름 (컨테이너와 호스트 양쪽 인터페이스)
    vethHost := "veth_" + containerID
    vethContainer := "eth0_" + containerID

    // veth 페어 생성
    if err := exec.Command("ip", "link", "add", vethHost, "type", "veth", "peer", "name", vethContainer).Run(); err != nil {
        return fmt.Errorf("veth 페어 생성 실패: %v", err)
    }

    // 호스트 측 veth를 브리지 네트워크에 연결
    if err := exec.Command("ip", "link", "set", vethHost, "up").Run(); err != nil {
        return fmt.Errorf("호스트 측 veth 설정 실패: %v", err)
    }

    // 브리지에 호스트 측 veth 연결
    if err := exec.Command("ip", "link", "set", vethHost, "master", "br0").Run(); err != nil {
        return fmt.Errorf("브리지에 veth 연결 실패: %v", err)
    }

    // 네트워크 네임스페이스에서 컨테이너 측 veth 활성화 및 IP 할당
    if err := exec.Command("ip", "link", "set", vethContainer, "netns", containerID).Run(); err != nil {
        return fmt.Errorf("네임스페이스로 veth 이동 실패: %v", err)
    }

    if err := exec.Command("ip", "netns", "exec", "/proc/"+containerID+"/ns/net", "ip", "link", "set", vethContainer, "up").Run(); err != nil {
        return fmt.Errorf("컨테이너 측 veth 활성화 실패: %v", err)
    }

    // 컨테이너 측 IP 할당
    if err := exec.Command("ip", "netns", "exec", "/proc/"+containerID+"/ns/net", "ip", "addr", "add", ipAddr, "dev", vethContainer).Run(); err != nil {
        return fmt.Errorf("IP 주소 할당 실패: %v", err)
    }

    return nil
}

