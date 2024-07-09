package daemon

import (
    "encoding/json"
    "net/http"
    "syscall"
    "os/exec"
)

type RunRequest struct {
    Image   string `json:"image"`
    Command string `json:"command"`
}

// RunContainer handles the /run endpoint request
func RunContainer(w http.ResponseWriter, r *http.Request) {
    var req RunRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    cmd := exec.Command("/proc/self/exe", "child", req.Command)
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWNET,
    }

    if err := cmd.Start(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    if err := cmd.Wait(); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "started"})
}

