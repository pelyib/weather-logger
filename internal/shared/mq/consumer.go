package mq

import (
	"fmt"

	"github.com/pelyib/weather-logger/internal/shared"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Exchange string
	Handlers map[string]Executor
	C        *amqp.Channel
  L shared.Logger
}

type Executor interface {
	Execute(msgBody []byte)
}

func (c Consumer) Consume() error {
	commands, err := c.C.Consume("command", c.Exchange, false, false, false, false, nil)

	if err != nil {
		c.L.Error("Can not consume from `command` queue")
    return err
	}

	c.L.Info("start to listening")

	forever := make(chan bool)

	go func() {
		for msg := range commands {
			c.L.Info(fmt.Sprintf("routingKey: %s | body: %s", string(msg.RoutingKey), string(msg.Body)))

			if _, ok := c.Handlers[msg.RoutingKey]; ok {
				c.Handlers[msg.RoutingKey].Execute(msg.Body)
				msg.Ack(false)
			} else {
				msg.Reject(true)
				c.L.Info(fmt.Sprintf("Missing executor for %s message", msg.RoutingKey))
			}
		}
	}()

	c.L.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever

  return nil
}
