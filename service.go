package rmq

import (
	"fmt"
	"log"
	"runtime"
	"strconv"

	"github.com/streadway/amqp"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

func NewService(exchanges ...string) (mS MessagingService) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("couldn't connect to RabbitMQ: %s", err.Error())
	}
	mS.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("couldn't open a RabbitMQ channel: %s", err.Error())
	}
	mS.ch = ch

	for _, e := range exchanges {
		mS.newExchange(e)
	}

	for i := 0; i < runtime.NumCPU(); i++ {
		q := mS.newQueue()
		for _, e := range exchanges {
			mS.bindQueue(q, e, i)
		}
	}

	mS.handlers = map[string]func(bytes []byte, index int){}

	return
}

func (mS *MessagingService) newQueue() string {
	q, err := mS.ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // auto-deleted
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Fatalf("couldn't declare a RabbitMQ queue: %s", err.Error())
	}

	mS.queues = append(mS.queues, q)

	return q.Name
}

func (mS MessagingService) newExchange(exchange string) {
	err := mS.ch.ExchangeDeclare(
		exchange, // name
		"direct", // type
		false,    // durable
		true,     // auto-deleted
		false,    // internal
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		log.Fatalf("couldn't declare a RabbitMQ exchange: %s", err.Error())
	}
}

func (mS MessagingService) bindQueue(queue string, exchange string, index int) {
	err := mS.ch.QueueBind(
		queue,             // queue name
		fmt.Sprint(index), // routing key
		exchange,          // exchange
		false,             // no-wait
		nil,               // arguments
	)
	if err != nil {
		log.Fatalf("couldn't declare a RabbitMQ exchange: %s", err.Error())
	}
}

func (mS MessagingService) Publish(id string, msg protoreflect.ProtoMessage, exchange string) error {
	err := mS.ch.Publish(
		exchange,                   // exchange
		fmt.Sprint(queueIndex(id)), // routing key
		false,                      // mandatory
		false,                      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        wrap(msg),
		})

	return err
}

func (mS *MessagingService) Bind(msg protoreflect.ProtoMessage, handler func(bytes []byte, index int)) {
	any, _ := anypb.New(msg)
	mS.handlers[any.TypeUrl] = handler

}

func (mS MessagingService) Consume() {
	forever := make(chan struct{})
	for _, q := range mS.queues {
		go func(qName string) {
			msgs, err := mS.ch.Consume(
				qName, // queue
				"",    // consumer
				true,  // auto ack
				false, // exclusive
				false, // no local
				false, // no wait
				nil,   // args
			)
			if err != nil {
				return
			}

			for msg := range msgs {
				if any, err := unwrap(msg.Body); err == nil {
					if handler, ok := mS.handlers[any.TypeUrl]; ok {
						index, _ := strconv.Atoi(msg.RoutingKey)
						handler(any.GetValue(), index)
					}
				}
			}
		}(q.Name)
	}
	<-forever
}
