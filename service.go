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

var queues = uint(8)

type SubscriptionMode uint8

func NewService(name string, exchanges []string) (mS MessagingService) {
	// Create RabbitMQ connection
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("couldn't connect to RabbitMQ: %s", err.Error())
	}
	mS.conn = conn

	// Create RabbitMQ channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("couldn't open a RabbitMQ channel: %s", err.Error())
	}
	mS.ch = ch

	// Set service name
	mS.name = name

	// Bind exchanges
	for _, e := range exchanges {
		mS.newExchange(e)
	}

	// Load queue count preference from env
	if str, ok := os.LookupEnv("QUEUE_COUNT"); ok {
		if count, err := strconv.Atoi(str); err == nil {
			queues = uint(count)
		}
	}

	// Create queues
	if len(exchanges) > 0 {
		// Create runtime.NumCPU() queues
		for i := 0; i < runtime.NumCPU(); i++ {
			q := mS.newQueue(i)
			// Bind exchanges to these queues
			for _, e := range exchanges {
				mS.bindQueue(q, e, i)
			}
		}

		// Add handlers and callbacks
		mS.handlers = map[string]func(queue uint, bytes []byte) *Response{}
		mS.callbacks = map[string]func(queue uint, bytes []byte){}
	}

	return
}

// Create queue on service
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

// Bind exchange to service
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

// Bind queue to exchange
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
func (mS *MessagingService) PublishRouted(route, exchange string, msg protoreflect.ProtoMessage) error {
	return mS.PublishAdvanced("", route, exchange, msg, nil)
}

// Publish without routing or wrapping id
func (mS *MessagingService) PublishEvent(exchange string, msg protoreflect.ProtoMessage, callback func(queue uint, bytes []byte)) error {
	return mS.PublishAdvanced("", "", exchange, msg, callback)
}

// Publish message
func (mS *MessagingService) PublishAdvanced(id, route, exchange string, msg protoreflect.ProtoMessage, callback func(queue uint, bytes []byte)) error {
	// Add callback ID if needed
	isCallback := id != ""
	if isCallback {
		id = uuid.NewString()
	}

	// Add routing key
	if route == "" {
		route = uuid.NewString()
	}

	// Set callback
	if callback != nil {
		mS.callbacks[id] = callback
	}

	// Publish message
	err := mS.ch.Publish(
		exchange,                              // exchange
		fmt.Sprint(queueIndex(route, queues)), // routing key
		false,                                 // mandatory
		false,                                 // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        wrap(id, isCallback, msg),
		})

	return err
}

// Bind handler to message
func (mS *MessagingService) Bind(msg protoreflect.ProtoMessage, handler func(queue uint, bytes []byte) *Response) {
	any, _ := anypb.New(msg)
	mS.handlers[any.TypeUrl] = handler
}

// Start consuming queues
func (mS MessagingService) Consume() {
	forever := make(chan struct{})
	for i := uint(0); i < queues; i++ {
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
				if id, isCallback, any, err := unwrap(msg.Body); err == nil {
					index, _ := strconv.Atoi(msg.RoutingKey)
					if isCallback {
						if callback, ok := mS.callbacks[id]; ok {
							callback(uint(index), any.GetValue())
						}
					} else {
						if handler, ok := mS.handlers[any.TypeUrl]; ok {
							if res := handler(uint(index), any.GetValue()); res != nil {
								mS.PublishAdvanced(id, res.Route, msg.Exchange, res.Message, nil)
							}
						}
					}
				}
			}
		}(i)
	}
	<-forever
}

// Close connection
func (mS *MessagingService) Close() {
	mS.ch.Close()
	mS.conn.Close()
}
