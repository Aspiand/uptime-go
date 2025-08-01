package monitor

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"uptime-go/internal/net"
	"uptime-go/internal/net/config"
	"uptime-go/internal/net/database"
)

// UptimeMonitor represents a service that periodically checks website uptime
type UptimeMonitor struct {
	configs  []*config.Monitor
	db       *database.Database
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewUptimeMonitor(db *database.Database, configs []*config.Monitor) (*UptimeMonitor, error) {
	return &UptimeMonitor{
		configs:  configs,
		db:       db,
		stopChan: make(chan struct{}),
	}, nil
}

func (m *UptimeMonitor) Start() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		m.Stop()
	}()

	fmt.Println("Starting uptime monitoring for", len(m.configs), "websites")
	fmt.Println("Press Ctrl+C to stop")

	// Start a goroutine for each website to monitor
	for _, cfg := range m.configs {
		m.wg.Add(1)
		go m.monitorWebsite(cfg)
	}

	m.wg.Wait()
}

func (m *UptimeMonitor) Stop() {
	close(m.stopChan)
	m.wg.Wait()
	fmt.Println("Monitoring stopped")
}

func (m *UptimeMonitor) monitorWebsite(cfg *config.Monitor) {
	defer m.wg.Done()

	ticker := time.NewTicker(cfg.Interval)
	defer ticker.Stop()

	// Perform initial check immediately
	m.checkWebsite(cfg)

	for {
		select {
		case <-ticker.C:
			m.checkWebsite(cfg)
		case <-m.stopChan:
			return
		}
	}
}

func (m *UptimeMonitor) checkWebsite(cfg *config.Monitor) {
	netConfig := &net.NetworkConfig{
		URL:     cfg.URL,
		Timeout: cfg.ResponseTimeThreshold,
		// FollowRedirects: cfg.FollowRedirects,
		SkipSSL: cfg.SSLMonitoring,
	}

	// result
	_, err := netConfig.CheckWebsite()
	if err != nil {
		log.Printf("Error checking %s: %v", cfg.URL, err)
		// Create a failed check result
		// result = &config.Monitor{
		// 	ID:           database.GenerateRandomID(),
		// 	URL:          cfg.URL,
		// 	LastCheck:    time.Now(),
		// 	ResponseTime: 0,
		// 	IsUp:         false,
		// 	StatusCode:   0,
		// 	ErrorMessage: err.Error(),
		// 	// TODO: add ssl expirate date
		// }

		// if os.IsTimeout(err) {
		// 	result.ResponseTime = int64(cfg.Timeout)
		// }
	}

	// lastRecord :=

	// TODO: hook

	// Log the result
	// statusText := "UP"

	// if !result.IsUp {
	// 	statusText = "DOWN"
	// 	m.db.SaveRecord(&config.Incident{
	// 		ID:              database.GenerateRandomID(),
	// 		URL:             result.URL,
	// 		Body:            result.ErrorMessage,
	// 		StatusCode:      result.StatusCode,
	// 		ResponseTime:    result.ResponseTime,
	// 		SSLExpirateDate: result.SSLExpirateDate,
	// 	})
	// }

	// log.Printf("%s - %s - Response time: %v - Status: %d",
	// 	cfg.URL, statusText, result.ResponseTime, result.StatusCode)

	// Save result to database

	// if err := m.db.UpsertRecord(result); err != nil {
	// 	log.Printf("Failed to save result to database: %v", err)
	// }

	// if err := m.db.SaveRecord(&config.MonitorHistory{
	// 	ID:           database.GenerateRandomID(),
	// 	URL:          result.URL,
	// 	ResponseTime: result.ResponseTime,
	// }); err != nil {
	// 	log.Printf("Failed to save history to database: %v", err)
	// }
}
