package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/config"
)

var profileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set a profile as the active default",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg, err := config.Load(ConfigPath())
		if err != nil {
			return err
		}

		if _, exists := cfg.Profiles[name]; !exists {
			return fmt.Errorf("profile '%s' not found", name)
		}

		cfg.ActiveProfile = name

		if err := config.Save(cfg, ConfigPath()); err != nil {
			return err
		}

		fmt.Printf("Active profile set to '%s'.\n", name)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileUseCmd)
}
