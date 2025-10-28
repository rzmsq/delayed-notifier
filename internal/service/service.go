package service

import (
	"delayed-notifier/internal/deliver"
	"delayed-notifier/internal/models"

	"github.com/wb-go/wbf/rabbitmq"
)

func CreateNotification(request *models.Notification, conn *rabbitmq.Channel) error {
	err := deliver.CreateNotification(conn, request, request.Channel)
	if err != nil {
		return err
	}

	return nil
}
