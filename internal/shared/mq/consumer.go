package mq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
  Exchange string
  Handlers map[string]Executor
  C *amqp.Channel
}

type Executor interface {
  Execute(msgBody []byte)
}

func (c Consumer) Consume() {
  commands, err := c.C.Consume("command", c.Exchange, false, false, false, false, nil)

  if err != nil {
    fmt.Println("Can not consume from `command` queue`")
  }

  fmt.Println("start to listening")

  forever := make(chan bool)

  go func () {
    for msg := range commands {
      fmt.Println(fmt.Sprintf("routingKey: %s | body: %s", string(msg.RoutingKey), string(msg.Body)))

      if _, ok := c.Handlers[msg.RoutingKey]; ok {
        c.Handlers[msg.RoutingKey].Execute(msg.Body)
        msg.Ack(false)
      } else {
        msg.Reject(true)
        fmt.Println(fmt.Sprintf("MQ Consumer | Missing executor for %s message", msg.RoutingKey))
      }
    }
  }()

  fmt.Printf(" [*] Waiting for messages. To exit press CTRL+C")
  <-forever
}
