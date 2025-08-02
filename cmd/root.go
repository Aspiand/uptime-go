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
	"uptime-go/internal/net/config"
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

	// Get uptime configuration
	// uptimeConfigs, err := configReader.GetUptimeConfig()
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error parsing configuration: %v\n", err)
	// 	os.Exit(ExitErrorConfig)
	// }

	// if len(uptimeConfigs) == 0 {
	// 	fmt.Fprintln(os.Stderr, "No valid website configurations found in config file")
	// 	os.Exit(ExitErrorConfig)
	// }

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

	// Get uptime configuration
	uptimeConfigs, err := configReader.ParseConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing configuration: %v\n", err)
		os.Exit(ExitErrorConfig)
	}

	if len(uptimeConfigs) == 0 {
		fmt.Fprintln(os.Stderr, "No valid website configurations found in config file")
		os.Exit(ExitErrorConfig)
	}

	// Initialize database
	db, err := database.InitializeDatabase()
	if err != nil {
		fmt.Printf("failed to initialize database: %v", err)
		os.Exit(ExitErrorConnection)
	}

	var urls []string

	// Generate a temporary ID for the config.
	// It will be replaced by the database ID if the record exists.
	for _, r := range uptimeConfigs {
		r.ID = config.GenerateRandomID()
		urls = append(urls, r.URL)
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
	rootCmd.Flags().StringVarP(&Config.ConfigFile, "config", "c", "", "Path to configuration file")
}
