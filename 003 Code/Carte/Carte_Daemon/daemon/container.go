package daemon

import (
    "log"
    "os"
    "os/exec"
)

func InitContainer(command string) {
    cmd := exec.Command("sh", "-c", command)
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Run(); err != nil {
        log.Fatal(err)
    }
}
