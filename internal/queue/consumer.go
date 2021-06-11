package queue

import (
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"time"
)

type Consumer struct {
	address string
}

func NewConsumer(user, password, host string, port int) *Consumer {
	return &Consumer{
		address: fmt.Sprintf("amqp://%s:%s@%s:%d", user, password, host, port),
	}
}

func (c *Consumer) Consume(network, index, name string, callback func(msg string) error) {
	xname := fmt.Sprintf("%s.%s.%s", network, index, name)
	qname := xname

	conn, err := amqp.Dial(c.address)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("Queue: Failed to connect to RabbitMQ")
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		zap.L().With(zap.Error(err)).Error("Queue: Failed to open a channel")
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(xname, "fanout", true, false, false, false, nil)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("Queue: Failed to declare an exchange")
		return
	}

	q, err := ch.QueueDeclare(qname, false, false, false, false, nil)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("Queue: Failed to declare a queue")
		return
	}

	err = ch.QueueBind(q.Name, "", xname, false, nil)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("Queue: Failed to bind a queue")
		return
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		zap.L().With(zap.Error(err)).Error("Queue: Failed to consume the queue")
		return
	}

	forever := make(chan bool)

	for d := range msgs {
		zap.L().With(zap.Error(err), zap.ByteString("body", d.Body)).Debug("Queue: Received message")

		if err := callback(string(d.Body)); err != nil {
			d.Nack(false, true)
		}
		time.Sleep(10 * time.Second)
	}

	zap.L().With(
		zap.String("network", network),
		zap.String("exchange", xname),
		zap.String("queue", qname),
	).Debug("Queue: Waiting for messages")

	<-forever
}
