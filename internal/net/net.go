package net

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
	"uptime-go/internal/net/config"
	"uptime-go/internal/net/database"
)

type NetworkConfig struct {
	URL             string
	RefreshInterval time.Duration
	Timeout         time.Duration
	FollowRedirects bool
	SkipSSL         bool
}

func (nc *NetworkConfig) CheckWebsite() (*config.CheckResults, error) {
	client := &http.Client{
		Timeout: nc.Timeout,

		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: nc.SkipSSL || isIPAddress(nc.URL)},
		},
	}

	if !nc.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	req, err := http.NewRequest(http.MethodGet, nc.URL, nil)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseTime := time.Since(start)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	isUp := success

	return &config.CheckResults{
		URL:          nc.URL,
		LastCheck:    time.Now(),
		ResponseTime: responseTime,
		IsUp:         isUp,
		StatusCode:   resp.StatusCode,
		ErrorMessage: "",
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

func NotifyHook(db *database.Database, result *config.CheckResults) {
	var payload []byte
	var err error
	url := "http://localhost:8005/uptime"
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	if result.IsUp {
		url += "/up"
		lastUpRecord := db.GetLastUpRecord(result.URL)

		if db.GetLastRecord(result.URL).IsUp {
			fmt.Println("Last record up")
			payload, err = json.Marshal(result)
		} else {
			fmt.Println("Last record down && downtime")
			payload, err = json.Marshal(struct {
				*config.CheckResults
				DownTime string `json:"downtime"`
			}{
				result,
				result.LastCheck.Sub(lastUpRecord.LastCheck).String(),
			})
		}
	} else {
		fmt.Println("down")
		url += "/down"
		payload, err = json.Marshal(result)
	}

	if err != nil {
		log.Printf("Error marshalling JSON for %s: %v", result.URL, err)
		return
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("Error creating request for %s: %v", url, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		// log.Printf("error while doing request to %s: %v", url, err)
		return
	}

	defer response.Body.Close()
}
