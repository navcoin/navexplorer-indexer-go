package queue

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
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
		handleError(err, "Failed to connect to RabbitMQ")
		defer conn.Close()

		ch, err := conn.Channel()
		handleError(err, "Failed to open a channel")
		defer ch.Close()

		err = ch.ExchangeDeclare(xname, "fanout", true, false, false, false, nil)
		handleError(err, "Failed to declare an exchange")

		err = ch.Publish(xname, "", false, false, amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(msg),
		})
		log.Infof("[Event] Sent %s to %s", msg, xname)
		handleError(err, "Failed to publish a message")
	}()
}

func handleError(err error, msg string) {
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", msg)
	}
}
