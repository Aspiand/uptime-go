/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"uptime-go/internal/configuration"
	"uptime-go/internal/monitor"
	"uptime-go/internal/net/database"

	"github.com/spf13/cobra"
)

// Constants for exit codes
const (
	ExitSuccess          = 0
	ExitErrorInvalidArgs = 1
	ExitErrorConnection  = 2
	ExitErrorConfig      = 3
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "uptime-go",
	Short: "An application to check website uptime",
	Long: `A command-line tool to monitor the uptime of websites.
It provides continuous monitoring of websites defined in the configuration file.

Usage: uptime-go [--config=path/to/uptime.yaml]`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		runMonitorMode()
	},
}

// runMonitorMode reads the configuration file and starts continuous monitoring
func runMonitorMode() {
	// Ensure config file is absolute
	if !filepath.IsAbs(configuration.Config.ConfigFile) {
		absPath, err := filepath.Abs(configuration.Config.ConfigFile)
		if err == nil {
			configuration.Config.ConfigFile = absPath
		}
	}

	// Read configuration
	fmt.Printf("Loading configuration from %s\n", configuration.Config.ConfigFile)
	configReader := configuration.NewConfigReader()
	if err := configReader.ReadConfig(configuration.Config.ConfigFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration: %v\n", err)
		os.Exit(ExitErrorConfig)
	}

	// Get uptime configuration
	uptimeConfigs, err := configReader.ParseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing configuration: %v\n", err)
		os.Exit(ExitErrorConfig)
	}

	var urls []string

	for _, r := range uptimeConfigs {
		urls = append(urls, r.URL)
	}

	// Initialize database
	db, err := database.InitializeDatabase()
	if err != nil {
		fmt.Printf("failed to initialize database: %v", err)
		os.Exit(ExitErrorConnection)
	}

	// Merge config
	db.UpsertRecord(uptimeConfigs, "url")
	db.DB.Where("url IN ?", urls).Find(&uptimeConfigs)

	// Initialize and start monitor
	uptimeMonitor, err := monitor.NewUptimeMonitor(db, uptimeConfigs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing monitor: %v\n", err)
		os.Exit(ExitErrorConfig)
	}

	uptimeMonitor.Start()
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configuration.Config.ConfigFile, "config", "c", "/etc/ojtguardian/plugins/uptime/config.yml", "Path to configuration file")
	rootCmd.PersistentFlags().StringVarP(&configuration.Config.DBFile, "database", "", "/etc/ojtguardian/plugins/uptime/uptime.db", "Path to database file")
}
