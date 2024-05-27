package models

import (
    "archive/tar"
    "compress/gzip"
    "encoding/json"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
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

func BuildImage(outputFilename string, sourceDir string) error {
    tempDir, err := os.MkdirTemp("", "image_layers")
    if err != nil {
        return fmt.Errorf("error creating temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    cartefilePath := filepath.Join(sourceDir, "Cartefile")
    cartefileContent, err := os.ReadFile(cartefilePath)
    if err != nil {
        return fmt.Errorf("error reading Cartefile: %v", err)
    }
    commands := parseCartefile(string(cartefileContent))

    var layers []string
    for i, command := range commands {
        layerDir := filepath.Join(tempDir, fmt.Sprintf("layer%d", i))
        if err := os.Mkdir(layerDir, 0755); err != nil {
            return fmt.Errorf("error creating layer directory: %v", err)
        }

        // Translate Cartefile command to shell command
        shellCommand := translateCartefileCommand(command)
        if err := runCommandInUserNamespace(shellCommand, layerDir, sourceDir); err != nil {
            return fmt.Errorf("error running command '%s': %v", shellCommand, err)
        }

        layerTar := filepath.Join(tempDir, fmt.Sprintf("layer%d.tar", i))
        if err := createTar(layerDir, layerTar); err != nil {
            return fmt.Errorf("error creating tar for layer %d: %v", i, err)
        }

        layers = append(layers, fmt.Sprintf("layer%d.tar", i))
    }

    manifest := createManifest(layers)
    config := createConfig(commands)

    // Create final tar.gz file
    tarFile, err := os.Create(outputFilename)
    if err != nil {
        return fmt.Errorf("error creating tar file: %v", err)
    }
    defer tarFile.Close()

    gzipWriter := gzip.NewWriter(tarFile)
    defer gzipWriter.Close()

    tarWriter := tar.NewWriter(gzipWriter)
    defer tarWriter.Close()

    // Add manifest.json to tar
    manifestBytes, err := json.Marshal(manifest)
    if err != nil {
        return fmt.Errorf("error marshalling manifest: %v", err)
    }
    if err := addFileToTar(tarWriter, "manifest.json", manifestBytes); err != nil {
        return fmt.Errorf("error adding manifest to tar: %v", err)
    }

    // Add config.json to tar
    configBytes, err := json.Marshal(config)
    if err != nil {
        return fmt.Errorf("error marshalling config: %v", err)
    }
    if err := addFileToTar(tarWriter, "config.json", configBytes); err != nil {
        return fmt.Errorf("error adding config to tar: %v", err)
    }

    // Add each layer tar to the final tar.gz file
    for _, layer := range layers {
        layerPath := filepath.Join(tempDir, layer)
        layerBytes, err := os.ReadFile(layerPath)
        if err != nil {
            return fmt.Errorf("error reading layer file: %v", err)
        }
        if err := addFileToTar(tarWriter, layer, layerBytes); err != nil {
            return fmt.Errorf("error adding layer to tar: %v", err)
        }
    }

    fmt.Println("Image tarball created successfully.")
    return nil
}

func translateCartefileCommand(command string) string {
    parts := strings.Fields(command)
    switch parts[0] {
    case "RUN":
        return strings.Join(parts[1:], " ")
    case "COPY":
        if len(parts) == 3 {
            return fmt.Sprintf("cp -r %s %s", parts[1], parts[2])
    }
    case "WORKDIR":
        return fmt.Sprintf("mkdir -p %s && cd %s", parts[1], parts[1])
    case "CMD":
        // CMD should be handled in the config, not as a command to run during build
        return ""
    default:
        return command
    }
    return ""
}

func runCommandInUserNamespace(command, layerDir, sourceDir string) error {
    if command == "" {
        return nil
    }

    cmd := exec.Command("unshare", "--user", "--map-root-user", "--mount-proc", "--pid", "--fork", "/bin/sh", "-c", command)
    cmd.Dir = sourceDir

    // Create a temporary file to capture the command output
    outputFile, err := os.CreateTemp("", "cmd_output")
    if err != nil {
        return fmt.Errorf("error creating output file: %v", err)
    }
    defer outputFile.Close()
    defer os.Remove(outputFile.Name())

    cmd.Stdout = outputFile
    cmd.Stderr = outputFile

    if err := cmd.Run(); err != nil {
        // Read the output from the file
        output, readErr := os.ReadFile(outputFile.Name())
        if readErr != nil {
            return fmt.Errorf("error reading command output: %v", readErr)
        }
        return fmt.Errorf("command output: %s, error: %v", string(output), err)
    }

    return copyDir(sourceDir, layerDir)
}

func copyDir(src string, dst string) error {
    return filepath.Walk(src, func(srcPath string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        relPath, err := filepath.Rel(src, srcPath)
        if err != nil {
            return err
        }
        dstPath := filepath.Join(dst, relPath)
        if info.IsDir() {
            return os.MkdirAll(dstPath, info.Mode())
        }
        return copyFile(srcPath, dstPath)
    })
}

func copyFile(src, dst string) error {
    sourceFile, err := os.Open(src)
    if err != nil {
        return err
    }
    defer sourceFile.Close()

    destinationFile, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer destinationFile.Close()

    _, err = io.Copy(destinationFile, sourceFile)
    return err
}

func createTar(sourceDir, tarPath string) error {
    tarFile, err := os.Create(tarPath)
    if err != nil {
        return fmt.Errorf("error creating tar file: %v", err)
    }
    defer tarFile.Close()

    tarWriter := tar.NewWriter(tarFile)
    defer tarWriter.Close()

    err = filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        header, err := tar.FileInfoHeader(fi, fi.Name())
        if err != nil {
            return err
        }

        header.Name, err = filepath.Rel(filepath.Dir(sourceDir), file)
        if err != nil {
            return err
        }

        if err := tarWriter.WriteHeader(header); err != nil {
            return fmt.Errorf("error writing tar header: %v", err)
        }

        if fi.IsDir() {
            return nil
        }

        f, err := os.Open(file)
        if err != nil {
            return err
        }
        defer f.Close()

        if _, err := io.Copy(tarWriter, f); err != nil {
            return fmt.Errorf("error writing file to tar: %v", err)
        }

        return nil
    })

    if err != nil {
        return fmt.Errorf("error adding directory to tar: %v", err)
    }

    return nil
}

func addFileToTar(tarWriter *tar.Writer, name string, content []byte) error {
    header := &tar.Header{
        Name: name,
        Size: int64(len(content)),
        Mode: 0600,
    }
    if err := tarWriter.WriteHeader(header); err != nil {
        return fmt.Errorf("error writing tar header: %v", err)
    }
    if _, err := tarWriter.Write(content); err != nil {
        return fmt.Errorf("error writing tar content: %v", err)
    }
    return nil
}

func parseCartefile(cartefileContent string) []string {
    lines := strings.Split(cartefileContent, "\n")
    var commands []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        commands = append(commands, line)
    }
    return commands
}

func createManifest(layers []string) Manifest {
    return Manifest{
        Config:   "config.json",
        RepoTags: []string{"myimage:latest"},
        Layers:   layers,
    }
}

func createConfig(commands []string) Config {
    return Config{
        Architecture: "amd64",
        Created:      time.Now(),
        OS:           "linux",
        Config: struct {
            Env        []string `json:"Env"`
            Entrypoint []string `json:"Entrypoint"`
            Cmd        []string `json:"Cmd"`
        }{
            Env:        []string{"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"},
            Entrypoint: []string{"/bin/sh"},
            Cmd:        commands,
        },
        RootFS: struct {
            Type    string   `json:"type"`
            DiffIDs []string `json:"diff_ids"`
        }{
            Type:    "layers",
            DiffIDs: []string{"sha256:<layer-sha256-hash>"},
        },
    }
}
