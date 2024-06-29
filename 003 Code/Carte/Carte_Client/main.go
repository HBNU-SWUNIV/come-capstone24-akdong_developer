package main

import (
    "flag"
    "fmt"

    "Carte/Carte_Client/client"
)

func main() {
    image := flag.String("image", "", "Container image")
    command := flag.String("command", "", "Command to run in container")
    flag.Parse()

    if *image == "" || *command == "" {
        flag.Usage()
        return
    }

    if err := client.RunContainer(*image, *command); err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Container started successfully")
    }
}
