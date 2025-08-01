package config

import (
	"crypto/rand"
	"encoding/hex"
	"time"
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

// /etc/ojtguardian/plugins/uptime/config.yml

func generateRandomID(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GenerateRandomID() string {
	return generateRandomID(4)
}
