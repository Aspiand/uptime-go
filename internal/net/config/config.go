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

type CheckResults struct {
	URL          string        `json:"url" gorm:"primaryKey"`
	LastCheck    time.Time     `json:"last_check"`
	ResponseTime time.Duration `json:"response_time"`
	IsUp         bool          `json:"-"`
	StatusCode   int           `json:"status_code"`
	ErrorMessage string        `json:"error_message"`
}
