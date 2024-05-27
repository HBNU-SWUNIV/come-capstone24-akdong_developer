package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "carte",
	Short: "Carte is a CLI tool for managing containers",
	Long:  `Carte is a CLI tool that allows users to build and run containers using simple commands.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

