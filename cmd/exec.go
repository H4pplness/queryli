package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/daemon"
	"github.com/dongt/queryli/internal/ipc"
)

var execCmd = &cobra.Command{
	Use:   "exec <file.sql>",
	Short: "Execute SQL statements from a file",
	Long:  `Read SQL from a file and send it to the daemon for execution.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]

		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read file: %w", err)
		}

		sqlStr := strings.TrimSpace(string(data))
		if sqlStr == "" {
			return fmt.Errorf("file is empty: %s", filePath)
		}

		queryliDir := QueryliDir()

		// Check daemon
		info, err := daemon.ReadPID(daemon.PIDFile(queryliDir))
		if err != nil || !info.Alive {
			daemon.CleanStaleFiles(queryliDir)
			return fmt.Errorf("not connected")
		}

		client := ipc.NewClient(daemon.SocketPath(queryliDir))
		resp, err := client.Send(&ipc.Request{Type: "exec", SQL: sqlStr})
		if err != nil {
			daemon.CleanStaleFiles(queryliDir)
			return fmt.Errorf("exec failed: %w", err)
		}

		if !resp.OK {
			return fmt.Errorf("exec error: %s", resp.Error)
		}

		if resp.RowsAffected > 0 {
			fmt.Printf("%d row(s) affected\n", resp.RowsAffected)
		} else {
			fmt.Println(resp.Message)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(execCmd)
}
