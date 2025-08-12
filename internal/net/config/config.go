package config

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"uptime-go/internal/net"
)

type ErrorType int

const (
	UnexpectedStatusCode ErrorType = iota
	SSLExpired
	Timeout
)

type Monitor struct {
	ID                       string           `json:"-" gorm:"primaryKey"`
	URL                      string           `json:"url" gorm:"unique"`
	Enabled                  bool             `json:"-"`
	Interval                 time.Duration    `json:"-"`
	ResponseTimeThreshold    time.Duration    `json:"-"`
	CertificateMonitoring    bool             `json:"-"` // enable ssl monitoring
	CertificateExpiredBefore *time.Duration   `json:"-"`
	IsUp                     *bool            `json:"is_up"` // remove and use last_up instead?; mls
	StatusCode               *int             `json:"status_code"`
	ResponseTime             *int64           `json:"response_time"`
	CertificateExpiredDate   *time.Time       `json:"certificate_expired_date"`
	LastUp                   *time.Time       `json:"last_up"`
	CreatedAt                time.Time        `json:"-"`
	UpdatedAt                time.Time        `json:"last_check"`
	Histories                []MonitorHistory `json:"-" gorm:"foreignKey:MonitorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Incidents                []Incident       `json:"-" gorm:"foreignKey:MonitorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

type MonitorHistory struct {
	ID           string    `json:"-" gorm:"primaryKey"`
	MonitorID    string    `json:"-"`
	IsUp         bool      `json:"is_up" gorm:"index"`
	StatusCode   int       `json:"-"`
	ResponseTime int64     `json:"response_time"` // in milliseconds
	CreatedAt    time.Time `json:"created_at" gorm:"index"`
	Monitor      Monitor   `json:"-" gorm:"foreignKey:MonitorID"`
}

type Incident struct {
	ID          string     `json:"id" gorm:"primaryKey"`
	MonitorID   string     `json:"monitor_id"`
	Type        ErrorType  `json:"type" gorm:"index"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	SolvedAt    *time.Time `json:"solved_at" gorm:"index"`
	Monitor     Monitor    `gorm:"foreignKey:MonitorID"`
}

func (e ErrorType) String() string {
	switch e {
	case Timeout:
		return "Timeout occurred"
	case SSLExpired:
		return "SSL certificate expired"
	case UnexpectedStatusCode:
		return "Unexpected status code"
	default:
		return "Unknown error"
	}
}

func (m *Monitor) ToNetworkConfig() *net.NetworkConfig {
	return &net.NetworkConfig{
		URL:             m.URL,
		RefreshInterval: m.Interval,
		Timeout:         m.ResponseTimeThreshold,
		// FollowRedirects: true,
		SkipSSL: !m.CertificateMonitoring,
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
