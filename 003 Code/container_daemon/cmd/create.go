package cmd

import (
    "archive/tar"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "syscall"
    // "bufio"
    //"strings"
    "golang.org/x/sys/unix"
	
	"github.com/spf13/cobra"
)

var containerName string
// 컨테이너 생성
var createCmd = &cobra.Command{
	Use: "create [imageName]",
	Short: "Container create",
    Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
        imageName := args[0]
        // containerName := imageName
        if containerName == "" {
            containerName = imageName
        }
        return CtCreate(imageName, containerName)
    },
}

func init(){
    createCmd.Flags().StringVarP(&containerName, "output", "o", "", "Container name (optional)")
    rootCmd.AddCommand(createCmd)
}

func CtCreate(imageName string, containerName string) error {
    imageTarPath := "/CarteTest/image/" + imageName + ".tar"
    containerPath := "/CarteTest/container/" + containerName

    fmt.Println("Start Create Container %v", containerName)

    // 이미지 압축 해제
    if _, err := os.Stat(imageTarPath); err == nil {
        fmt.Println("Found tar file...")
        if _, err := os.Stat(containerPath); os.IsNotExist(err) {
            if err := os.MkdirAll(containerPath, 0755); err != nil {
                return fmt.Errorf("failed to create container directory: %v", err)
            }
            if err := extractTar(imageTarPath, containerPath); err != nil {
                return fmt.Errorf("failed to extract tar file: %v", err)
            }
            if err := prepareContainer(containerPath); err != nil {
                return fmt.Errorf("error preparing container: %v", err)
            }

            fmt.Println("Container created successfully")
        } else {
            fmt.Println("Container already prepared")
        }
    } else {
        return fmt.Errorf("image file not found: %s", imageTarPath)
    }

    return nil
}

// tar 파일 압축 해제
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

            // tar 헤더에서 파일 권한 설정
            if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
                return fmt.Errorf("failed to set file permissions: %v", err)
            }
        default:
            log.Printf("Unknown type: %v in %s", header.Typeflag, header.Name)
        }
    }
    return nil
}

func setupSystemDirs() error {
    // /proc 마운트
    if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
        return fmt.Errorf("failed to mount /proc: %v", err)
    }

    // /sys 마운트
    if err := syscall.Mount("sysfs", "/sys", "sysfs", 0, ""); err != nil {
        return fmt.Errorf("failed to mount /sys: %v", err)
    }

    // /dev 마운트
    if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
        return fmt.Errorf("failed to mount /dev: %v", err)
    }

    return nil
}


// 루트 파일 시스템 변경 함수
func prepareContainer(newRoot string) error {
    // 새로운 루트로 변경 (chroot)
    fmt.Printf("checking path: %s\n", newRoot)
    if _, err := os.Stat(newRoot); err != nil {
        return fmt.Errorf("command not found or inaccessible: %v", err) // 오류 메시지 수정 및 오류 반환
    }

    if err := syscall.Chroot(newRoot); err != nil {
        return fmt.Errorf("failed to chroot to new root: %v", err)
    }

    // 루트 디렉토리로 이동
    if err := os.Chdir("/"); err != nil {
        return fmt.Errorf("failed to change directory to new root: %v", err)
    }

    // 시스템 디렉토리 설정
    if err := setupSystemDirs(); err != nil {
        return err
    }

    // /dev/null 및 기타 장치 파일 생성
    if err := createDevFiles(); err != nil {
        return err
    }

    return nil
}




func createDevFiles() error {
    // /dev/null 생성
    if err := syscall.Mknod("/dev/null", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 3))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/null: %v", err)
    }

    // /dev/zero 생성
    if err := syscall.Mknod("/dev/zero", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 5))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/zero: %v", err)
    }

    // /dev/random 생성
    if err := syscall.Mknod("/dev/random", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 8))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/random: %v", err)
    }

    // /dev/urandom 생성
    if err := syscall.Mknod("/dev/urandom", syscall.S_IFCHR|0666, int(unix.Mkdev(1, 9))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/urandom: %v", err)
    }

    // /dev/tty 생성
    if err := syscall.Mknod("/dev/tty", syscall.S_IFCHR|0666, int(unix.Mkdev(5, 0))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/tty: %v", err)
    }

    // /dev/console 생성
    if err := syscall.Mknod("/dev/console", syscall.S_IFCHR|0622, int(unix.Mkdev(5, 1))); err != nil && !os.IsExist(err) {
        return fmt.Errorf("failed to create /dev/console: %v", err)
    }

    return nil
}