package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"uptime-go/internal/api"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var noTimeInLog bool

var (
	enableAPI bool
	apiPort   string
	apiBind   string
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Starts the continuous monitoring process for the configured websites",
	Long: `The 'run' command starts the monitoring service.
It loads websites from the configuration and continuously checks their uptime.

Use this command to start the monitoring service.
Example:
  uptime-go run --config /path/to/your/config.yml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// if len(configuration.Config.Monitor) <= 0 {
		// 	return fmt.Errorf("no valid website configurations found in config file")
		// }

		// if noTimeInLog {
		// 	log.SetFlags(0)
		// }

		// configs := configuration.Config.Monitor

		// var urls []string

		// for _, r := range configs {
		// 	r.ID = helper.GenerateRandomID()
		// 	urls = append(urls, r.URL)
		// }

		// // Initialize database
		// db, err := database.InitializeDatabase()
		// if err != nil {
		// 	fmt.Printf("failed to initialize database: %v", err)
		// 	os.Exit(ExitErrorConnection)
		// }

		// // Merge config
		// db.UpsertRecord(configs, "url", &[]string{
		// 	"url",
		// 	"enabled",
		// 	"response_time_threshold",
		// 	"interval",
		// 	"certificate_monitoring",
		// 	"certificate_expired_before",
		// })
		// db.DB.Where("url IN ?", urls).Find(&configs)

		// // Initialize and start monitor
		// uptimeMonitor, err := monitor.NewUptimeMonitor(db, configs)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error initializing monitor: %v\n", err)
		// 	os.Exit(ExitErrorConfig)
		// }

		// uptimeMonitor.Start()

		// Create a base context that we can cancel when shutting down
		// ctx, cancel := context.WithCancel(context.Background())
		// defer cancel()

		// Monitoring Section
		//

		// API Section
		var apiServer *api.Server
		if enableAPI {
			log.Info().Msg("API server enabled, starting...")

			apiServer = api.NewServer(api.ServerConfig{
				Bind: apiBind,
				Port: apiPort,
			})

			go func() {
				if err := apiServer.Start(); err != nil {
					log.Error().Err(err).Msg("API server failed")
				}
			}()
		}

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

		// Wait for shutdown signal
		<-sigChan
		log.Info().Msg("Shutdown signal received, shutting down...")

		if apiServer != nil {
			apiServer.Shutdown()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&noTimeInLog, "no-time", false, "hide time in log")

	// API flags
	runCmd.Flags().BoolVar(&enableAPI, "api", false, "Enable API server for remote management")
	runCmd.Flags().StringVar(&apiPort, "api-port", "5002", "API server port")
	runCmd.Flags().StringVar(&apiBind, "api-bind", "127.0.0.1", "API server bind address")
}
