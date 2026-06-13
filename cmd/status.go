package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/daemon"
	"github.com/dongt/queryli/internal/ipc"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show daemon and connection status",
	RunE: func(cmd *cobra.Command, args []string) error {
		queryliDir := QueryliDir()

		// Check if daemon is running
		info, err := daemon.ReadPID(daemon.PIDFile(queryliDir))
		if err != nil || !info.Alive {
			daemon.CleanStaleFiles(queryliDir)
			return fmt.Errorf("not connected")
		}

		client := ipc.NewClient(daemon.SocketPath(queryliDir))
		resp, err := client.Send(&ipc.Request{Type: "status"})
		if err != nil {
			daemon.CleanStaleFiles(queryliDir)
			return fmt.Errorf("daemon not responding: %w", err)
		}

		if !resp.OK {
			return fmt.Errorf("daemon error: %s", resp.Error)
		}

		s := resp.Status
		fmt.Printf("Daemon running (PID %d)\n", info.PID)
		fmt.Printf("Profile: %s\n", s.Profile)
		fmt.Printf("Type:    %s\n", s.DBType)
		fmt.Printf("Host:    %s\n", s.Host)
		fmt.Printf("Uptime:  %s\n", s.Uptime)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
