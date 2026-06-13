package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dongt/queryli/internal/config"
)

var profileRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a connection profile",
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

		delete(cfg.Profiles, name)

		if cfg.ActiveProfile == name {
			cfg.ActiveProfile = ""
			// Pick the first remaining profile as active
			for k := range cfg.Profiles {
				cfg.ActiveProfile = k
				break
			}
		}

		if err := config.Save(cfg, ConfigPath()); err != nil {
			return err
		}

		fmt.Printf("Profile '%s' removed.\n", name)
		return nil
	},
}

func init() {
	profileCmd.AddCommand(profileRemoveCmd)
}
