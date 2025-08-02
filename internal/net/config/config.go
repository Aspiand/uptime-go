package config

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"uptime-go/internal/net"
)

// TODO:
// - change const to change to string?
// - change SSLExpiredBefore to time.Time?
// - move monitor to net

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
	ID                    string           `json:"id" gorm:"primaryKey"`
	URL                   string           `json:"url" gorm:"unique"`
	Enabled               bool             `json:"enabled"`
	Interval              time.Duration    `json:"-"`              // can be second/minutes/hour (s/m/h)
	ResponseTimeThreshold time.Duration    `json:"-"`              // can be second/minutes (s/m)
	SSLMonitoring         bool             `json:"ssl_monitoring"` // enable ssl monitoring
	SSLExpiredBefore      *time.Duration   `json:"-"`              // can be day/month/year (d/m/y)
	IsUp                  *bool            `json:"is_up"`          // duplicate entry (requested)
	StatusCode            *int             `json:"status_code"`    // duplicate entry (requested)
	CreatedAt             time.Time        `json:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at"`
	Histories             []MonitorHistory `gorm:"foreignKey:MonitorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type MonitorHistory struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	MonitorID    string    `json:"-"`
	IsUp         bool      `json:"is_up" gorm:"index"`
	StatusCode   int       `json:"status_code"`
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

func (m *Monitor) ToNetworkConfig() *net.NetworkConfig {
	return &net.NetworkConfig{
		URL:             m.URL,
		RefreshInterval: m.Interval,
		Timeout:         m.ResponseTimeThreshold,
		// FollowRedirects: true,
		SkipSSL: !m.SSLMonitoring,
	}
}

func generateRandomID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GenerateRandomID() string {
	return generateRandomID(4)
}
