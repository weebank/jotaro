package msg

import (
	"encoding/json"
	"errors"
)

// Message interface
type forwardableMessage interface {
	toForwardableMessage() (*Message, error)
}

// Payload object
type payloadObject struct {
	Content []byte
	Err     error
}

// Initialize Payload Object
func newPayloadObject(v any, err error) (*payloadObject, error) {
	// Marshal content
	body, errMarshal := json.Marshal(v)
	if errMarshal != nil {
		return nil, err
	}

	return &payloadObject{Content: body, Err: err}, nil
}

// Message struct
type Message struct {
	id      string
	err     error
	origin  string
	event   string
	target  string
	payload map[string]payloadObject
}

// Message to Forwardable Message
func (m Message) toForwardableMessage() (*Message, error) {
	return &m, nil
}

// Message (with Payload) struct
type StatefulMessage struct {
	Event   string
	Content any
	Err     error
}

// Message to Forwardable Message
func (sM StatefulMessage) toForwardableMessage() (*Message, error) {
	pO, err := newPayloadObject(sM.Content, sM.Err)
	if err != nil {
		return nil, err
	}
	return &Message{event: sM.Event, payload: map[string]payloadObject{sM.Event: *pO}}, nil
}

// Internal message struct
type message struct {
	ID      string
	Err     string
	Origin  string
	Event   string
	Payload map[string]payloadObject
}

// Get ID
func (m Message) ID() string {
	return m.id
}

// Get Origin
func (m Message) Origin() string {
	return m.origin
}

// Get Event
func (m Message) Event() string {
	return m.event
}

// Forward Message
func (m *Message) Forward(event, target string) {
	m.event = event
	m.target = target
}

// Bind payload object related to current event
func (m Message) Bind(v any) error {
	return m.BindPrevious(m.event, v)
}

// Bind payload object
func (m Message) BindPrevious(event string, v any) error {
	err := json.Unmarshal(m.payload[event].Content, v)
	if err != nil {
		return err
	}
	return m.payload[event].Err
}

// Export fields
func (m Message) exportFields() message {
	return message{ID: m.id, Origin: m.origin, Event: m.event, Payload: m.payload}
}

// Import fields
func (m *Message) importFields(M message) {
	m.id = M.ID
	m.err = errors.New(M.Err)
	m.origin = M.Origin
	m.event = M.Event
	m.payload = M.Payload
}

// Wrap Message
func (m Message) wrap() ([]byte, error) {
	v, err := json.Marshal(m.exportFields())
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Unwrap Message
func unwrap(body []byte) (m Message, err error) {
	M := message{}
	err = json.Unmarshal(body, &M)
	if err != nil {
		return
	}
	m.importFields(M)
	return
}

// Build response event
func ResponseEvent(event, service string) string {
	return event + "-" + service
}
