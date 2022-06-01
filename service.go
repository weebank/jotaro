package rmq

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

func NewService(name string, exchanges ...string) (mS MessagingService) {
	mS.name = name

	mS.queues = 8
	if str, ok := os.LookupEnv("QUEUE_COUNT"); ok {
		if count, err := strconv.Atoi(str); err == nil {
			mS.queues = uint(count)
		}
	}

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
		q := mS.newQueue(i)
		for _, e := range exchanges {
			mS.bindQueue(q, e, i)
		}
	}

	mS.handlers = map[string]func(id string, bytes []byte, index int){}

	return
}

func (mS *MessagingService) newQueue(index int) string {
	q, err := mS.ch.QueueDeclare(
		fmt.Sprint(mS.name, "-", index), // name
		true,                            // durable
		false,                           // auto-deleted
		false,                           // exclusive
		false,                           // no-wait
		nil,                             // arguments
	)
	if err != nil {
		log.Fatalf("couldn't declare a RabbitMQ queue: %s", err.Error())
	}

	return q.Name
}

func (mS MessagingService) newExchange(exchange string) {
	err := mS.ch.ExchangeDeclare(
		exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-deleted
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

// Publish without callback, routing or wrapping id
func (mS *MessagingService) Publish(exchange string, msg protoreflect.ProtoMessage) error {
	return mS.PublishAdvanced("", "", exchange, msg, nil)
}

// Publish without callback, or wrapping id
func (mS *MessagingService) PublishRouted(routing, exchange string, msg protoreflect.ProtoMessage) error {
	return mS.PublishAdvanced("", routing, exchange, msg, nil)
}

// Publish without routing or wrapping id
func (mS *MessagingService) PublishEvent(exchange string, msg protoreflect.ProtoMessage, callback func(bytes []byte, index int)) error {
	return mS.PublishAdvanced("", "", exchange, msg, callback)
}

// Publish without callback or routing
func (mS *MessagingService) PublishResponse(id, exchange string, msg protoreflect.ProtoMessage) error {
	return mS.PublishAdvanced(id, "", exchange, msg, nil)
}

// Publish message
func (mS *MessagingService) PublishAdvanced(id, routing, exchange string, msg protoreflect.ProtoMessage, callback func(bytes []byte, index int)) error {
	// Add callback ID
	if id == "" {
		id = uuid.NewString()
	}

	// Add routing key
	if routing == "" {
		routing = uuid.NewString()
	}

	// Set callback
	if callback != nil {
		mS.callbacks[id] = callback
	}

	// Publish message
	err := mS.ch.Publish(
		exchange, // exchange
		fmt.Sprint(queueIndex(routing, mS.queues)), // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        wrap(id, msg),
		})

	return err
}

func (mS *MessagingService) Bind(msg protoreflect.ProtoMessage, handler func(id string, bytes []byte, index int)) {
	any, _ := anypb.New(msg)
	mS.handlers[any.TypeUrl] = handler
}

func (mS MessagingService) Consume() {
	forever := make(chan struct{})
	for i := uint(0); i < mS.queues; i++ {
		go func(i uint) {
			msgs, err := mS.ch.Consume(
				fmt.Sprint(mS.name, "-", i), // queue
				"",                          // consumer
				true,                        // auto ack
				true,                        // exclusive
				false,                       // no local
				false,                       // no wait
				nil,                         // args
			)
			if err != nil {
				return
			}

			for msg := range msgs {
				if id, any, err := unwrap(msg.Body); err == nil {
					index, _ := strconv.Atoi(msg.RoutingKey)
					if callback, ok := mS.callbacks[id]; ok {
						callback(any.GetValue(), index)
					}
					if handler, ok := mS.handlers[any.TypeUrl]; ok {
						handler(id, any.GetValue(), index)
					}
				}
			}
		}(i)
	}
	<-forever
}

func (mS *MessagingService) Close() {
	mS.ch.Close()
	mS.conn.Close()
}
