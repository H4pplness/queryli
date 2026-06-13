package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/config"
)

var profileAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new connection profile",
	Long: `Add a new database connection profile.

Examples:
  queryli profile add --name prod-pg --type postgres --host db.prod.com --port 5432 --user admin --db myapp
  queryli profile add --name local-sqlite --type sqlite --path ./dev.db
  queryli profile add --name oracle-db --type oracle --host oracle.corp --port 1521 --user app --service ORCL`,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		dbType, _ := cmd.Flags().GetString("type")
		host, _ := cmd.Flags().GetString("host")
		portStr, _ := cmd.Flags().GetString("port")
		user, _ := cmd.Flags().GetString("user")
		dbName, _ := cmd.Flags().GetString("db")
		path, _ := cmd.Flags().GetString("path")
		service, _ := cmd.Flags().GetString("service")
		sslmode, _ := cmd.Flags().GetString("sslmode")

		if name == "" {
			return fmt.Errorf("--name is required")
		}
		if dbType == "" {
			return fmt.Errorf("--type is required")
		}

		port := 0
		if portStr != "" {
			var err error
			port, err = strconv.Atoi(portStr)
			if err != nil {
				return fmt.Errorf("invalid port: %s", portStr)
			}
		}

		cfg, err := config.Load(ConfigPath())
		if err != nil {
			return err
		}

		if _, exists := cfg.Profiles[name]; exists {
			return fmt.Errorf("profile '%s' already exists", name)
		}

		profile := config.Profile{
			Type:    dbType,
			Host:    host,
			Port:    port,
			User:    user,
			DBName:  dbName,
			SSLMode: sslmode,
			Path:    path,
			Service: service,
		}

		// Don't store password in config — user passes it at connect time
		cfg.Profiles[name] = profile

		// If this is the first profile, set as active
		if cfg.ActiveProfile == "" {
			cfg.ActiveProfile = name
		}

		if err := config.Save(cfg, ConfigPath()); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "Profile '%s' (%s) added.\n", name, dbType)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileAddCmd)

	profileAddCmd.Flags().String("name", "", "Profile name (required)")
	profileAddCmd.Flags().String("type", "", "Database type: postgres, mysql, sqlite, oracle (required)")
	profileAddCmd.Flags().String("host", "", "Database host")
	profileAddCmd.Flags().String("port", "", "Database port")
	profileAddCmd.Flags().String("user", "", "Database user")
	profileAddCmd.Flags().String("db", "", "Database name")
	profileAddCmd.Flags().String("path", "", "SQLite file path")
	profileAddCmd.Flags().String("service", "", "Oracle service name or SID")
	profileAddCmd.Flags().String("sslmode", "", "PostgreSQL SSL mode")
}
