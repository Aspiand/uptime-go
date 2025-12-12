package config

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"sync"
	"time"
	"uptime-go/internal/helper"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

var (
	ErrInvalidConfig = errors.New("invalid configuration")
	durationRegex    = regexp.MustCompile(`(\d+)([smhd])`)
)

type Config struct {
	Monitors      []MonitorConfig
	Agent         AgentConfig
	onChange      []func(fsnotify.Event)
	v             *viper.Viper
	mu            *sync.RWMutex
	debounceTimer *time.Timer
}

type MonitorConfig struct {
	URL                      string        `mapstructure:"url" yaml:"url" json:"url"`
	Enabled                  bool          `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Interval                 time.Duration `mapstructure:"interval" yaml:"interval" json:"interval"`
	ResponseTimeThreshold    time.Duration `mapstructure:"response_time_threshold" yaml:"response_time_threshold" json:"response_time_threshold"`
	CertificateMonitoring    bool          `mapstructure:"certificate_monitoring" yaml:"certificate_monitoring" json:"certificate_monitoring"`
	CertificateExpiredBefore time.Duration `mapstructure:"certificate_expired_before" yaml:"certificate_expired_before" json:"certificate_expired_before"`
}

type AgentConfig struct {
	MasterHost string `yaml:"master_host" mapstructure:"master_host"`
	Auth       struct {
		Token string
	}
}

func New() *Config {
	v := viper.New()

	return &Config{
		v:        v,
		onChange: make([]func(fsnotify.Event), 0),
		mu:       &sync.RWMutex{},
	}
}

func (c *Config) OnConfigChange(run func(in fsnotify.Event)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onChange = append(c.onChange, run)
}

func (c *Config) Load(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	c.v.SetConfigType("yaml")
	c.v.SetConfigFile(path)

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if err := c.v.WriteConfig(); err != nil {
				return err
			}
		}

		return err
	} else {
		if err := c.v.ReadInConfig(); err != nil {
			return nil
		}
	}

	if err := c.Parse(); err != nil {
		return err
	}

	c.v.OnConfigChange(func(in fsnotify.Event) {
		c.mu.Lock()
		defer c.mu.Unlock()

		if c.debounceTimer != nil {
			c.debounceTimer.Stop()
		}

		c.debounceTimer = time.AfterFunc(500*time.Millisecond, func() {
			c.reload(in)
		})
	})

	c.v.WatchConfig()

	return nil
}

func (c *Config) reload(in fsnotify.Event) {
	startTime := time.Now()
	log.Info().Msg("config file changed, reloading...")

	if err := c.Parse(); err != nil {
		log.Err(err).Msg("failed to parse config")
		return
	}

	// Create a copy of the slice to avoid holding the lock while executing potentially long-running callbacks.
	// This prevents other goroutines from being blocked if a callback takes a long time,
	// and ensures that the list of handlers doesn't change while we are iterating over it.
	c.mu.RLock()
	onChangeHandlers := make([]func(fsnotify.Event), len(c.onChange))
	copy(onChangeHandlers, c.onChange)
	c.mu.RUnlock()

	for _, run := range onChangeHandlers {
		run(in)
	}

	log.Info().Dur("reloadTime", time.Since(startTime)).Msg("config loaded successfully")
}

func (c *Config) Parse() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	monitors, ok := c.v.Get("monitor").([]any)
	if !ok {
		return ErrInvalidConfig
	}

	c.Monitors = nil

	for _, item := range monitors {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		var uri, interval, responseTimeThreshold, certExpiredDays string

		if uri, ok = m["url"].(string); !ok {
			continue
		}

		uri = helper.NormalizeURL(uri)

		enabled, ok := m["enabled"].(bool)
		if !ok || !enabled {
			continue
		}

		certMon, _ := m["certificate_monitoring"].(bool)

		if val, ok := m["interval"].(string); ok {
			interval = val
		}

		if val, ok := m["response_time_threshold"].(string); ok {
			responseTimeThreshold = val
		}

		if val, ok := m["certificate_expired_before"].(string); ok {
			certExpiredDays = val
		}

		c.Monitors = append(c.Monitors, MonitorConfig{
			URL:                      uri,
			Enabled:                  enabled,
			Interval:                 parseDuration(interval, "5m"),
			ResponseTimeThreshold:    parseDuration(responseTimeThreshold, "1m"),
			CertificateMonitoring:    certMon,
			CertificateExpiredBefore: parseDuration(certExpiredDays, "30d"),
		})
	}

	return nil
}

func (c *Config) LoadAgentConfig() error {
	return nil
}

func (c *Config) Unmarshal() error {
	return nil
}

func parseDuration(input string, defaultValue string) time.Duration {
	matches := durationRegex.FindAllStringSubmatch(input, -1)

	if len(matches) == 0 && defaultValue != "" {
		return parseDuration(defaultValue, "")
	}

	var total time.Duration
	for _, match := range matches {
		value, _ := strconv.Atoi(match[1])
		unit := match[2]

		switch unit {
		case "s":
			total += time.Duration(value) * time.Second
		case "m":
			total += time.Duration(value) * time.Minute
		case "h":
			total += time.Duration(value) * time.Hour
		case "d":
			total += time.Duration(value) * 24 * time.Hour
		}
	}

	return total
}
