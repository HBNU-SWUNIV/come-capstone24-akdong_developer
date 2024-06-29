package client

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type RunRequest struct {
    Image   string `json:"image"`
    Command string `json:"command"`
}

func RunContainer(image, command string) error {
    req := RunRequest{Image: image, Command: command}
    body, err := json.Marshal(req)
    if err != nil {
        return err
    }

    resp, err := http.Post("http://localhost:8080/run", "application/json", bytes.NewBuffer(body))
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to start container: %s", resp.Status)
    }

    return nil
}
