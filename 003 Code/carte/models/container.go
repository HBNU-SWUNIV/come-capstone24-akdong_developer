package models

import (
	"fmt"
	"os"
	"os/exec"
)

func CreateContainer(commands []string) error {
	// cgroups를 사용하여 메모리 제한 설정
	err := cgroups()
	if err != nil {
		return err
	}

	// 각 네임스페이스 설정 함수 호출
	setupUTSNamespace()
	setupPIDNamespace()
	setupNetworkNamespace()
	setupIPCNamespace()
	setupMountNamespace()

	// 컨테이너 내부에서 호스트 이름, PID, IP 주소, IPC 등을 확인하기 위해 명령어 실행
	for _, command := range commands {
		cmd := exec.Command("/bin/sh", "-c", command)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error running command '%s': %v", command, err)
		}
	}

	return nil
}


