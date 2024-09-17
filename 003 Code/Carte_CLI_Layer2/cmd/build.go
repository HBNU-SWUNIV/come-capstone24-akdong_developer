package cmd

import (
    "bufio"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "os/user"
    "path/filepath"
    "strings"
    "time"

    "github.com/spf13/cobra"
)

// Metadata : 이미지 메타데이터 저장
type Metadata struct {
    Name      string    `json:"name"`
    Timestamp time.Time `json:"timestamp"`
    Layers    []string  `json:"layers"`
    Author    string    `json:"author"`
    ExePath   string    `json:"exe_path"`
    BaseImage string    `json:"base_image"`
}

var buildCmd = &cobra.Command{
    Use:   "build",
    Short: "Build command that handles image compression",
    RunE: func(cmd *cobra.Command, args []string) error {
        return buildImage()
    },
}

// 'build' 명령어 루트 명령어에 추가
func init() {
    rootCmd.AddCommand(buildCmd)
}

func buildImage() error {
    fmt.Println("----------------------- Start Build -----------------------")

    baseDir := "/Carte/images"

    // 베이스 디렉토리 생성
    err := os.MkdirAll(baseDir, 0755)
    if err != nil {
        return fmt.Errorf("failed to create base directory: %v", err)
    }

    // Cartefile.txt 파일 열기
    file, err := os.Open("Cartefile.txt")
    if err != nil {
        return fmt.Errorf("failed to open Cartefile.txt: %v", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)

    var name, srcPath, exePath, baseImage string

    // Cartefile.txt 파일 읽기
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "FROM ") {
            baseImage = strings.TrimSpace(strings.TrimPrefix(line, "FROM "))
        } else if strings.HasPrefix(line, "NAME ") {
            name = strings.TrimSpace(strings.TrimPrefix(line, "NAME "))
        } else if strings.HasPrefix(line, "PATH ") {
            srcPath = strings.TrimSpace(strings.TrimPrefix(line, "PATH "))
        } else if strings.HasPrefix(line, "exe_PATH ") {
            exePath = strings.TrimSpace(strings.TrimPrefix(line, "exe_PATH "))
        }
    }

    if err := scanner.Err(); err != nil {
        return fmt.Errorf("error reading Cartefile.txt: %v", err)
    }

    // 필수 값 확인
    if baseImage == "" {
        return fmt.Errorf("base image is missing in Cartefile.txt")
    }
    if name == "" {
        return fmt.Errorf("name is missing in Cartefile.txt")
    }
    if srcPath == "" {
        return fmt.Errorf("path is missing in Cartefile.txt")
    }
    if exePath == "" {
        return fmt.Errorf("exe_path is missing in Cartefile.txt")
    }

    targetDir := filepath.Join(baseDir, name)
    err = os.MkdirAll(targetDir, 0755)
    if err != nil {
        return fmt.Errorf("failed to create target directory: %v", err)
    }

    layers := []string{}

    // 레이어 1: 소스 코드 복사, exe_PATH는 제외
    layer1Dir := filepath.Join(targetDir, "layer1")
    err = createLayer(srcPath, layer1Dir, exePath)
    if err != nil {
        return fmt.Errorf("failed to create layer1: %v", err)
    }
    layers = append(layers, "layer1")

    // 레이어 2: 빌드된 파일 복사
    layer2Dir := filepath.Join(targetDir, "layer2")
    err = createLayer(exePath, layer2Dir, "")
    if err != nil {
        return fmt.Errorf("failed to create layer2: %v", err)
    }
    layers = append(layers, "layer2")

    // 레이어 3: setup.sh 복사
    layer3Dir := filepath.Join(targetDir, "layer3")
    setupScriptPath := "setup.sh" // setup.sh 파일 경로
    err = createLayer(setupScriptPath, layer3Dir, "")
    if err != nil {
        return fmt.Errorf("failed to create layer3: %v", err)
    }
    layers = append(layers, "layer3")

    // 현재 사용자 이름 가져오기
    usr, err := user.Current()
    if err != nil {
        return fmt.Errorf("failed to get current user: %v", err)
    }

    // 메타데이터 생성
    metadata := Metadata{
        Name:      name,
        Timestamp: time.Now(),
        Layers:    layers,
        Author:    usr.Username,
        ExePath:   exePath,
        BaseImage: baseImage,
    }

    metadataFile := filepath.Join(baseDir, name+"_metadata.json")
    err = saveMetadata(metadataFile, metadata)
    if err != nil {
        return fmt.Errorf("failed to save metadata: %v", err)
    }

    // tar.gz 파일로 압축
    tarFileName := name + ".tar.gz"
    tarFilePath := filepath.Join(baseDir, tarFileName)
    err = createImage(targetDir, tarFilePath)
    if err != nil {
        return fmt.Errorf("failed to create tar.gz file: %v", err)
    }

    fmt.Println("---------------------- Image Build Complete ----------------------")
    return nil
}

// 소스 경로에서 타겟 디렉토리로 레이어를 생성
func createLayer(srcPath, dstDir, excludePath string) error {
    info, err := os.Stat(srcPath)
    if err != nil {
        return err
    }

    if info.IsDir() {
        return copyDirectoryContents(srcPath, dstDir, excludePath)
    }

    // 단일 파일인 경우
    return copySingleFile(srcPath, dstDir)
}

// 디렉토리의 내용 복사
func copyDirectoryContents(srcDir, dstDir, excludePath string) error {
    absExcludePath, err := filepath.Abs(excludePath)
    if err != nil {
        return err
    }

    return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        relPath, err := filepath.Rel(srcDir, path)
        if err != nil {
            return err
        }

        targetPath := filepath.Join(dstDir, relPath)

        absPath, err := filepath.Abs(path)
        if err != nil {
            return err
        }

        if absExcludePath != "" && absPath == absExcludePath {
            return nil
        }

        if info.IsDir() {
            return os.MkdirAll(targetPath, info.Mode())
        }

        return copyFile(path, targetPath)
    })
}

// 단일 파일 복사
func copySingleFile(srcFile, dstDir string) error {
    err := os.MkdirAll(dstDir, 0755)
    if err != nil {
        return err
    }

    dstFile := filepath.Join(dstDir, filepath.Base(srcFile))
    return copyFile(srcFile, dstFile)
}

// 파일 복사
func copyFile(srcFile, dstFile string) error {
    src, err := os.Open(srcFile)
    if err != nil {
        return err
    }
    defer src.Close()

    dst, err := os.Create(dstFile)
    if err != nil {
        return err
    }
    defer dst.Close()

    _, err = io.Copy(dst, src)
    if err != nil {
        return err
    }

    info, err := os.Stat(srcFile)
    if err != nil {
        return err
    }

    return os.Chmod(dstFile, info.Mode())
}

// 디렉토리 tar.gz 파일로 압축
func createImage(srcDir, dstFile string) error {
    args := []string{"-czvf", dstFile, "-C", srcDir, "."}
    cmd := exec.Command("tar", args...)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    return cmd.Run()
}

// 메타데이터 JSON 파일로 저장
func saveMetadata(filePath string, metadata Metadata) error {
    file, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("failed to create metadata file: %v", err)
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    err = encoder.Encode(metadata)
    if err != nil {
        return fmt.Errorf("failed to encode metadata: %v", err)
    }

    return nil
}
