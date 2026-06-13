package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/daemon"
	"github.com/dongt/queryli/internal/ipc"
)

var disconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from the database (stops daemon)",
	RunE: func(cmd *cobra.Command, args []string) error {
		queryliDir := QueryliDir()

		// Check if daemon is running
		info, err := daemon.ReadPID(daemon.PIDFile(queryliDir))
		if err != nil || !info.Alive {
			// Clean up stale files and report not connected
			daemon.CleanStaleFiles(queryliDir)
			return fmt.Errorf("not connected")
		}

		// Read meta for display
		meta, _ := daemon.LoadMeta(daemon.MetaPath(queryliDir))

		// Send shutdown
		client := ipc.NewClient(daemon.SocketPath(queryliDir))
		resp, err := client.Send(&ipc.Request{Type: "shutdown"})
		if err != nil {
			// Force cleanup if daemon is unresponsive
			fmt.Fprintf(os.Stderr, "Warning: daemon not responding, cleaning up...\n")
			daemon.CleanStaleFiles(queryliDir)
			return fmt.Errorf("daemon did not respond, cleaned up stale files")
		}

		if !resp.OK {
			fmt.Fprintf(os.Stderr, "Daemon responded with error: %s\n", resp.Error)
		}

		// Clean up files
		os.Remove(daemon.PIDFile(queryliDir))
		os.Remove(daemon.SocketPath(queryliDir))
		os.Remove(daemon.MetaPath(queryliDir))

		if meta != nil {
			fmt.Printf("Disconnected from [%s]. Daemon stopped.\n", meta.Profile)
		} else {
			fmt.Println("Disconnected. Daemon stopped.")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(disconnectCmd)
}
