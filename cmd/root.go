/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"uptime-go/internal/configuration"
	"uptime-go/internal/monitor"
	"uptime-go/internal/net/database"

	"github.com/spf13/cobra"
)

type AppConfig struct {
	ConfigFile string
}

var Config AppConfig

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
	if Config.ConfigFile == "" {
		Config.ConfigFile = configuration.ConfigFile
	}

	// Ensure config file is absolute
	if !filepath.IsAbs(Config.ConfigFile) {
		absPath, err := filepath.Abs(Config.ConfigFile)
		if err == nil {
			Config.ConfigFile = absPath
		}
	}

	// Read configuration
	fmt.Printf("Loading configuration from %s\n", Config.ConfigFile)
	configReader := configuration.NewConfigReader()
	if err := configReader.ReadConfig(Config.ConfigFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading configuration: %v\n", err)
		os.Exit(ExitErrorConfig)
	}

	config, err := configReader.ParseConfig()
	if err != nil {
		fmt.Printf("Error while parsing config: %w", err)
	}

	for _, c := range config {

		fmt.Printf("\n--- Website  ---\n")
		fmt.Printf("ID: %s\n", c.ID)
		fmt.Printf("URL: %s\n", c.URL)
		fmt.Printf("Enabled: %t\n", c.Enabled)
		fmt.Printf("Interval: %d\n", c.Interval)
		fmt.Printf("SSL Monitoring: %t\n", c.SSLMonitoring)
		fmt.Printf("SSL Expired Before: %d\n", c.SSLExpiredBefore)
		fmt.Printf("Response Time Threshold: %d\n", c.ResponseTimeThreshold)
		fmt.Printf("Created At: %s\n", c.CreatedAt.Format(time.RFC3339))
		fmt.Printf("Updated At: %s\n", c.UpdatedAt.Format(time.RFC3339))
	}

	return

	// Get uptime configuration
	uptimeConfigs, err := configReader.GetUptimeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing configuration: %v\n", err)
		os.Exit(ExitErrorConfig)
	}

	if len(uptimeConfigs) == 0 {
		fmt.Fprintln(os.Stderr, "No valid website configurations found in config file")
		os.Exit(ExitErrorConfig)
	}

	// TODO: idk
	// // Get domains from agent config
	// domains, err := configReader.GetDomains("/etc/ojtguardian/domains")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error getting domain on agent config: %v\n", err)
	// 	os.Exit(ExitErrorConfig)
	// }

	// if len(domains) == 0 {
	// 	fmt.Fprintln(os.Stderr, "No valid website configurations found in config file")
	// 	os.Exit(ExitErrorConfig)
	// }

	// for _, d := range domains {
	// 	uptimeConfigs = append(uptimeConfigs, &config.NetworkConfig{
	// 		URL:             d,
	// 		RefreshInterval: 1 * time.Minute,
	// 		Timeout:         10 * time.Second,
	// 		FollowRedirects: true,
	// 		SkipSSL:         true,
	// 	})
	// }

	// Initialize database
	db, err := database.InitializeDatabase()
	if err != nil {
		fmt.Errorf("failed to initialize database: %w", err)
		os.Exit(ExitErrorConnection)
	}

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
	rootCmd.Flags().StringVarP(&Config.ConfigFile, "config", "c", "", "Path to configuration file")
}
