package config

import (
	"time"
)

type NetworkConfig struct {
	URL             string        // URL to check
	RefreshInterval time.Duration // Interval between checks (for monitoring mode)
	Timeout         time.Duration // HTTP request timeout
	FollowRedirects bool          // Whether to follow HTTP redirects
	SkipSSL         bool          // Whether to skip SSL certificate verification
}

type Monitor struct {
	ID           string    `gorm:"primaryKey"`
	URL          string    `json:"url"`
	LastCheck    time.Time `json:"last_check"`
	ResponseTime int64     `json:"response_time"` // milliseconds
	IsUp         bool      `json:"-"`
	StatusCode   int       `json:"status_code"`
	ErrorMessage string    `json:"error_message"`
}

type MonitorHistory struct {
	ID           string `gorm:"primaryKey"`
	URL          string
	ResponseTime int64 // milliseconds
	CreatedAt    time.Time
}

type Incident struct {
	ID              string `gorm:"primaryKey"`
	URL             string `gorm:"index"`
	Body            string
	StatusCode      *uint
	ResponseTime    *int64
	SSLExpirateDate *time.Time
	CreatedAt       time.Time `gorm:"index"`
}
