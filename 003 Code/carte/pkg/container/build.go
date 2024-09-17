package container

import (
    "archive/tar"
    "bufio"
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "github.com/google/uuid"
)

// BuildConfig 구조체: 명령어와 설정을 저장
type BuildConfig struct {
    BaseImage string   // 베이스 이미지
    Commands  []string // RUN 명령어들
}

// buildContainer 함수: Cartefile을 읽고 이미지 빌드
func BuildContainer(cartefilePath string, imageName string) {
    imagesDir := "/var/run/carte/images/"
    createDirIfNotExists(imagesDir)

    imageID := uuid.New().String() // 이미지 ID 생성
    imagePath := filepath.Join(imagesDir, imageID)
    createDirIfNotExists(imagePath)

    // 1. Cartefile 읽기
    buildConfig, err := parseCartefile(cartefilePath)
    if err != nil {
        fmt.Printf("Cartefile 읽기 실패: %v\n", err)
        return
    }

    // 2. 베이스 이미지 파일 시스템 복사
    baseImagePath := filepath.Join(imagesDir, buildConfig.BaseImage)
    if _, err := os.Stat(baseImagePath); os.IsNotExist(err) {
        fmt.Printf("베이스 이미지 %s를 찾을 수 없습니다.\n", buildConfig.BaseImage)
        return
    }
    err = copyBaseImageLayer(baseImagePath, imagePath)
    if err != nil {
        fmt.Printf("베이스 이미지 레이어 복사 실패: %v\n", err)
        return
    }

    // 3. RUN 명령어 실행 및 레이어 생성
    for _, cmd := range buildConfig.Commands {
        fmt.Printf("명령어 실행: %s\n", cmd)
        err := executeCommandAndCreateLayer(cmd, imagePath)
        if err != nil {
            fmt.Printf("명령어 실행 실패: %v\n", err)
            return
        }
    }

    // 4. 이미지 메타데이터(config.json) 작성
    config := generateOCIConfig(buildConfig)
    configPath := filepath.Join(imagePath, "config.json")
    if err := ioutil.WriteFile(configPath, config, 0644); err != nil {
        fmt.Printf("이미지 메타데이터 저장 실패: %v\n", err)
        return
    }

    fmt.Printf("이미지 %s가 성공적으로 빌드되었습니다. ID: %s\n", imageName, imageID)
}

// Cartefile을 읽어서 BuildConfig로 변환하는 함수
func parseCartefile(cartefilePath string) (BuildConfig, error) {
    file, err := os.Open(cartefilePath)
    if err != nil {
        return BuildConfig{}, err
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    var config BuildConfig

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if strings.HasPrefix(line, "FROM ") {
            config.BaseImage = strings.TrimPrefix(line, "FROM ")
        } else if strings.HasPrefix(line, "RUN ") {
            config.Commands = append(config.Commands, strings.TrimPrefix(line, "RUN "))
        }
    }

    if err := scanner.Err(); err != nil {
        return BuildConfig{}, err
    }

    if config.BaseImage == "" {
        return BuildConfig{}, fmt.Errorf("베이스 이미지가 정의되지 않았습니다.")
    }

    return config, nil
}

// 베이스 이미지 복사 함수 (이전과 동일)
func copyBaseImageLayer(baseImagePath, imagePath string) error {
    srcFile, err := os.Open(filepath.Join(baseImagePath, "layer.tar"))
    if err != nil {
        return err
    }
    defer srcFile.Close()

    dstFile, err := os.Create(filepath.Join(imagePath, "layer.tar"))
    if err != nil {
        return err
    }
    defer dstFile.Close()

    _, err = io.Copy(dstFile, srcFile)
    return err
}

// 명령어 실행 및 레이어 생성 함수 (이전과 동일)
func executeCommandAndCreateLayer(command, imagePath string) error {
    layerPath := filepath.Join(imagePath, "new_layer.tar")
    tarFile, err := os.Create(layerPath)
    if err != nil {
        return err
    }
    defer tarFile.Close()

    tarWriter := tar.NewWriter(tarFile)
    defer tarWriter.Close()

    // 임시 예시로 빈 레이어 생성
    if err := tarWriter.WriteHeader(&tar.Header{
        Name: "example.txt",
        Mode: 0600,
        Size: 0,
    }); err != nil {
        return err
    }

    return nil
}

// OCI 메타데이터 생성 함수 (이전과 동일)
func generateOCIConfig(buildConfig BuildConfig) []byte {
    config := fmt.Sprintf(`
{
    "ociVersion": "1.0.0",
    "process": {
        "terminal": true,
        "user": { "uid": 0, "gid": 0 },
        "args": [%q],
        "env": [],
        "cwd": "/",
        "capabilities": {
            "bounding": ["CAP_AUDIT_WRITE", "CAP_KILL", "CAP_NET_BIND_SERVICE"]
        }
    },
    "root": {
        "path": "rootfs"
    },
    "mounts": [
        { "destination": "/proc", "type": "proc", "source": "proc" }
    ],
    "linux": {
        "namespaces": [
            { "type": "pid" },
            { "type": "network" }
        ]
    }
}
`, buildConfig.Commands)
    return []byte(config)
}

// 중복되는 디렉토리 생성 함수 (이전과 동일)
func createDirIfNotExists(dir string) {
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        os.MkdirAll(dir, 0755)
    }
}
