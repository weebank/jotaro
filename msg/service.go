package msg

import (
	"encoding/json"
	"log"
	"runtime"
)

// Messaging Service
type MessagingService struct {
	name     string
	channel  Channel
	handlers map[string]func(m Message) (content any, err error)
}

// Initialize Messaging Service
func NewService(name string) (mS *MessagingService) {
	// Declare Messaging Service
	mS = new(MessagingService)

	// Set connection
	mS.channel = connect()
	// Set service name
	mS.name = name

	// Create service queue
	mS.channel.newQueue(mS.name)

	// Add handlers, callbacks and subscriptions
	mS.handlers = make(map[string]func(m Message) (content any, err error))

	return
}

// Publish Message Internal
func publish(mS *MessagingService, target, event string, msg any, msgErr error) error {
	// Try to wrap message
	v, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	body, err := Message{err: msgErr, origin: mS.name, event: event, payload: v}.wrap()
	if err != nil {
		return err
	}

	// Publish message
	if err := mS.channel.publish(body, target); err != nil {
		return err
	}

	return nil
}

// Publish Message
func (mS *MessagingService) Publish(to, event string, msg any) error {
	return publish(mS, to, event, msg, nil)
}

// Bind handler
func (mS *MessagingService) On(event string, function func(m Message) (any, error)) {
	mS.handlers[event] = function
}

// Start consuming queues
func (mS *MessagingService) Consume() {
	forever := make(chan struct{})
	for i := 0; i < runtime.NumCPU(); i++ {
		go func(i int) {
			// Consume queue
			for msg := range mS.channel.consumeQueue(mS.name) {
				// Unwrap message
				m, err := unwrap(msg.Body)
				if err != nil {
					log.Println("couldn't unwrap received message", m)
					continue
				}
				// Call handler func
				handler, ok := mS.handlers[m.event]
				if !ok {
					log.Println("received event \"" + m.event + "\" has no assigned handler")
					continue
				}
				res, err := handler(m)

				// Return message to origin
				if res != nil || err != nil {
					publish(mS, m.origin, m.event, res, err)
				}

				// Acknowledge
				msg.Ack(false)
			}
		}(i)
	}
	<-forever
}

// Close connection
func (mS *MessagingService) Close() {
	mS.channel.close()
}
