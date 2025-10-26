package main

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	rabbitmqUrl = os.Getenv("RABBITMQ_URL")
)

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

	args := make(amqp.Table)
	args["x-delayed-type"] = "direct"
	err = ch.ExchangeDeclare(
		"delayed_notification", // Exchange name
		"x-delayed-message",    // Exchange type
		true,                   // durable
		false,                  // auto-deleted
		false,                  // internal
		false,                  // no-wait
		args,                   // arguments
	)
	panicOnError(err)

	q, err := ch.QueueDeclare(
		"tg_queue", // name
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	panicOnError(err)

	err = ch.QueueBind(
		q.Name,                 // queue name
		"telegram",             // routing key
		"delayed_notification", // exchange
		false,
		nil)
	panicOnError(err)

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto ack -> false
		false,  // exclusive
		false,  // no local
		false,  // no wait
		nil,    // args
	)
	panicOnError(err)

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf(" [x] %s", d.Body)
			err := d.Ack(false) // Acknowledge the message
			if err != nil {
				log.Printf("Error acknowledging message: %s", err)
			}
		}
	}()

	log.Printf(" [*] Waiting for logs. To exit press CTRL+C")
	<-forever
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}
