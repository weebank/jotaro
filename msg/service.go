package msg

import (
	"errors"
	"runtime"
)

// Messaging Service
type MessagingService struct {
	name      string
	channel   Channel
	handlers  map[string]func(m Message) any
	DebugMode bool
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
	mS.handlers = make(map[string]func(m Message) any)

	return
}

// Publish Message Internal
func publish(mS *MessagingService, target, event string, payload map[string]PayloadObject) error {
	// Create message
	msg := Message{origin: mS.name, event: event, Payload: payload}
	body, err := msg.wrap()
	if err != nil {
		return err
	}

	// Show message if in debug mode
	if mS.DebugMode {
		logger.WithField("message", msg).Info("published")
	}

	// Publish message
	if err := mS.channel.publish(body, target); err != nil {
		return err
	}

	return nil
}

// Publish Message
func (mS *MessagingService) Publish(base Message, target, event string) error {
	// Check validity of event and target
	if event == "" {
		return errors.New("\"event\" cannot be empty")
	}
	if target == "" {
		return errors.New("\"target\" cannot be empty")
	}

	return publish(mS, target, event, base.Payload)
}

// Bind handler
func (mS *MessagingService) On(event string, function func(m Message) any) {
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
					logger.WithError(err).WithField("message", m).Error("couldn't unwrap received message")
					continue
				}

				// Show message if in debug mode
				if mS.DebugMode {
					logger.WithField("message", m).Info("received")
				}

				// Call handler func
				handler, ok := mS.handlers[m.event]
				if !ok {
					logger.WithField("event", m.event).Error("received event has no assigned handler")
					continue
				}

				// Store event and call handler
				if response := handler(m); response != nil {
					// Create response event name
					event := ResponseEvent(m.event)
					// Build payload object
					m.Payload[event], _ = NewPayloadObject(response)
					// Publish response
					mS.Publish(m, m.Origin(), event)
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
