package queue

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
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
		log.WithError(err).Errorf("[Event] %s", "Failed to connect to RabbitMQ")
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to open a channel")
		return
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(xname, "fanout", true, false, false, false, nil)
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to declare an exchange")
		return
	}

	q, err := ch.QueueDeclare(qname, false, false, false, false, nil)
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to declare a queue")
		return
	}

	err = ch.QueueBind(q.Name, "", xname, false, nil)
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to bind a queue")
		return
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", "Failed to consume the queue")
		return
	}

	forever := make(chan bool)

	for d := range msgs {
		log.Debugf("[Event] Received message: %s", d.Body)

		err := callback(string(d.Body))
		if err != nil {
			d.Nack(false, true)
		}
		time.Sleep(10 * time.Second)
	}

	log.WithFields(log.Fields{"network": network, "exchange": xname, "queue": qname}).Debugf("[Event] Waiting for messages")
	<-forever
}

func (c *Consumer) handleError(err error, msg string) {
	if err != nil {
		log.WithError(err).Errorf("[Event] %s", msg)
	}
}
