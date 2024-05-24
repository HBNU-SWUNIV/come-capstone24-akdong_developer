package models

import (
    "fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"

	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"time"
)

type Manifest struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

type Config struct {
	Architecture string    `json:"architecture"`
	Created      time.Time `json:"created"`
	OS           string    `json:"os"`
	Config       struct {
		Env        []string `json:"Env"`
		Entrypoint []string `json:"Entrypoint"`
		Cmd        []string `json:"Cmd"`
	} `json:"config"`
	RootFS struct {
		Type    string   `json:"type"`
		DiffIDs []string `json:"diff_ids"`
	} `json:"rootfs"`
}



func CreateContainer() error {
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
	cmd := exec.Command("/bin/sh", "-c", "hostname; echo '------'; ps aux; echo '------'; ip a; echo '------'; ipcs")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

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

func setupUTSNamespace() {
	// 호스트 이름 변경
	cmd := exec.Command("/bin/hostname", "container1")
	cmd.Run()
}

func setupPIDNamespace() {
	// 프로세스 ID 변경
	syscall.Sethostname([]byte("container1"))

	// PID 변경
	cmd := exec.Command("/bin/sh", "-c", "echo 1 > /proc/self/ns/pid")
	cmd.Run()
}

func setupNetworkNamespace() {
	// 네트워크 설정 (생략)
	cmd := exec.Command("/bin/sh", "-c", "ip link add veth0 type veth peer name veth1")
	cmd.Run()
}

func setupIPCNamespace() {
	// IPC 설정 (생략)
	cmd := exec.Command("/bin/sh", "-c", "ipcmk -M 1024")
	cmd.Run()
}

func setupMountNamespace() {
	// 파일 시스템 설정 (생략)
	cmd := exec.Command("/bin/sh", "-c", "mkdir /mnt/containerroot; mount -t tmpfs none /mnt/containerroot")
	cmd.Run()
}









func BuildImage(filename string) error {
	// Create tar.gz file
	tarFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating tar file: %v", err)
	}
	defer tarFile.Close()

	gzipWriter := gzip.NewWriter(tarFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Add files to tar
	err = addFileToTar(tarWriter, "manifest.json", createManifest())
	if err != nil {
		return err
	}

	err = addFileToTar(tarWriter, "config.json", createConfig())
	if err != nil {
		return err
	}

	err = addLayerToTar(tarWriter, "layer.tar")
	if err != nil {
		return err
	}

	fmt.Println("Image tarball created successfully.")
	return nil
}

func addFileToTar(tarWriter *tar.Writer, name string, content []byte) error {
	header := &tar.Header{
		Name: name,
		Size: int64(len(content)),
		Mode: 0600,
	}
	err := tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("error writing tar header: %v", err)
	}
	_, err = tarWriter.Write(content)
	if err != nil {
		return fmt.Errorf("error writing tar content: %v", err)
	}
	return nil
}

func addLayerToTar(tarWriter *tar.Writer, name string) error {
	// Add file system layer
	// Example: add an empty directory for demonstration
	header := &tar.Header{
		Name: name,
		Mode: 0755,
		Size: 0,
	}
	err := tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("error writing tar layer header: %v", err)
	}
	return nil
}

func createManifest() []byte {
	manifest := Manifest{
		Config:   "config.json",
		RepoTags: []string{"myimage:latest"},
		Layers:   []string{"layer.tar"},
	}
	manifestBytes, _ := json.Marshal(manifest)
	return manifestBytes
}

func createConfig() []byte {
	config := Config{
		Architecture: "amd64",
		Created:      time.Now(),
		OS:           "linux",
	}
	config.Config.Env = []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}
	config.Config.Entrypoint = []string{"/bin/sh"}
	config.RootFS.Type = "layers"
	config.RootFS.DiffIDs = []string{"sha256:<layer-sha256-hash>"}
	configBytes, _ := json.Marshal(config)
	return configBytes
}