package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/daemon"
	"github.com/dongt/queryli/internal/ipc"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Check if daemon and database connection are alive",
	RunE: func(cmd *cobra.Command, args []string) error {
		queryliDir := QueryliDir()

		// Check PID
		info, err := daemon.ReadPID(daemon.PIDFile(queryliDir))
		if err != nil || !info.Alive {
			daemon.CleanStaleFiles(queryliDir)
			return fmt.Errorf("not connected")
		}

		client := ipc.NewClient(daemon.SocketPath(queryliDir))
		resp, elapsed, err := client.Ping()
		if err != nil {
			daemon.CleanStaleFiles(queryliDir)
			return fmt.Errorf("daemon not responding: %w", err)
		}

		if !resp.OK {
			return fmt.Errorf("ping failed: %s", resp.Error)
		}

		latency := elapsed.Round(100 * 1000) // round to nearest 100µs
		if resp.Status != nil {
			fmt.Printf("Connected to [%s] (%s) — %v\n", resp.Status.Profile, resp.Status.Host, latency)
		} else {
			fmt.Printf("pong — %v\n", latency)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
