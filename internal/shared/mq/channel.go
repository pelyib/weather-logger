package mq

import (
	"fmt"
	"os"

	"github.com/pelyib/weather-logger/internal/shared"

	amqp "github.com/rabbitmq/amqp091-go"
)

func MakeChannel(cnf shared.Mq, l shared.Logger) *amqp.Channel {
	conn, err := amqp.Dial(
		fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
			cnf.User,
			cnf.Password,
			cnf.Domain,
			cnf.Port,
			cnf.Vhost,
		),
	)

	if err != nil {
		l.Error(fmt.Sprintf("Could not connect to RabbitMQ, reason: %s", err))
		os.Exit(16)
	}

	c, err := conn.Channel()
	if err != nil {
		l.Error(fmt.Sprintf("Could not open channel, reason: %s", err))
		os.Exit(17)
	}

	return c
}
