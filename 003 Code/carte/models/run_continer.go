package models

import (
    "archive/tar"
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "syscall"
    "strings"  
)

func RunContainer(imageFilename string) error {
    tempDir, err := os.MkdirTemp("", "container")
    if err != nil {
        return fmt.Errorf("error creating temp directory: %v", err)
    }
    defer os.RemoveAll(tempDir)

    err = extractTarGz(imageFilename, tempDir)
    if err != nil {
        return fmt.Errorf("error extracting image: %v", err)
    }

    err = setupNamespacesAndRun(tempDir)
    if err != nil {
        return fmt.Errorf("error running container: %v", err)
    }

    fmt.Println("Container ran successfully.")
    return nil
}

func extractTarGz(tarGzPath, destDir string) error {
    file, err := os.Open(tarGzPath)
    if err != nil {
        return err
    }
    defer file.Close()

    gzipReader, err := gzip.NewReader(file)
    if err != nil {
        return err
    }
    defer gzipReader.Close()

    tarReader := tar.NewReader(gzipReader)

    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }

        target := filepath.Join(destDir, header.Name)
        if header.Typeflag == tar.TypeDir {
            if err := os.MkdirAll(target, header.FileInfo().Mode()); err != nil {
                return err
            }
        } else {
            fileToWrite, err := os.Create(target)
            if err != nil {
                return err
            }
            defer fileToWrite.Close()

            if _, err := io.Copy(fileToWrite, tarReader); err != nil {
                return err
            }
        }
    }

    return nil
}

func setupNamespacesAndRun(rootDir string) error {
    // PID 네임스페이스 설정
    cmd := exec.Command("/proc/self/exe")
    cmd.Dir = rootDir
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET,
    }
    cmd.Env = append(os.Environ(), "_CONTAINER_INIT=1")
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error starting container process: %v", err)
    }

    return nil
}

func init() {
    if os.Getenv("_CONTAINER_INIT") == "1" {
        if err := setupContainerEnvironment(); err != nil {
            fmt.Fprintf(os.Stderr, "error setting up container environment: %v\n", err)
            os.Exit(1)
        }
    }
}

func setupContainerEnvironment() error {
    if err := cgroups(); err != nil {
        return err
    }

    setupUTSNamespace()
    setupPIDNamespace()
    setupNetworkNamespace()
    setupIPCNamespace()
    setupMountNamespace()

    // 컨테이너 내부에서 명령어 실행
    cartefilePath := "/Cartefile"
    cartefileContent, err := os.ReadFile(cartefilePath)
    if err != nil {
        return fmt.Errorf("error reading Cartefile: %v", err)
    }

    commands := parseCartefile(string(cartefileContent))
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

func setupUTSNamespace() {
    cmd := exec.Command("/bin/hostname", "container1")
    cmd.Run()
}

func setupPIDNamespace() {
    syscall.Sethostname([]byte("container1"))

    cmd := exec.Command("/bin/sh", "-c", "echo 1 > /proc/self/ns/pid")
    cmd.Run()
}

func setupNetworkNamespace() {
    cmd := exec.Command("/bin/sh", "-c", "ip link add veth0 type veth peer name veth1")
    cmd.Run()
}

func setupIPCNamespace() {
    cmd := exec.Command("/bin/sh", "-c", "ipcmk -M 1024")
    cmd.Run()
}

func setupMountNamespace() {
    cmd := exec.Command("/bin/sh", "-c", "mkdir /mnt/containerroot; mount -t tmpfs none /mnt/containerroot")
    cmd.Run()
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


