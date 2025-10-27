package service

import (
	"delayed-notifier/internal/deliver"
	"delayed-notifier/internal/models"

	"github.com/wb-go/wbf/rabbitmq"
)

func CreateNotification(request *models.CreateNotificationRequest, conn *rabbitmq.Connection) error {
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
