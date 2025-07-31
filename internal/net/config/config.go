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

// TODO: move status_code, error_message/body, ssl_expirate_date to Status struct.

type Monitor struct {
	ID              string    `json:"id" gorm:"primaryKey"`
	URL             string    `json:"url"`
	LastCheck       time.Time `json:"last_check" gorm:"index"`
	ResponseTime    int64     `json:"response_time"` // milliseconds
	IsUp            bool      `json:"-"`
	StatusCode      int       `json:"status_code"`
	ErrorMessage    string    `json:"error_message"`
	SSLExpirateDate time.Time `gorm:"-"`
	// Add SSL expirate date, but don't save to table
}

type MonitorHistory struct {
	ID           string `gorm:"primaryKey"`
	URL          string
	ResponseTime int64     // milliseconds
	CreatedAt    time.Time `gorm:"index"`
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
