package controller

import (
    "encoding/json"
    "net/http"
    "Carte_Daemon/daemon"
    "Carte_Daemon/utils"
    "log"
)

type RunRequest struct {
    Image   string `json:"image"`
    Command string `json:"command"`
}

func RunContainer(w http.ResponseWriter, r *http.Request) {
    var req RunRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
        log.Printf("Invalid request payload: %v", err)
        return
    }

    if err := daemon.StartContainer(req.Image, req.Command); err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to start container")
        log.Printf("Failed to start container: %v", err)
        return
    }

    utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "started"})
}
