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
		if !cfg.Enabled {
			log.Printf("%s - skipped because disabled\n", cfg.URL)
			continue
		}

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
		// Create a failed check result
		result = &net.CheckResults{
			URL:          monitor.URL,
			LastCheck:    time.Now(),
			ResponseTime: 0,
			IsUp:         false,
			StatusCode:   0,
			ErrorMessage: err.Error(),
		}
	}

	statusText := "UP"

	if err != nil || !result.IsUp {
		statusText = "DOWN"

		incident := config.Incident{
			ID:          config.GenerateRandomID(),
			MonitorID:   monitor.ID,
			Description: err.Error(),
		}

		if !result.IsUp {
			incident.Type = config.UnexpectedStatusCode
		} else if os.IsTimeout(err) {
			result.ResponseTime = monitor.ResponseTimeThreshold
			incident.Type = config.Timeout
		} // TODO: ssl expired

		lastIncident := m.db.GetLastIncident(monitor.URL, incident.Type)
		if lastIncident.CreatedAt.IsZero() {
			log.Printf("%s - New Incident detected! - Status Code: %s", monitor.URL, incident.Type.String()) // TODO: improve message
			m.db.DB.Create(&incident)
		}
	} else {
		now := time.Now()

		// Mark incident with unexpected status code or timeout to solved
		lastIncident := m.db.GetLastIncident(monitor.URL, config.UnexpectedStatusCode)
		if !lastIncident.CreatedAt.IsZero() {
			lastIncident.SolvedAt = &now
			m.db.UpsertRecord(lastIncident, "id")
			log.Printf("%s - Incident Solved - Downtime %s\n", monitor.URL, time.Since(lastIncident.CreatedAt))
		}

		lastSSLIncident := m.db.GetLastIncident(monitor.URL, config.SSLExpired)
		if time.Until(*result.SSLExpiredDate) <= *monitor.SSLExpiredBefore {
			if lastSSLIncident.CreatedAt.IsZero() {
				log.Printf("%s - [%s] - Please update SSL", config.SSLExpired.String(), monitor.URL)
				m.db.DB.Create(&config.Incident{
					ID:          config.GenerateRandomID(),
					MonitorID:   monitor.ID,
					Description: fmt.Sprintf("SSL will be expired on %s", result.SSLExpiredDate),
				})
			}
		} else {
			// if lastSSLIncident exists in database; mark solved
			if !lastSSLIncident.CreatedAt.IsZero() {
				lastSSLIncident.SolvedAt = &now
				m.db.UpsertRecord(lastSSLIncident, "id")
			}
		}
	}

	monitor.UpdatedAt = result.LastCheck
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

	// TODO: hook

	// Log the result
	log.Printf("%s - %s - Response time: %v - Status: %d",
		monitor.URL, statusText, result.ResponseTime, result.StatusCode)

	// Save result to database

	if err := m.db.UpsertRecord(monitor, "id"); err != nil {
		log.Printf("Failed to save result to database: %v", err)
	}
}
