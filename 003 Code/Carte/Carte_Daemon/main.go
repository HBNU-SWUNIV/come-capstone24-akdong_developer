package main

import (
    "log"
    "net/http"

    "Carte/Carte_Daemon/controller"
    "Carte/Carte_Daemon/middleware"
)

func main() {
    mux := http.NewServeMux()
    mux.HandleFunc("/run", controller.RunContainer)

    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", middleware.Logging(mux)))
}

