package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/daemon"
	"github.com/dongt/queryli/internal/ipc"
	"github.com/dongt/queryli/internal/output"
)

var queryCmd = &cobra.Command{
	Use:   "query <sql>",
	Short: "Run a SQL query against the connected database",
	Long:  `Send a SQL query to the daemon and display results.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(c *cobra.Command, args []string) error {
		return runQuery(args[0])
	},
}

func runQuery(sqlStr string) error {
	queryliDir := QueryliDir()

	// Check daemon
	info, err := daemon.ReadPID(daemon.PIDFile(queryliDir))
	if err != nil || !info.Alive {
		daemon.CleanStaleFiles(queryliDir)
		return fmt.Errorf("not connected")
	}

	client := ipc.NewClient(daemon.SocketPath(queryliDir))
	resp, err := client.Send(&ipc.Request{Type: "query", SQL: sqlStr})
	if err != nil {
		daemon.CleanStaleFiles(queryliDir)
		return fmt.Errorf("query failed: %w", err)
	}

	if !resp.OK {
		return fmt.Errorf("query error: %s", resp.Error)
	}

	formatter := output.NewFormatter(format)
	result := &output.QueryResult{
		Columns: resp.Columns,
		Rows:    resp.Rows,
	}
	fmt.Print(formatter.Format(result))
	return nil
}

func init() {
	rootCmd.AddCommand(queryCmd)
}
