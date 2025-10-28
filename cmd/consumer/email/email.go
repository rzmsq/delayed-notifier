package main

import (
	"context"
	app_config "delayed-notifier/internal/app-config"
	"delayed-notifier/internal/models"
	"encoding/json"
	"flag"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"
)

const yamlPath = "config.yaml"

func main() {
	zlog.InitConsole()

	var configPath string
	flag.StringVar(&configPath, "config", yamlPath, "Path to app-config file")
	flag.Parse()

	err := app_config.InitConfigs(configPath)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Error loading app-config file")
		os.Exit(1)
	}
	cfg := app_config.Cfg

	redisClient := redis.New(cfg.RedisConfig.Host+":"+cfg.RedisConfig.Port, cfg.RedisConfig.Passwords, cfg.RedisConfig.Db)

	conn, err := amqp.Dial(cfg.RabbitConfig.RabbitmqUrl)
	panicOnError(err)
	defer func() {
		err = conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	run(err, conn, redisClient)
}

func run(err error, conn *amqp.Connection, redisClient *redis.Client) {
	ch, err := conn.Channel()
	panicOnError(err)
	defer func() {
		err = ch.Close()
		if err != nil {
			panic(err)
		}
	}()

	args := make(amqp.Table)
	args["x-delayed-type"] = "direct"
	err = ch.ExchangeDeclare(
		"delayed_notification",
		"x-delayed-message",
		true,
		false,
		false,
		false,
		args,
	)
	panicOnError(err)

	q, err := ch.QueueDeclare(
		"email_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	panicOnError(err)

	err = ch.QueueBind(
		q.Name,
		"email",
		"delayed_notification",
		false,
		nil)
	panicOnError(err)

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	panicOnError(err)

	var forever chan struct{}

	ctx := context.Background()
	go func() {
		for d := range msgs {
			var notification models.Notification
			err = json.Unmarshal(d.Body, &notification)
			if err != nil {
				zlog.Logger.Error().Err(err).Msg("unmarshalling message")
				err = d.Reject(false)
				if err != nil {
					zlog.Logger.Error().Err(err).Msg("rejecting message")
				}
				continue
			}

			err = d.Ack(false)
			if err != nil {
				zlog.Logger.Error().Err(err).Msg("ack message")
				continue
			}

			notification.Status = models.StatusSent
			notificationJSON, marshalErr := json.Marshal(notification)
			if marshalErr != nil {
				zlog.Logger.Error().Err(marshalErr).Msg("failed to marshal notification for redis")
				continue
			}

			err = redisClient.Set(ctx, notification.ID, notificationJSON)
			if err != nil {
				zlog.Logger.Error().Err(err).Msg("redis set status notification")
			} else {
				zlog.Logger.Info().Interface("notification", notification).Msg("redis set status notification")
			}
		}
	}()

	zlog.Logger.Info().Msg(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
