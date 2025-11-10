package config

import (
	"errors"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	OJTGUARDIAN_PATH   = "/etc/ojtguardian"
	OJTGUARDIAN_CONFIG = OJTGUARDIAN_PATH + "/main.yml"
	PLUGIN_PATH        = OJTGUARDIAN_PATH + "/plugins/uptime"
	PLUGIN_CONFIG_PATH = OJTGUARDIAN_PATH + "/plugins.yml"
)

var (
	instance *Config
	once     sync.Once
)

func Get() *Config {
	once.Do(func() {
		instance = &Config{
			v: viper.New(),
		}
	})

	return instance
}

type Config struct {
	v *viper.Viper
	// mu      sync.RWMutex
	Agent   AgentConfig
	Monitor Monitor
}

type AgentConfig struct {
	MasterHost string `mapstructure:"master_host"`
	Auth       struct {
		Token string
	}
}

type Monitor struct {
	URL                      string `mapstructure:"url" yaml:"url" json:"url"`
	Enabled                  bool   `mapstructure:"enabled" yaml:"enabled" json:"enabled"`
	Interval                 string `mapstructure:"interval" yaml:"interval" json:"interval"`
	ResponseTimeThreshold    string `mapstructure:"response_time_threshold" yaml:"response_time_threshold" json:"response_time_threshold"`
	CertificateMonitoring    bool   `mapstructure:"certificate_monitoring" yaml:"certificate_monitoring" json:"certificate_monitoring"`
	CertificateExpiredBefore string `mapstructure:"certificate_expired_before" yaml:"certificate_expired_before" json:"certificate_expired_before"`
}

func Init() error {
	Get()

	instance.v.SetConfigType("yml")
	instance.v.SetConfigName("config")
	instance.v.AddConfigPath("/etc/uptime-go")
	instance.v.AddConfigPath(PLUGIN_PATH)
	instance.v.AddConfigPath("./configs")

	if err := loadAgentConfig(&instance.Agent); err != nil {
		// TODO: optional?
		return err
	}

	instance.v.OnConfigChange(func(in fsnotify.Event) {
		var monitor []Monitor
		if err := instance.v.Unmarshal(&monitor); err != nil {
			// TODO: handle
		}

		for _, m := range monitor {
			log.Info().Msgf("%+v", m)
		}
	})

	return nil
}

func Load() {
}

func loadAgentConfig(agentConfig *AgentConfig) error {
	v := viper.New()
	v.SetConfigFile(OJTGUARDIAN_CONFIG)

	var fileLookupError viper.ConfigFileNotFoundError
	if err := v.ReadInConfig(); err != nil {
		if errors.As(err, &fileLookupError) {
			// TODO: improve log and set default
			log.Warn().Msgf("failed to read file: %s", err)
			return nil
		}

		return err
	}

	// if err := v.UnmarshalKey("plugins.uptime-go", agentConfig); err != nil {
	// 	// TODO: handle
	// 	return err
	// }

	if err := v.Unmarshal(agentConfig); err != nil {
		// TODO: handle
		return err
	}

	return nil
}
