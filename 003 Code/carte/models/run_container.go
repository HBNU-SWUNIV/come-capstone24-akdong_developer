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
)

// RunContainer runs a container from the specified image file
func RunContainer(imageFile string) error {
    tarFile, err := os.Open(imageFile)
    if err != nil {
        return fmt.Errorf("error opening tar file: %v", err)
    }
    defer tarFile.Close()

    gzipReader, err := gzip.NewReader(tarFile)
    if err != nil {
        return fmt.Errorf("error creating gzip reader: %v", err)
    }
    defer gzipReader.Close()

    tarReader := tar.NewReader(gzipReader)

    // Extract files from tarball
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("error reading tar file: %v", err)
        }

        targetPath := filepath.Join("/tmp/container", header.Name)
        if header.Typeflag == tar.TypeDir {
            if err := os.MkdirAll(targetPath, 0755); err != nil {
                return err
            }
        } else {
            file, err := os.Create(targetPath)
            if err != nil {
                return err
            }
            defer file.Close()

            if _, err := io.Copy(file, tarReader); err != nil {
                return err
            }
        }
    }

    fmt.Println("Container files extracted successfully.")

    // Set environment variables (if any)
    envVars := make(map[string]string)
    envFile := filepath.Join("/tmp/container", ".env")
    if _, err := os.Stat(envFile); err == nil {
        file, err := os.Open(envFile)
        if err != nil {
            return err
        }
        defer file.Close()

        scanner := bufio.NewScanner(file)
        for scanner.Scan() {
            line := scanner.Text()
            parts := strings.SplitN(line, "=", 2)
            if len(parts) == 2 {
                envVars[parts[0]] = parts[1]
            }
        }
    }

    // Run the main application (assuming it's a single binary for simplicity)
    cmdPath := filepath.Join("/tmp/container", "app") // Adjust according to your main binary location
    cmd := exec.Command(cmdPath)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr

    // Set environment variables for the command
    for key, value := range envVars {
        cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
    }

    if err := cmd.Run(); err != nil {
        return fmt.Errorf("error running container: %v", err)
    }

    fmt.Println("Container ran successfully.")
    return nil
}
