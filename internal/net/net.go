package net

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"time"
)

type Monitor struct {
	ID                    string           `json:"id" gorm:"primaryKey"`
	URL                   string           `json:"url" gorm:"unique"`
	Enabled               bool             `json:"enabled"`
	Interval              time.Duration    `json:"-"`              // can be second/minutes/hour (s/m/h)
	ResponseTimeThreshold time.Duration    `json:"-"`              // can be second/minutes (s/m)
	SSLMonitoring         bool             `json:"ssl_monitoring"` // enable ssl monitoring
	SSLExpiredBefore      *time.Duration   `json:"-"`              // can be day/month/year (d/m/y)
	IsUp                  *bool            `json:"is_up"`          // duplicate entry (requested)
	StatusCode            *uint            `json:"status_code"`    // duplicate entry (requested)
	CreatedAt             time.Time        `json:"created_at"`
	UpdatedAt             time.Time        `json:"updated_at"`
	Histories             []MonitorHistory `gorm:"foreignKey:MonitorID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
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

type NetworkConfig struct {
	URL             string
	RefreshInterval time.Duration
	Timeout         time.Duration
	FollowRedirects bool
	SkipSSL         bool
}

func (nc *NetworkConfig) CheckWebsite() (*config.Monitor, error) {
	client := &http.Client{
		Timeout: nc.Timeout,

		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: nc.SkipSSL || isIPAddress(nc.URL)},
		},
	}

	// TODO: later
	// if true {
	// 	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
	// 		return http.ErrUseLastResponse
	// 	}
	// }

	req, err := http.NewRequest(http.MethodGet, nc.URL, nil)
	if err != nil {
		return nil, err
	}

	// start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// if tls := resp.TLS; tls != nil {
	// 	fmt.Printf("TLS: %v\n", resp.TLS.PeerCertificates[0].NotAfter.Format(time.RFC1123))
	// }

	// responseTime := time.Since(start).Milliseconds()
	// success := resp.StatusCode >= 200 && resp.StatusCode < 300
	// isUp := success

	return &Monitor{
		// ID:           database.GenerateRandomID(),
		// URL:          nc.URL,
		// LastCheck:    time.Now(),
		// ResponseTime: responseTime,
		// IsUp:         isUp,
		// StatusCode:   resp.StatusCode,
		// ErrorMessage: "",
		// TODO: add ssl expirate date
	}, nil
}

func isIPAddress(host string) bool {
	u, err := url.Parse(host)
	if err != nil {
		return false
	}
	hostname := u.Hostname()

	return net.ParseIP(hostname) != nil
}

// func NotifyHook(db *database.Database, result *config.Monitor) {
// 	var payload []byte
// 	var err error
// 	url := "http://localhost:8005/uptime"
// 	client := &http.Client{
// 		Timeout: 10 * time.Second,
// 	}

// 	if result.IsUp {
// 		url += "/up"
// 		lastUpRecord := db.GetLastUpRecord(result.URL)
// 		lastRecord := db.GetLastRecord(result.URL)

// 		if lastUpRecord.LastCheck.IsZero() ||
// 			lastRecord.LastCheck.IsZero() ||
// 			lastRecord.IsUp {
// 			payload, err = json.Marshal(result)
// 		} else {
// 			payload, err = json.Marshal(struct {
// 				*config.Monitor
// 				DownTime string `json:"downtime"`
// 			}{
// 				result,
// 				result.LastCheck.
// 					Sub(lastUpRecord.LastCheck).
// 					Round(time.Second).String(),
// 			})
// 		}
// 	} else {
// 		url += "/down"
// 		payload, err = json.Marshal(result)
// 	}

// 	if err != nil {
// 		log.Printf("Error marshalling JSON for %s: %v", result.URL, err)
// 		return
// 	}

// 	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
// 	if err != nil {
// 		log.Printf("Error creating request for %s: %v", url, err)
// 		return
// 	}

// 	request.Header.Set("Content-Type", "application/json")

// 	response, err := client.Do(request)
// 	if err != nil {
// 		log.Printf("error while doing request to %s: %v", url, err)
// 		return
// 	}

// 	defer response.Body.Close()
// }
