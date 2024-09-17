package models

import (
    "archive/tar"
    "bufio"
    "compress/gzip"
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "syscall"
)

// Layer represents a filesystem layer in the image
type Layer struct {
    ID   string
    Path string
}

// ImageConfig holds configuration for the image
type ImageConfig struct {
    Workdir      string
    Entrypoint   []string
    Cmd          []string
    ExposedPorts []string
    EnvVars      []string
    BaseImage    string
}

// BuildImage builds a container image from the specified source directory
func BuildImage(outputFilename, sourceDir, cartefilePath string) error {
    if err := runInitSetup(); err != nil {
        return fmt.Errorf("error during initial setup: %v", err)
    }

    layers, config, err := createLayers(sourceDir, cartefilePath)
    if err != nil {
        return err
    }

    return createImageTarball(outputFilename, layers, config)
}

// runInitSetup runs the initial setup script to configure necessary permissions
func runInitSetup() error {
    cmd := exec.Command("sh", "-c", "./init_setup.sh")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error running initial setup script: %v", err)
    }

    return nil
}

// createLayers creates layers from the source directory based on Cartefile instructions
func createLayers(sourceDir, cartefilePath string) ([]Layer, ImageConfig, error) {
    var layers []Layer
    var config ImageConfig
    envVars := make(map[string]string)
    workdir := ""
    ignorePatterns, _ := readCarteignore(filepath.Join(sourceDir, ".carteignore"))

    // Read instructions from Cartefile
    instructions, err := readCartefile(cartefilePath)
    if err != nil {
        return nil, config, err
    }

    for i, instruction := range instructions {
        layerID := fmt.Sprintf("layer%d", i+1)
        layerPath := filepath.Join("/tmp", layerID) // Use a temporary directory for each layer

        if err := os.MkdirAll(layerPath, 0755); err != nil {
            return nil, config, err
        }

        fmt.Printf("Processing instruction: %s\n", instruction) // Debug message

        // Apply instruction to create a new layer
        if strings.HasPrefix(instruction, "FROM") {
            // Handle FROM instruction
            parts := strings.Fields(instruction)
            if len(parts) == 2 {
                config.BaseImage = parts[1]
            }
        } else if strings.HasPrefix(instruction, "WORKDIR") {
            // Handle WORKDIR instruction
            parts := strings.SplitN(instruction, " ", 2)
            if len(parts) == 2 {
                workdir = parts[1]
                config.Workdir = workdir
                fullPath := filepath.Join(layerPath, workdir)
                if err := os.MkdirAll(fullPath, 0755); err != nil {
                    return nil, config, err
                }
                fmt.Printf("Created WORKDIR: %s\n", fullPath) // Debug message
            }
        } else if strings.HasPrefix(instruction, "COPY") {
            // Handle COPY instruction
            parts := strings.Split(instruction, " ")
            src := parts[1]
            dst := filepath.Join(workdir, parts[2])

            dstPath := filepath.Join(layerPath, dst)
            if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
                return nil, config, err
            }

            if err := copyDir(filepath.Join(sourceDir, src), dstPath, ignorePatterns); err != nil {
                return nil, config, err
            }
        } else if strings.HasPrefix(instruction, "RUN") {
            // Handle RUN instruction (execute command)
            if err := runInNamespace(layerPath, workdir, instruction[4:]); err != nil {
                return nil, config, err
            }
        } else if strings.HasPrefix(instruction, "ENV") {
            // Handle ENV instruction
            parts := strings.SplitN(instruction, " ", 2)
            if len(parts) == 2 {
                envParts := strings.SplitN(parts[1], "=", 2)
                if len(envParts) == 2 {
                    envVars[envParts[0]] = envParts[1]
                    config.EnvVars = append(config.EnvVars, fmt.Sprintf("%s=%s", envParts[0], envParts[1]))
                }
            }
        } else if strings.HasPrefix(instruction, "ENTRYPOINT") {
            // Handle ENTRYPOINT instruction
            parts := strings.Fields(instruction)
            config.Entrypoint = parts[1:]
        } else if strings.HasPrefix(instruction, "CMD") {
            // Handle CMD instruction
            parts := strings.Fields(instruction)
            config.Cmd = parts[1:]
        } else if strings.HasPrefix(instruction, "EXPOSE") {
            // Handle EXPOSE instruction
            parts := strings.Fields(instruction)
            config.ExposedPorts = append(config.ExposedPorts, parts[1:]...)
        }

        layers = append(layers, Layer{
            ID:   layerID,
            Path: layerPath,
        })
    }

    return layers, config, nil
}

// readCartefile reads the instructions from a Cartefile
func readCartefile(cartefilePath string) ([]string, error) {
    file, err := os.Open(cartefilePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var instructions []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        instructions = append(instructions, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return instructions, nil
}

// readCarteignore reads the patterns from a .carteignore file
func readCarteignore(carteignorePath string) ([]string, error) {
    file, err := os.Open(carteignorePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var patterns []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        patterns = append(patterns, scanner.Text())
    }

    if err := scanner.Err(); err != nil {
        return nil, err
    }

    return patterns, nil
}

// createImageTarball creates a tarball of the image layers and configuration
func createImageTarball(outputFilename string, layers []Layer, config ImageConfig) error {
    tarFile, err := os.Create(outputFilename)
    if err != nil {
        return fmt.Errorf("error creating tar file: %v", err)
    }
    defer tarFile.Close()

    gzipWriter := gzip.NewWriter(tarFile)
    defer gzipWriter.Close()

    tarWriter := tar.NewWriter(gzipWriter)
    defer tarWriter.Close()

    for _, layer := range layers {
        err = filepath.Walk(layer.Path, func(file string, fi os.FileInfo, err error) error {
            if err != nil {
                return err
            }

            header, err := tar.FileInfoHeader(fi, fi.Name())
            if err != nil {
                return err
            }

            header.Name, err = filepath.Rel(layer.Path, file)
            if err != nil {
                return err
            }

            header.Name = filepath.Join(layer.ID, header.Name)

            if err := tarWriter.WriteHeader(header); err != nil {
                return err
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
                return err
            }

            return nil
        })

        if err != nil {
            return fmt.Errorf("error adding files to tar: %v", err)
        }
    }

    // Add image configuration to the tarball
    configData := fmt.Sprintf(`
    {
        "Workdir": "%s",
        "Entrypoint": "%s",
        "Cmd": "%s",
        "ExposedPorts": "%s",
        "EnvVars": "%s",
        "BaseImage": "%s"
    }`,
        config.Workdir,
        strings.Join(config.Entrypoint, " "),
        strings.Join(config.Cmd, " "),
        strings.Join(config.ExposedPorts, " "),
        strings.Join(config.EnvVars, " "),
        config.BaseImage,
    )

    configHeader := &tar.Header{
        Name: "config.json",
        Mode: 0600,
        Size: int64(len(configData)),
    }

    if err := tarWriter.WriteHeader(configHeader); err != nil {
        return err
    }

    if _, err := tarWriter.Write([]byte(configData)); err != nil {
        return err
    }

    
    fmt.Println("Image tarball created successfully.")
    return nil
}

// copyDir copies a directory from src to dst
func copyDir(src, dst string, ignorePatterns []string) error {
    return filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
        if err != nil {
            return err
        }

        if shouldIgnore(file, ignorePatterns) {
            if fi.IsDir() {
                return filepath.SkipDir
            }
            return nil
        }

        relPath, err := filepath.Rel(src, file)
        if err != nil {
            return err
        }

        dstPath := filepath.Join(dst, relPath)

        if fi.IsDir() {
            return os.MkdirAll(dstPath, fi.Mode())
        }

        srcFile, err := os.Open(file)
        if err != nil {
            return err
        }
        defer srcFile.Close()

        dstFile, err := os.Create(dstPath)
        if err != nil {
            return err
        }
        defer dstFile.Close()

        if _, err := io.Copy(dstFile, srcFile); err != nil {
            return err
        }

        return nil
    })
}

// shouldIgnore checks if a file should be ignored based on .carteignore patterns
func shouldIgnore(file string, patterns []string) bool {
    for _, pattern := range patterns {
        match, err := filepath.Match(pattern, file)
        if err != nil {
            continue
        }
        if match {
            return true
        }
    }
    return false
}

// runInNamespace runs a command in a new namespace using pivot_root and cgroups
func runInNamespace(layerPath, workdir, command string) error {
    cmd := exec.Command("sh", "-c", command)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET,
    }
    cmd.Dir = workdir
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    cmd.Env = append(os.Environ(), "PATH=/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin")

    // Create a new mount namespace and set up pivot_root
    if err := syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
        return fmt.Errorf("error creating new mount namespace: %v", err)
    }
    if err := syscall.Mount(layerPath, layerPath, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
        return fmt.Errorf("error binding mount layerPath: %v", err)
    }

    putOld := filepath.Join(layerPath, "put_old")
    if err := os.Mkdir(putOld, 0755); err != nil {
        return fmt.Errorf("error creating put_old directory: %v", err)
    }
    if err := syscall.PivotRoot(layerPath, putOld); err != nil {
        return fmt.Errorf("error during pivot_root: %v", err)
    }
    if err := os.Chdir("/"); err != nil {
        return fmt.Errorf("error changing directory to new root: %v", err)
    }
    if err := syscall.Unmount("/put_old", syscall.MNT_DETACH); err != nil {
        return fmt.Errorf("error unmounting put_old: %v", err)
    }
    if err := os.Remove("/put_old"); err != nil {
        return fmt.Errorf("error removing put_old directory: %v", err)
    }

    // Create and enter a new cgroup
    cgroupPath := "/sys/fs/cgroup/my_cgroup"
    if err := os.MkdirAll(cgroupPath, 0755); err != nil {
        return fmt.Errorf("error creating cgroup: %v", err)
    }
    defer os.RemoveAll(cgroupPath)

    // Add the current process to the cgroup
    if err := os.WriteFile(filepath.Join(cgroupPath, "tasks"), []byte(fmt.Sprintf("%d", os.Getpid())), 0644); err != nil {
        return fmt.Errorf("error adding process to cgroup: %v", err)
    }

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error running command in namespace: %v", err)
    }

    return nil
}
