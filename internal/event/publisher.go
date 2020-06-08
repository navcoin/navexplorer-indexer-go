package event

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type Publisher struct {
	network string
	address string
}

func NewPublisher(network string, user string, password string, host string, port int) *Publisher {
	return &Publisher{
		network: network,
		address: fmt.Sprintf("amqp://%s:%s@%s:%d", user, password, host, port),
	}
}

func (p *Publisher) PublishToQueue(name string, msg string) {
	go func() {
		xname := fmt.Sprintf("%s.%s", p.network, name)

		conn, err := amqp.Dial(p.address)
		p.handleError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		ch, err := conn.Channel()
		p.handleError(err, "Failed to open a channel")
		defer ch.Close()

		err = ch.ExchangeDeclare(xname, "fanout", true, false, false, false, nil)
		p.handleError(err, "Failed to declare an exchange")

		err = ch.Publish(xname, "", false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
		log.Infof("[Event] Sent %s to %s", msg, xname)
		p.handleError(err, "Failed to publish a message")
	}()
}

func (p *Publisher) handleError(err error, msg string) {
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", msg)
	}
}
