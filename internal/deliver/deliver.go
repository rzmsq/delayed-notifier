package deliver

import (
	"delayed-notifier/internal/models"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func CreateNotification(ch *amqp.Channel, notify *models.Notification, key string) error {
	args := amqp.Table{
		"x-delayed-type": "direct",
	}
	err := ch.ExchangeDeclare(
		"delayed_notification",
		"x-delayed-message",
		true,
		false,
		false,
		false,
		args,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(notify)
	if err != nil {
		return err
	}

	delay := time.Until(notify.SendAt).Milliseconds()
	if delay < 0 {
		delay = 0
	}

	headers := amqp.Table{
		"x-delay": delay,
	}

	err = ch.Publish(
		"delayed_notification",
		key,
		false,
		false,
		amqp.Publishing{
			Headers:      headers,
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if err != nil {
		return err
	}
	return nil
}
