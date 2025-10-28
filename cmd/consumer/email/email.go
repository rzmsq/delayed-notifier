package main

import (
	"context"
	appConfig "delayed-notifier/internal/app-config"
	"delayed-notifier/internal/models"
	"encoding/json"
	"flag"
	"os"
	"strconv"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-gomail/gomail"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"
)

const (
	yamlPath       = "config.yaml"
	maxElapseTimes = 2 * time.Minute
	maxAttempts    = 5
)

func main() {
	zlog.InitConsole()

	var configPath string
	flag.StringVar(&configPath, "config", yamlPath, "Path to app-config file")
	flag.Parse()

	err := appConfig.InitConfigs(configPath)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Error loading app-config file")
		os.Exit(1)
	}
	cfg := appConfig.Cfg

	redisClient := redis.New(cfg.RedisConfig.Host+":"+cfg.RedisConfig.Port, cfg.RedisConfig.Passwords, cfg.RedisConfig.Db)

	conn, err := amqp.Dial(cfg.RabbitConfig.RabbitmqUrl)
	panicOnError(err)
	defer func() {
		err = conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	run(conn, redisClient)
}

func run(conn *amqp.Connection, redisClient *redis.Client) {
	cfg := appConfig.Cfg
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
		for delivery := range msgs {
			var notification models.Notification
			err = json.Unmarshal(delivery.Body, &notification)
			if err != nil {
				zlog.Logger.Error().Err(err).Msg("unmarshalling message")
				err = delivery.Reject(false)
				if err != nil {
					zlog.Logger.Error().Err(err).Msg("rejecting message")
				}
				continue
			}

			m := gomail.NewMessage()
			m.SetHeader("From", cfg.SMTPConfig.From)
			m.SetHeader("To", notification.Recipient)
			m.SetHeader("Subject", "Notification")
			m.SetBody("text/html", notification.Message)

			var port int64
			port, err = strconv.ParseInt(cfg.SMTPConfig.Port, 10, 64)
			if err != nil {
				zlog.Logger.Error().Err(err).Msg("parsing port number")
				err = delivery.Reject(false)
				if err != nil {
					zlog.Logger.Error().Err(err).Msg("rejecting message")
				}
				continue
			}
			d := gomail.NewDialer(
				cfg.SMTPConfig.Host,
				int(port),
				cfg.SMTPConfig.From,
				cfg.SMTPConfig.Passwords,
			)

			if err = d.DialAndSend(m); err != nil {
				zlog.Logger.Error().Err(err).Msg("failed to send email")
				continue
			}

			zlog.Logger.Info().Str("recipient", notification.Recipient).Msg("email sent successfully")

			err = delivery.Ack(false)
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

			operation := func() error {
				return redisClient.Set(ctx, notification.ID, notificationJSON)
			}

			expBackoff := backoff.NewExponentialBackOff()
			expBackoff.MaxElapsedTime = maxElapseTimes

			err = backoff.Retry(operation, backoff.WithMaxRetries(expBackoff, maxAttempts))

			if err != nil {
				zlog.Logger.Error().Err(err).Msg("redis set status notification failed after multiple retries")
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
