package controller

import (
    "encoding/json"
    "net/http"

    "Carte/Carte_Daemon/daemon"
    "Carte/Carte_Daemon/utils"
)

type RunRequest struct {
    Image   string `json:"image"`
    Command string `json:"command"`
}

func RunContainer(w http.ResponseWriter, r *http.Request) {
    var req RunRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
        return
    }

    if err := daemon.StartContainer(req.Image, req.Command); err != nil {
        utils.RespondWithError(w, http.StatusInternalServerError, "Failed to start container")
        return
    }

    utils.RespondWithJSON(w, http.StatusOK, map[string]string{"status": "started"})
}
