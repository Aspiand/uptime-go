package config

import (
	"time"
)

const (
	UP = iota
)

type NetworkConfig struct {
	URL             string        // URL to check
	RefreshInterval time.Duration // Interval between checks (for monitoring mode)
	Timeout         time.Duration // HTTP request timeout (second)
	FollowRedirects bool          // Whether to follow HTTP redirects
	SkipSSL         bool          // Whether to skip SSL certificate verification
}

// TODO: move status_code, error_message/body, ssl_expirate_date to Status struct.

type Monitor struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	URL       string    `json:"url" gorm:"unique"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	// LastCheck time.Time `json:"last_check" gorm:"index"`
}

type MonitorHistory struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	URL          string    `json:"url"`
	IsUp         bool      `json:"is_up" gorm:"index"`
	ResponseTime int64     `json:"response_time"` // milliseconds
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
}

type Incident struct {
	ID              string `gorm:"primaryKey"`
	URL             string `gorm:"index"`
	Body            string
	StatusCode      int
	ResponseTime    int64
	SSLExpirateDate time.Time
	CreatedAt       time.Time `gorm:"index"`
}

// /etc/ojtguardian/plugins/uptime/config.yml
