package msg

import (
	"encoding/json"
	"errors"
)

// Payload object
type payloadObject struct {
	Content []byte
	Err     error
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
func (m Message) BindPrevious(k string, v any) error {
	err := json.Unmarshal(m.payload[k].Content, v)
	if err != nil {
		return err
	}
	return m.payload[m.event].Err
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
