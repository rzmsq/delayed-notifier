package main

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rabbitmqUrl = os.Getenv("RABBITMQ_URL")
	queueName   = "message_queue"
)

type messageRequest struct {
	Type      string `json:"type"`
	Recipient string `json:"recipient"`
	Text      string `json:"text"`
	DelaySec  int    `json:"delay_sec"`
}

func main() {

	conn, err := amqp.Dial(rabbitmqUrl)
	panicOnError(err)
	defer func() {
		err = conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	ch, err := conn.Channel()
	panicOnError(err)
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
	panicOnError(err)

	err = ch.Qos(
		1,
		0,
		false,
	)
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
	var msgReq messageRequest

	go func() {
		for msg := range msgs {
			slog.Info("Received a message:", msg.Body)

			err = json.Unmarshal(msg.Body, &msgReq)
			panicOnError(err)

			time.Sleep(time.Duration(msgReq.DelaySec) * time.Second)

			slog.Info("Sending message:", msg.Body)

			err = msg.Ack(false)
			panicOnError(err)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
