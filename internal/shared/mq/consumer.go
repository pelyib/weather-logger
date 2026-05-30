package mq

import (
	"context"
	"fmt"

	"github.com/pelyib/weather-logger/internal/shared"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Exchange string
	Handlers map[string]Executor
	C        *amqp.Channel
	L        shared.Logger
}

type Executor interface {
	Execute(msgBody []byte) error
}

func (c Consumer) Consume(ctx context.Context) error {
	commands, err := c.C.Consume("command", c.Exchange, false, false, false, false, nil)

	if err != nil {
		c.L.Error("Can not consume from `command` queue")
		return err
	}

	c.L.Info("start to listening")

	go func() {
		for msg := range commands {
			c.L.Info(fmt.Sprintf("routingKey: %s | body: %s", string(msg.RoutingKey), string(msg.Body)))

			if handler, ok := c.Handlers[msg.RoutingKey]; ok {
				if err := handler.Execute(msg.Body); err != nil {
					c.L.Error(fmt.Sprintf("Executor failed for %s: %s", msg.RoutingKey, err))
					msg.Nack(false, false)
				} else {
					msg.Ack(false)
				}
			} else {
				msg.Reject(true)
				c.L.Info(fmt.Sprintf("Missing executor for %s message", msg.RoutingKey))
			}
		}
	}()

	c.L.Info(" [*] Waiting for messages. To exit press CTRL+C")
	<-ctx.Done()
	c.L.Info("Consumer shutting down")

	return nil
}
