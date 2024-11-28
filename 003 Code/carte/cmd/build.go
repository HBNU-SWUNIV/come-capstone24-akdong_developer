package cmd

import (
    "carte/models"
    "fmt"
    "os"
    "github.com/spf13/cobra"
)

var imageName string

var buildCmd = &cobra.Command{
    Use:   "build",
    Short: "Build a container image",
    Run: func(cmd *cobra.Command, args []string) {
        cartefilePath := "Cartefile"
        if _, err := os.Stat(cartefilePath); os.IsNotExist(err) {
            fmt.Println("Cartefile not found in the current directory")
            return
        }

        workingDir, err := os.Getwd()
        if err != nil {
            fmt.Printf("Error getting current directory: %s\n", err)
            return
        }

        if imageName == "" {
            imageName = "image.tar.gz"
        }

        fmt.Printf("Building container image with name: %s...\n", imageName)

        err = models.BuildImage(imageName, workingDir, cartefilePath)
        if err != nil {
            fmt.Printf("Error building image: %s\n", err)
            return
        }

        fmt.Println("Image built successfully.")
    },
}

func init() {
    rootCmd.AddCommand(buildCmd)
    buildCmd.Flags().StringVarP(&imageName, "name", "n", "", "Name of the output image file (default is 'image.tar.gz')")
}
