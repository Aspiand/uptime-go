package config

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// TODO:
// - change const to change to string?
// - change SSLExpiredBefore to time.Time?

const (
	Timeout = iota
	SSLExpired
	StatusCode
)

type NetworkConfig struct {
	URL             string        // URL to check
	RefreshInterval time.Duration // Interval between checks (for monitoring mode)
	Timeout         time.Duration // HTTP request timeout (second)
	FollowRedirects bool          // Whether to follow HTTP redirects
	SkipSSL         bool          // Whether to skip SSL certificate verification
}

type Monitor struct {
	ID                    string    `json:"id" gorm:"primaryKey"`
	URL                   string    `json:"url" gorm:"unique"`
	Enabled               bool      `json:"enabled"`
	Interval              uint      `json:"-"`              // can be second/minutes/hour (s/m/h)
	SSLMonitoring         bool      `json:"ssl_monitoring"` // enable ssl monitoring
	SSLExpiredBefore      uint      `json:"-"`              // can be day/month/year (d/m/y)
	ResponseTimeThreshold uint      `json:"-"`              // can be second/minutes (s/m)
	IsUp                  *bool     `json:"is_up"`          // duplicate entry (requested)
	StatusCode            *uint     `json:"status_code"`    // duplicate entry (requested)
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type MonitorHistory struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	MonitorID    string    `json:"-"`
	IsUp         bool      `json:"is_up" gorm:"index"`
	StatusCode   uint      `json:"status_code"`
	ResponseTime int64     `json:"response_time"` // in milliseconds
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
}

type Incident struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	MonitorID   string     `json:"monitor_id"`
	Type        uint       `json:"type"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	SolvedAt    *time.Time `json:"solved_at" gorm:"index"`
}

// /etc/ojtguardian/plugins/uptime/config.yml

func generateRandomID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GenerateRandomID() string {
	return generateRandomID(4)
}
