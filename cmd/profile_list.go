package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/config"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all connection profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(ConfigPath())
		if err != nil {
			return err
		}

		if len(cfg.Profiles) == 0 {
			fmt.Println("No profiles configured. Use 'queryli profile add' to create one.")
			return nil
		}

		for name, p := range cfg.Profiles {
			active := ""
			if name == cfg.ActiveProfile {
				active = " (active)"
			}

			p.Password = maskPassword(p.Password)

			fmt.Printf("[%s]%s\n", name, active)
			fmt.Printf("  Type:     %s\n", p.Type)
			switch p.Type {
			case "sqlite":
				fmt.Printf("  Path:     %s\n", p.Path)
			case "oracle":
				fmt.Printf("  Host:     %s:%d\n", p.Host, p.Port)
				fmt.Printf("  User:     %s\n", p.User)
				fmt.Printf("  Service:  %s\n", p.Service)
			default:
				fmt.Printf("  Host:     %s:%d\n", p.Host, p.Port)
				fmt.Printf("  User:     %s\n", p.User)
				fmt.Printf("  DBName:   %s\n", p.DBName)
				if p.SSLMode != "" {
					fmt.Printf("  SSLMode:  %s\n", p.SSLMode)
				}
			}
			if p.Password != "" {
				fmt.Fprintf(os.Stderr, "  (password stored in config; consider using --password flag instead)\n")
			}
			fmt.Println()
		}
		return nil
	},
}

func maskPassword(pwd string) string {
	if pwd == "" {
		return ""
	}
	if len(pwd) <= 4 {
		return strings.Repeat("*", len(pwd))
	}
	return pwd[:2] + strings.Repeat("*", len(pwd)-4) + pwd[len(pwd)-2:]
}

func init() {
	profileCmd.AddCommand(profileListCmd)
}
