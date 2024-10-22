package subsystem

import (
    "context"
    "fmt"
    "github.com/containernetworking/cni/libcni"
    "github.com/containernetworking/cni/pkg/types/040" // CNI types version
    "os"
    "path/filepath"
)

func SetupCNINetwork(containerID string) (string, error) {
    cniConfPath := "/etc/cni/net.d"
    cniPath := "/opt/cni/bin"

    // CNI 구성 로드
    cniConfig := libcni.NewCNIConfig([]string{cniPath}, nil)

    // CNI 네트워크 목록 로드
    confFiles, err := libcni.ConfFiles(cniConfPath, []string{".conf", ".conflist"})
    if err != nil {
        return "", fmt.Errorf("CNI 설정 파일 로드 실패: %v", err)
    }

    if len(confFiles) == 0 {
        return "", fmt.Errorf("CNI 설정 파일을 찾을 수 없습니다.")
    }

    confBytes, err := os.ReadFile(confFiles[0])
    if err != nil {
        return "", fmt.Errorf("CNI 설정 파일 읽기 실패: %v", err)
    }

    var networkConfig *libcni.NetworkConfigList
    if filepath.Ext(confFiles[0]) == ".conflist" {
        networkConfig, err = libcni.ConfListFromBytes(confBytes)
        if err != nil {
            return "", fmt.Errorf("CNI ConfList 생성 실패: %v", err)
        }
    } else {
        conf, err := libcni.ConfFromBytes(confBytes)
        if err != nil {
            return "", fmt.Errorf("CNI Conf 생성 실패: %v", err)
        }
        networkConfig, err = libcni.ConfListFromConf(conf)
        if err != nil {
            return "", fmt.Errorf("CNI 네트워크 설정 실패: %v", err)
        }
    }

    // PID 기반 NetNS 경로 설정
    runtimeConf := &libcni.RuntimeConf{
        ContainerID: containerID,
        NetNS:       fmt.Sprintf("/proc/%s/ns/net", containerID),  // 실제 PID 기반 네임스페이스
        IfName:      "eth0",
    }

    // 네트워크 적용
    result, err := cniConfig.AddNetworkList(context.Background(), networkConfig, runtimeConf)
    if err != nil {
        return "", fmt.Errorf("CNI 네트워크 추가 실패: %v", err)
    }

    // CNI 결과를 0.4.0 형식으로 변환하여 IP 정보 가져오기
    result040, err := types040.GetResult(result)
    if err != nil {
        return "", fmt.Errorf("CNI 결과 변환 실패: %v", err)
    }

    // IP 주소 추출
    if len(result040.IPs) == 0 {
        return "", fmt.Errorf("IP 설정을 찾을 수 없습니다.")
    }

    ipAddr := result040.IPs[0].Address.String()
    return ipAddr, nil
}
