package service

import (
	"delayed-notifier/internal/deliver"
	"delayed-notifier/internal/models"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var rabbitmqUrl = os.Getenv("RABBITMQ_URL")

func CreateNotification(request *models.CreateNotificationRequest) error {
	conn, err := amqp.Dial(rabbitmqUrl)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer func() {
		err = ch.Close()
		if err != nil {
			panic(err)
		}
	}()

	err = deliver.CreateNotification(ch, request, request.Channel)
	if err != nil {
		return err
	}

	return nil
}
