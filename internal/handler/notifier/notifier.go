package notifier

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const rabbitmqUrl = "amqp://guest:guest@localhost:5672/"

func CreateNotify(w http.ResponseWriter, r *http.Request) {
	conn, err := amqp.Dial(rabbitmqUrl)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			slog.Error("Failed to close connection", err)
		}
	}()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer func() {
		err = ch.Close()
		if err != nil {
			panic(err)
		}
	}()

	q, err := ch.QueueDeclare(
		"hello",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body := "hello world"
	err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		},
	)
	if err != nil {
		panic(err)
	}
	slog.Info("Sent", body)

}

func GetNotify(w http.ResponseWriter, r *http.Request) {}

func DeleteNotify(w http.ResponseWriter, r *http.Request) {}
