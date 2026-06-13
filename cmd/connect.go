package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/config"
	"github.com/dongt/queryli/internal/daemon"
	"github.com/dongt/queryli/internal/ipc"
)

var connectCmd = &cobra.Command{
	Use:   "connect [profile]",
	Short: "Connect to a database (starts daemon)",
	Long: `Connect to a database profile, starting a background daemon process
that keeps the connection alive for subsequent queries.

If no profile is specified, the active profile from config is used.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		queryliDir := QueryliDir()

		// Resolve profile
		profileName := ""
		if len(args) > 0 {
			profileName = args[0]
		}

		cfg, err := config.Load(ConfigPath())
		if err != nil {
			return err
		}

		profileName, err = config.ResolveProfile(cfg, profileName)
		if err != nil {
			return err
		}

		// Check if daemon is already running
		info, err := daemon.ReadPID(daemon.PIDFile(queryliDir))
		if err == nil && info.Alive {
			// Check if it's our daemon that responds
			client := ipc.NewClient(daemon.SocketPath(queryliDir))
			resp, _, _ := client.Ping()
			if resp != nil && resp.OK {
				if resp.Status != nil {
					return fmt.Errorf("already connected to [%s] (%s). Run 'queryli disconnect' first",
						resp.Status.Profile, resp.Status.Host)
				}
				return fmt.Errorf("already connected. Run 'queryli disconnect' first")
			}
		}

		// Clean up stale files
		daemon.CleanStaleFiles(queryliDir)

		// Get the current executable path
		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("find executable: %w", err)
		}

		// Open log file for daemon
		logFile, err := os.OpenFile(daemon.LogPath(queryliDir), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			return fmt.Errorf("open log: %w", err)
		}

		// Spawn daemon
		daemonCmd := exec.Command(exe, "__daemon", "--profile", profileName)
		daemonCmd.SysProcAttr = daemonSysProcAttr()
		daemonCmd.Stdout = logFile
		daemonCmd.Stderr = logFile
		// Pass through QUERYLI_PASSWORD env if set
		if pwd := GetPassword(); pwd != "" {
			daemonCmd.Env = append(os.Environ(), "QUERYLI_PASSWORD="+pwd)
		}

		if err := daemonCmd.Start(); err != nil {
			logFile.Close()
			return fmt.Errorf("start daemon: %w", err)
		}

		logFile.Close()

		// Wait for daemon to be ready
		client := ipc.NewClient(daemon.SocketPath(queryliDir))
		clientTimeout := 5 * time.Second
		deadline := time.Now().Add(clientTimeout)

		var lastErr error
		for time.Now().Before(deadline) {
			time.Sleep(100 * time.Millisecond)
			resp, _, err := client.Ping()
			if err == nil && resp.OK {
				status := resp.Status
				host := "unknown"
				if status != nil {
					host = status.Host
				}
				fmt.Printf("Daemon started. Connected to [%s] (%s)\n", profileName, host)
				return nil
			}
			lastErr = err
		}

		// If we get here, timeout — daemon may have crashed
		return fmt.Errorf("daemon did not respond within %v: %v", clientTimeout, lastErr)
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
