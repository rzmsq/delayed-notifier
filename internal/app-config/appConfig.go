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
	SMTPConfig   *smtpConfig
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

type smtpConfig struct {
	Host      string
	Port      string
	From      string
	Passwords string
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
		SMTPConfig: &smtpConfig{
			Host:      cfg.GetString("SMTP_HOST"),
			Port:      cfg.GetString("SMTP_PORT"),
			From:      cfg.GetString("SMTP_FROM"),
			Passwords: cfg.GetString("SMTP_PASSWORD"),
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
