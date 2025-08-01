package configuration

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"uptime-go/internal/net/config"

	"github.com/spf13/viper"
)

type ConfigReader struct {
	viper *viper.Viper
}

func NewConfigReader() *ConfigReader {
	return &ConfigReader{
		viper: viper.New(),
	}
}

func (cr *ConfigReader) ReadConfig(filePath string) error {
	// Set the file name and path
	cr.viper.SetConfigFile(filePath)

	// Set the file type
	cr.viper.SetConfigType("yaml")

	// Set the environment variable prefix
	cr.setDefaults()

	if err := cr.viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func (c *ConfigReader) setDefaults() {
	c.viper.SetDefault("timeout", "5s")
	c.viper.SetDefault("refresh_interval", "10")
	c.viper.SetDefault("follow_redirects", true)
	c.viper.SetDefault("skip_ssl", false)
}

func (c *ConfigReader) ParseConfig() ([]*config.Monitor, error) {
	// TODO: optimize code

	var configList []*config.Monitor

	// Get the monitor configurations
	monitors := c.viper.Get("monitor")
	monitorsList, ok := monitors.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid monitor configuration format")
	}

	for _, m := range monitorsList {
		monitor, ok := m.(map[string]any)
		if !ok {
			continue
		}

		config := &config.Monitor{
			ID: config.GenerateRandomID(),
		}

		// Get URL
		if url, ok := monitor["url"].(string); ok {
			config.URL = url
		}

		// Get enabled
		if enabled, ok := monitor["enabled"].(bool); ok {
			config.Enabled = enabled
		}

		// Get refresh interval
		if refreshInterval, ok := monitor["interval"].(string); ok {
			interval, err := ParseDuration(refreshInterval)
			if err != nil {
				fmt.Printf("%s > failed to parse %s", config.URL, refreshInterval)
			} else {
				config.Interval = interval
			}
		} else {
			config.Interval = 60 * time.Second // Default refresh interval
		}

		// Get timeout
		if timeout, ok := monitor["response_time_threshold"].(string); ok {
			t, err := ParseDuration(timeout)
			if err != nil {
				fmt.Printf("%s > failed to parse %s", config.URL, timeout)
			} else {
				config.ResponseTimeThreshold = t
			}
		} else {
			config.ResponseTimeThreshold = 5 * time.Second // Default timeout
		}

		// Get skip SSL verification
		if skipSSL, ok := monitor["ssl_monitoring"].(bool); ok {
			config.SSLMonitoring = skipSSL
		} else {
			config.SSLMonitoring = false // Default skip SSL
		}

		// Get SSL expired before
		if sslExpired, ok := monitor["ssl_expired_before"].(string); ok {
			expired, err := ParseDuration(sslExpired)
			if err != nil {
				fmt.Printf("%s > failed to parse %s", config.URL, sslExpired)
			} else {
				config.SSLExpiredBefore = expired
			}
		}

		configList = append(configList, config)
	}

	return configList, nil
}

func (c *ConfigReader) GetUptimeConfig() ([]*config.NetworkConfig, error) {
	var configList []*config.NetworkConfig

	// Get the monitor configurations
	monitors := c.viper.Get("monitor")
	monitorsList, ok := monitors.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid monitor configuration format")
	}

	for _, m := range monitorsList {
		monitor, ok := m.(map[string]interface{})
		if !ok {
			continue
		}

		config := &config.NetworkConfig{}

		// Get URL
		if url, ok := monitor["url"].(string); ok {
			config.URL = url
		}

		// Get refresh interval
		if refreshInterval, ok := monitor["refresh_interval"].(int); ok {
			config.RefreshInterval = time.Duration(refreshInterval) * time.Second
		} else {
			config.RefreshInterval = 60 * time.Second // Default refresh interval
		}

		// Get timeout
		if timeout, ok := monitor["timeout"].(int); ok {
			config.Timeout = time.Duration(timeout) * time.Second
		} else {
			config.Timeout = 5 * time.Second // Default timeout
		}

		// Get follow redirects
		if followRedirects, ok := monitor["follow_redirects"].(bool); ok {
			config.FollowRedirects = followRedirects
		} else {
			config.FollowRedirects = true // Default follow redirects
		}

		// Get skip SSL verification
		if skipSSL, ok := monitor["skip_ssl_verification"].(bool); ok {
			config.SkipSSL = skipSSL
		} else {
			config.SkipSSL = false // Default skip SSL
		}

		configList = append(configList, config)
	}

	return configList, nil
}

func (c *ConfigReader) GetDomains(path string) ([]string, error) {
	var domains []string
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("error while reading directory %s", path)
	}

	if !strings.HasSuffix(path, "/") {
		path += "/"
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		filePath := path + f.Name()

		c.ReadConfig(filePath)
		if domain, ok := c.viper.Get("domain").(string); ok {
			domains = append(domains, "https://"+domain)
			continue
		}

		fmt.Printf("can't read domain at %s\n", filePath)
	}

	return domains, nil
}

func ParseDuration(input string) (time.Duration, error) {
	re := regexp.MustCompile(`(\d+)([smhd])`)
	matches := re.FindAllStringSubmatch(input, -1)

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
		default:
			return 0, fmt.Errorf("unknown unit: %s", unit)
		}
	}

	return total, nil
}
