package app_config

import (
	"time"

	"github.com/wb-go/wbf/config"
)

var Cfg *AppConfig

type AppConfig struct {
	ServerConfig *serverConfig
	RabbitConfig *rabbitConfig
	RedisConfig  *redisConfig
}

type serverConfig struct {
	Port string
}

type rabbitConfig struct {
	RabbitmqUrl  string
	RetryConnect int
	PauseRetry   time.Duration
}

type redisConfig struct {
	Host      string
	Port      string
	Passwords string
	Db        int
}

func newAppConfig(configPath string) (*AppConfig, error) {
	cfg := config.New()
	err := cfg.LoadConfigFiles(configPath)
	if err != nil {
		return nil, err
	}

	return &AppConfig{
		ServerConfig: &serverConfig{
			Port: cfg.GetString("PORT"),
		},
		RabbitConfig: &rabbitConfig{
			RabbitmqUrl:  cfg.GetString("RABBIT_URL"),
			RetryConnect: cfg.GetInt("RETRY_CONNECT"),
			PauseRetry:   cfg.GetDuration("PAUSE_RETRY"),
		},
		RedisConfig: &redisConfig{
			Host:      cfg.GetString("REDIS_HOST"),
			Port:      cfg.GetString("REDIS_PORT"),
			Passwords: cfg.GetString("REDIS_PASSWORD"),
			Db:        cfg.GetInt("REDIS_DB"),
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
