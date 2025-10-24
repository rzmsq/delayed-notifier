package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rabbitmqUrl = os.Getenv("RABBITMQ_URL")
	queueName   = "message_queue"
)

func CreateNotify(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := r.Body.Close()
		if err != nil {
			slog.Error("body request close error: ", err)
		}
	}()

	conn, err := amqp.Dial(rabbitmqUrl)
	if isError(w, err, http.StatusInternalServerError) {
		return
	}

	ch, err := conn.Channel()
	if isError(w, err, http.StatusInternalServerError) {
		return
	}
	defer func() {
		err = ch.Close()
		if err != nil {
			panic(err)
		}
	}()

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if isError(w, err, http.StatusInternalServerError) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	body, err := io.ReadAll(r.Body)
	if isError(w, err, http.StatusBadRequest) {
		return
	}

	err = ch.PublishWithContext(
		ctx,
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
	if isError(w, err, http.StatusInternalServerError) {
		return
	}

	slog.Info("Sent", body)
	w.WriteHeader(http.StatusOK)
}

func GetNotify(w http.ResponseWriter, r *http.Request) {}

func DeleteNotify(w http.ResponseWriter, r *http.Request) {}

func isError(w http.ResponseWriter, err error, code int) bool {
	if err != nil {
		slog.Error("error", err)
		w.WriteHeader(code)
		return true
	}
	return false
}
