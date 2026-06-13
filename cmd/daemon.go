package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/daemon"
)

var daemonProfile string

var daemonCmd = &cobra.Command{
	Use:    "__daemon",
	Short:  "Internal daemon process",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := QueryliDir()

		d, err := daemon.New(configDir, daemonProfile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Daemon error: %v\n", err)
			os.Exit(1)
		}

		if err := d.Run(); err != nil {
			d.Shutdown()
			fmt.Fprintf(os.Stderr, "Daemon error: %v\n", err)
			os.Exit(1)
		}

		d.Shutdown()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().StringVar(&daemonProfile, "profile", "", "Profile name to connect to")
}
