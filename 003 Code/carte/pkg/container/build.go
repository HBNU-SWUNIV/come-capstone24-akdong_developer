package container

import (
    "archive/tar"
    "bufio"       // bufio 사용 복구
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "path/filepath"
    "strings"     // strings 사용 복구
    "github.com/google/uuid"
)

// BuildConfig 구조체: 명령어와 설정을 저장
type BuildConfig struct {
    BaseImage string   // 베이스 이미지
    Commands  []string // RUN 명령어들
}

// buildContainer 함수 내에서 이미지 경로 확인 로직
func BuildContainer(cartefilePath string, imageName string) {
    imagesDir := "/var/run/carte/images/"
    createDirIfNotExists(imagesDir)

    // 이미지 경로 출력
    fmt.Printf("이미지 경로 확인: %s\n", imagesDir)

    imageID := uuid.New().String() // 이미지 ID 생성
    imagePath := filepath.Join(imagesDir, imageID)
    createDirIfNotExists(imagePath)

    // Cartefile 읽기
    buildConfig, err := parseCartefile(cartefilePath)
    if err != nil {
        fmt.Printf("Cartefile 읽기 실패: %v\n", err)
        return
    }

    // 베이스 이미지 경로 확인
    baseImagePath := filepath.Join(imagesDir, buildConfig.BaseImage)
    fmt.Printf("베이스 이미지 경로 확인: %s\n", baseImagePath)

    // 해당 경로가 올바른지 확인
    if _, err := os.Stat(filepath.Join(baseImagePath, "blobs")); os.IsNotExist(err) {
        fmt.Printf("베이스 이미지 %s를 찾을 수 없습니다.\n", buildConfig.BaseImage)
        return
    }

    err = copyBaseImageLayer(baseImagePath, imagePath)
    if err != nil {
        fmt.Printf("베이스 이미지 레이어 복사 실패: %v\n", err)
        return
    }

    // RUN 명령어 실행 및 레이어 생성
    for _, cmd := range buildConfig.Commands {
        fmt.Printf("명령어 실행: %s\n", cmd)
        err := executeCommandAndCreateLayer(cmd, imagePath)
        if err != nil {
            fmt.Printf("명령어 실행 실패: %v\n", err)
            return
        }
    }

    // 이미지 메타데이터(config.json) 작성
    config := generateOCIConfig(buildConfig)
    configPath := filepath.Join(imagePath, "config.json")
    if err := ioutil.WriteFile(configPath, config, 0644); err != nil {
        fmt.Printf("이미지 메타데이터 저장 실패: %v\n", err)
        return
    }

    // 이미지 이름과 ID를 repositories 파일에 기록
    updateRepositories(imageName, imageID, imagesDir)

    fmt.Printf("이미지 %s가 성공적으로 빌드되었습니다. ID: %s\n", imageName, imageID)
}

// parseCartefile 함수: Cartefile을 읽고 BuildConfig 구조체를 반환하는 함수
func parseCartefile(cartefilePath string) (BuildConfig, error) {
    file, err := os.Open(cartefilePath)
    if err != nil {
        return BuildConfig{}, err
    }
    defer file.Close()

    var config BuildConfig

    // 간단하게 파일에서 FROM과 RUN을 찾아 파싱하는 예시
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "FROM ") {
            config.BaseImage = strings.TrimSpace(strings.TrimPrefix(line, "FROM "))
        } else if strings.HasPrefix(line, "RUN ") {
            config.Commands = append(config.Commands, strings.TrimSpace(strings.TrimPrefix(line, "RUN ")))
        }
    }

    if err := scanner.Err(); err != nil {
        return BuildConfig{}, err
    }

    return config, nil
}

// 베이스 이미지 복사 함수
func copyBaseImageLayer(baseImagePath, imagePath string) error {
    blobsPath := filepath.Join(baseImagePath, "blobs", "sha256")
    
    // 블랍 디렉토리 확인
    blobFiles, err := ioutil.ReadDir(blobsPath)
    if err != nil {
        return fmt.Errorf("블랍 디렉토리 읽기 실패: %v", err)
    }

    // layer 디렉토리 생성
    layerPath := filepath.Join(imagePath, "layer")
    err = os.MkdirAll(layerPath, 0755)
    if err != nil {
        return fmt.Errorf("layer 디렉토리 생성 실패: %v", err)
    }

    // 블랍 파일들을 하나씩 복사
    for _, blobFile := range blobFiles {
        srcFilePath := filepath.Join(blobsPath, blobFile.Name())
        dstFilePath := filepath.Join(layerPath, blobFile.Name()) // 각 블랍 파일을 layer 디렉토리로 복사

        fmt.Printf("블랍 파일 복사 중: %s -> %s\n", srcFilePath, dstFilePath)

        srcFile, err := os.Open(srcFilePath)
        if err != nil {
            return fmt.Errorf("블랍 파일 열기 실패: %v", err)
        }
        defer srcFile.Close()

        dstFile, err := os.Create(dstFilePath)
        if err != nil {
            return fmt.Errorf("블랍 파일 복사 실패: %v", err)
        }
        defer dstFile.Close()

        _, err = io.Copy(dstFile, srcFile)
        if err != nil {
            return fmt.Errorf("블랍 파일 복사 중 오류 발생: %v", err)
        }
    }

    return nil
}

// 이미지 이름과 ID를 repositories 파일에 기록하는 함수
func updateRepositories(imageName, imageID, imagesDir string) {
    reposFile := filepath.Join(imagesDir, "repositories")
    file, err := os.OpenFile(reposFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Printf("repositories 파일 열기 실패: %v\n", err)
        return
    }
    defer file.Close()

    entry := fmt.Sprintf("%s: %s\n", imageName, imageID)
    if _, err := file.WriteString(entry); err != nil {
        fmt.Printf("repositories 파일에 기록 실패: %v\n", err)
    }
}

// 명령어 실행 및 레이어 생성 함수
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

// OCI 메타데이터 생성 함수
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

// 중복되는 디렉토리 생성 함수
func createDirIfNotExists(dir string) {
    if _, err := os.Stat(dir); os.IsNotExist(err) {
        os.MkdirAll(dir, 0755)
    }
}
