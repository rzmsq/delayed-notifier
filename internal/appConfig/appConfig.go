package appConfig

import (
	"time"

	"github.com/wb-go/wbf/config"
)

var Cfg *AppConfig

type AppConfig struct {
	ServerConfig *serverConfig
	RabbitConfig *rabbitConfig
}

type serverConfig struct {
	Addr string
}

type rabbitConfig struct {
	RabbitmqUrl  string
	RetryConnect int
	PauseRetry   time.Duration
}

func newAppConfig(configPath string) (*AppConfig, error) {
	cfg := config.New()
	err := cfg.LoadConfigFiles(configPath)
	if err != nil {
		return nil, err
	}

	return &AppConfig{
		ServerConfig: &serverConfig{
			Addr: cfg.GetString("PORT"),
		},
		RabbitConfig: &rabbitConfig{
			RabbitmqUrl:  cfg.GetString("RABBIT_URL"),
			RetryConnect: cfg.GetInt("RETRY_CONNECT"),
			PauseRetry:   cfg.GetDuration("PAUSE_RETRY"),
		},
	}, nil
}

func InitConfigs(configPath string) error {
	var err error
	Cfg, err = newAppConfig(configPath)
	if err != nil {
		return err
	}
	return nil
}
