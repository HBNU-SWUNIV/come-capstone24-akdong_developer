package main

import (
    "flag"
    "fmt"
    "time"
    "Carte_Client/client"
)

func main() {
    image := flag.String("image", "", "Container image")
    command := flag.String("command", "/usr/sbin/nginx", "Command to run in container")
    flag.Parse()

    if *image == "" || *command == "" {
        flag.Usage()
        return
    }

    start := time.Now()  // 시작 시간 기록

    if err := client.RunContainer(*image, *command); err != nil {
        fmt.Println("Error:", err)
    } else {
        fmt.Println("Container started successfully")
    }

    elapsed := time.Since(start)  // 경과 시간 계산
    fmt.Printf("Execution time: %s\n", elapsed)
}

