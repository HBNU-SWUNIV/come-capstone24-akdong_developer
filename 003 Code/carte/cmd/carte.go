package main

import (
    "fmt"
    "os"
    "carte/pkg/container"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("사용 가능한 명령어: run, list_c, stop, start, remove, list_i, build")
        return
    }

    cmd := os.Args[1]

    switch cmd {
    case "run":
        if len(os.Args) < 5 {
            fmt.Println("사용법: carte run <이름> <이미지> <CPU 제한> <메모리 제한>")
            return
        }
        name := os.Args[2]
        image := os.Args[3]
        cpuLimit := os.Args[4]
        memoryLimit := os.Args[5]
        container.RunContainer(name, image, cpuLimit, memoryLimit)
    case "list_c":
        container.ListContainer()
    case "stop":
        name := os.Args[2]
        container.StopContainer(name)
    case "start":
        name := os.Args[2]
        container.StartContainer(name)
    case "remove":
        name := os.Args[2]
        container.RemoveContainer(name)
    case "list_i":
        container.ListImage()
    case "build":
        cartefilePath := os.Args[2]
        imageName := os.Args[3]
        container.BuildContainer(cartefilePath, imageName)
    default:
        fmt.Println("알 수 없는 명령어:", cmd)
    }
}