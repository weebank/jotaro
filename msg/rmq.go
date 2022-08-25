package msg

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Channel struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// Connect to RabbitMQ
func connect() Channel {
	// Create RabbitMQ connection
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		logger.WithError(err).Error("couldn't connect to RabbitMQ")
	}
	// Create RabbitMQ channel
	ch, err := conn.Channel()
	if err != nil {
		logger.WithError(err).Error("couldn't open a RabbitMQ channel")
	}
	return Channel{conn, ch}
}

// Create queue on service
func (c Channel) newQueue(name string) {
	_, err := c.ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // auto-deleted
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		logger.WithError(err).Error("couldn't declare RabbitMQ queue")
	}
}

// Create a consumer for a specified queue
func (c Channel) consumeQueue(name string) <-chan amqp.Delivery {
	msgs, err := c.ch.Consume(
		name,  // queue
		"",    // consumer
		false, // auto ack
		false, // exclusive
		false, // no local
		false, // no wait
		nil,   // args
	)
	if err != nil {
		logger.WithError(err).Error("couldn't consume RabbitMQ queue")
	}

	return msgs
}

// Publish message
func (c Channel) publish(body []byte, target string) error {
	return c.ch.PublishWithContext(
		context.Background(),
		"",     // exchange
		target, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

// Close channel and connection
func (c Channel) close() {
	c.ch.Close()
	c.conn.Close()
}
