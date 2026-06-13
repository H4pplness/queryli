package cmd

import (
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage connection profiles",
	Long:  `Add, list, remove, and set active database connection profiles.`,
}

func init() {
	rootCmd.AddCommand(profileCmd)
}
