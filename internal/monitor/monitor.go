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

func (m *UptimeMonitor) checkWebsite(monitor *config.Monitor) {
	result, err := monitor.ToNetworkConfig().CheckWebsite()
	if err != nil {
		log.Printf("Error checking %s: %v", monitor.URL, err)
		// TODO: handle
		// Create a failed check result
		result = &net.CheckResults{
			URL:          monitor.URL,
			LastCheck:    time.Now(),
			ResponseTime: 0,
			IsUp:         false,
			StatusCode:   0,
			ErrorMessage: err.Error(),
			// TODO: add ssl expirate date
		}

		// incident := config.Incident{
		// 	ID:          config.GenerateRandomID(),
		// 	MonitorID:   monitor.ID,
		// 	Description: err.Error(),
		// }

		if os.IsTimeout(err) {
			result.ResponseTime = monitor.ResponseTimeThreshold
			// incident.Type = config.Timeout
		}

		// m.db.DB.Create(&incident)
	}

	monitor.IsUp = &result.IsUp
	monitor.StatusCode = &result.StatusCode
	monitor.Histories = []config.MonitorHistory{
		{
			ID:           config.GenerateRandomID(),
			IsUp:         result.IsUp,
			StatusCode:   result.StatusCode,
			ResponseTime: result.ResponseTime.Milliseconds(),
		},
	}
	m.db.UpsertRecord(monitor, "id")

	// TODO: hook

	statusText := "UP"

	if !result.IsUp {
		statusText = "DOWN"
		// m.db.SaveRecord(&config.Incident{
		// 	ID:              database.GenerateRandomID(),
		// 	URL:             result.URL,
		// 	Body:            result.ErrorMessage,
		// 	StatusCode:      result.StatusCode,
		// 	ResponseTime:    result.ResponseTime,
		// 	SSLExpirateDate: result.SSLExpirateDate,
		// })
	}

	// Log the result
	log.Printf("%s - %s - Response time: %v - Status: %d",
		monitor.URL, statusText, result.ResponseTime, result.StatusCode)

	// Save result to database

	// if err := m.db.UpsertRecord(result); err != nil {
	// 	log.Printf("Failed to save result to database: %v", err)
	// }

	// if err := m.db.SaveRecord(&net.MonitorHistory{
	// 	ID:           database.GenerateRandomID(),
	// 	URL:          result.URL,
	// 	ResponseTime: result.ResponseTime,
	// }); err != nil {
	// 	log.Printf("Failed to save history to database: %v", err)
	// }
}
