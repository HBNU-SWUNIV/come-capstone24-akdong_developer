package cmd

import (
    "archive/tar"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    //"syscall"
    //"bufio"
    //"strings"
    //"golang.org/x/sys/unix"
    "github.com/spf13/cobra"
)

var containerName string

var createCmd = &cobra.Command{
    Use:   "create [imageName]",
    Short: "Container create",
    Args:  cobra.ExactArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        imageName := args[0]
        if containerName == "" {
            containerName = imageName
        }
        return CtCreate(imageName, containerName)
    },
}

func init() {
    createCmd.Flags().StringVarP(&containerName, "output", "o", "", "Container name (optional)")
    rootCmd.AddCommand(createCmd)
}

func CtCreate(imageName, containerName string) error {
    imageTarPath := "/CarteTest/image/" + imageName + ".tar"
    containerPath := "/CarteTest/container/" + containerName

    fmt.Printf("Start Create Container %v\n", containerName)

    if _, err := os.Stat(imageTarPath); err == nil {
        fmt.Println("Found tar file...")
        if _, err := os.Stat(containerPath); os.IsNotExist(err) {
            if err := os.MkdirAll(containerPath, 0755); err != nil {
                return fmt.Errorf("failed to create container directory: %v", err)
            }
            if err := extractTar(imageTarPath, containerPath); err != nil {
                return fmt.Errorf("failed to extract tar file: %v", err)
            }
            fmt.Println("Container created successfully")
        } else {
            fmt.Println("Container already exists")
        }
    } else {
        return fmt.Errorf("image file not found: %s", imageTarPath)
    }

    return nil
}

func extractTar(tarFile, destDir string) error {
    file, err := os.Open(tarFile)
    if err != nil {
        return fmt.Errorf("failed to open tar file: %v", err)
    }
    defer file.Close()

    tarReader := tar.NewReader(file)
    for {
        header, err := tarReader.Next()
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("failed to read tar file: %v", err)
        }

        targetPath := filepath.Join(destDir, header.Name)

        // 특수 파일을 무시하고 건너뜁니다.
        if header.Typeflag == tar.TypeChar || header.Typeflag == tar.TypeBlock || header.Typeflag == tar.TypeSymlink {
            log.Printf("Skipping special file: %s", header.Name)
            continue
        }

        switch header.Typeflag {
        case tar.TypeDir:
            if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
                return fmt.Errorf("failed to create directory: %v", err)
            }
        case tar.TypeReg:
            outFile, err := os.Create(targetPath)
            if err != nil {
                return fmt.Errorf("failed to create file: %v", err)
            }
            if _, err := io.Copy(outFile, tarReader); err != nil {
                outFile.Close()
                return fmt.Errorf("failed to write file: %v", err)
            }
            outFile.Close()

            if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
                return fmt.Errorf("failed to set file permissions: %v", err)
            }
        default:
            log.Printf("Unknown type: %v in %s", header.Typeflag, header.Name)
        }
    }
    return nil
}
