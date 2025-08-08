package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"uptime-go/internal/net/config"
	"uptime-go/internal/net/database"

	"github.com/spf13/cobra"
)

var domainName string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "A brief description of your command",
	Long:  ``, // TODO: add later
	Run: func(cmd *cobra.Command, args []string) {
		db, err := database.InitializeDatabase()
		if err != nil {
			fmt.Printf("failed to initialize database: %v", err)
			os.Exit(ExitErrorConnection)
		}

		if domainName == "" {
			var monitor []config.Monitor
			db.DB.Find(&monitor)

			output, err := json.Marshal(struct {
				Results []config.Monitor `json:"results"`
			}{
				Results: monitor,
			})
			if err != nil {
				fmt.Println("Error while serializing output")
			}

			fmt.Println(string(output))
			return
		}

		var histories []config.MonitorHistory
		db.DB.Joins("JOIN monitors ON monitors.id = monitor_histories.monitor_id").
			Where("monitors.url = ?", domainName).
			Order("monitor_histories.created_at DESC").
			Limit(100).
			Find(&histories)

		output, err := json.Marshal(struct {
			Histories []config.MonitorHistory `json:"histories"`
		}{
			Histories: histories,
		})
		if err != nil {
			fmt.Println("Error while serializing output")
		}

		fmt.Println(string(output))
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().StringVarP(&domainName, "url", "u", "", "URL")
}
