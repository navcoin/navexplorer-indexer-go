package queue

import (
	"fmt"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Publisher interface {
	PublishToQueue(name string, msg string)
}

type publisher struct {
	network string
	index   string
	address string
}

func NewPublisher(network string, index string, user string, password string, host string, port int) Publisher {
	return publisher{
		network: network,
		index:   index,
		address: fmt.Sprintf("amqp://%s:%s@%s:%d", user, password, host, port),
	}
}

func (p publisher) PublishToQueue(name string, msg string) {
	go func() {
		xname := fmt.Sprintf("%s.%s.%s", p.network, p.index, name)

		conn, err := amqp.Dial(p.address)
		if err != nil {
			zap.L().With(zap.Error(err)).Error("Publisher: Failed to connect to RabbitMQ")
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			zap.L().With(zap.Error(err)).Error("Publisher: Failed to open a channel")
		}
		defer ch.Close()

		err = ch.ExchangeDeclare(xname, "fanout", true, false, false, false, nil)
		if err != nil {
			zap.L().With(zap.Error(err)).Error("Publisher: Failed to declare an exchange")
		}

		err = ch.Publish(xname, "", false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
		if err != nil {
			zap.L().With(zap.Error(err)).Error("Publisher: Failed to publish a message")
		}

		zap.L().With(zap.String("msg", msg), zap.String("exchange", xname)).Info("Publisher: Message Sent")
	}()
}
