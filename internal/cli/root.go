package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kudo",
	Short: "Kudo - Lightweight orchestration tool",
	Long:  "Kudo is a lightweight container/process orchestration tool for deploying and managing applications across multiple servers.",
}

func Execute() error {
	return rootCmd.Execute()
}
