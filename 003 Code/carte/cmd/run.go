package cmd

import (
    "carte/models"
    "fmt"
    "github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
    Use:   "run",
    Short: "Run a container from an image",
    Run: func(cmd *cobra.Command, args []string) {
        if len(args) < 1 {
            fmt.Println("Image file is required")
            return
        }

        imageFile := args[0]

        fmt.Printf("Running container from image: %s...\n", imageFile)

        err := models.RunContainer(imageFile)
        if err != nil {
            fmt.Printf("Error running container: %s\n", err)
            return
        }

        fmt.Println("Container ran successfully.")
    },
}

func init() {
    rootCmd.AddCommand(runCmd)
}
