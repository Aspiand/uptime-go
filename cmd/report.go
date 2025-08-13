package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"uptime-go/internal/net/config"
	"uptime-go/internal/net/database"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var domainURL string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate monitoring report",
	Long: `Generate a JSON report of the monitoring status.

Without a URL flag, it reports all monitored sites.
With a URL flag, it provides a detailed report for the specified site, including the last 100 history records.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.InitializeDatabase()
		if err != nil {
			fmt.Printf("failed to initialize database: %v", err)
			os.Exit(ExitErrorConnection)
		}

		if domainURL == "" {
			var monitor []config.Monitor
			db.DB.Find(&monitor)

			output, err := json.Marshal(monitor)
			if err != nil {
				fmt.Println("Error while serializing output")
			}

			fmt.Println(string(output))
			return
		}

		var monitor config.Monitor
		if err := db.DB.
			Preload("Histories", func(db *gorm.DB) *gorm.DB {
				return db.Order("monitor_histories.created_at DESC").Limit(100)
			}).
			Where("url = ?", domainURL).
			First(&monitor).Error; err != nil {
			fmt.Printf("%s: error while getting record\n", domainURL)
		}

		output, err := json.Marshal(monitor)
		if err != nil {
			fmt.Printf("%s: error while encoding result\n", domainURL)
		}

		fmt.Print(string(output))
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVarP(&domainURL, "url", "u", "", "URL")
}
