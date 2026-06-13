package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/dongt/queryli/internal/config"
)

var (
	cfgFile   string
	password  string
	format    string
	profile   string
)

var rootCmd = &cobra.Command{
	Use:   "queryli",
	Short: "Multi-database CLI query tool",
	Long: `queryli is a CLI tool for managing database connections and running SQL queries.

It supports PostgreSQL, MySQL, SQLite, and Oracle databases through a
daemon-based architecture that keeps a persistent connection alive.`,
	SilenceUsage:  true,
	SilenceErrors: false,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default ~/.queryli/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "", "override active profile")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "database password (if empty, prompts or uses QUERYLI_PASSWORD env)")
	rootCmd.PersistentFlags().StringVarP(&format, "format", "f", "table", "output format (table|json|csv)")

	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error finding home directory:", err)
			os.Exit(1)
		}
		queryliDir := filepath.Join(home, ".queryli")
		viper.AddConfigPath(queryliDir)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")

		// Ensure ~/.queryli/ exists
		os.MkdirAll(queryliDir, 0700)
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintln(os.Stderr, "Error reading config:", err)
		}
		// Config file not found is OK — will be created on profile add
	}
}

// GetConfig loads and returns the config from viper
func GetConfig() (*config.Config, error) {
	return config.Load(viper.ConfigFileUsed())
}

// SaveConfig saves the given config
func SaveConfig(cfg *config.Config) error {
	return config.Save(cfg, viper.ConfigFileUsed())
}

// ConfigPath returns the config file path
func ConfigPath() string {
	if cfgFile != "" {
		return cfgFile
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".queryli", "config.yaml")
}

// QueryliDir returns the ~/.queryli directory
func QueryliDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".queryli")
}

// GetPassword resolves password: flag > env > empty
func GetPassword() string {
	if password != "" {
		return password
	}
	return os.Getenv("QUERYLI_PASSWORD")
}
