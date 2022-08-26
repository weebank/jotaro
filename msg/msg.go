package msg

import (
	"encoding/json"
	"errors"
)

// Message struct
type Message struct {
	id      string
	err     error
	origin  string
	event   string
	target  string
	payload map[string][]byte
}

// Internal message struct
type message struct {
	ID      string
	Err     string
	Origin  string
	Event   string
	Payload map[string][]byte
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

// Prepare Message
func (m *Message) Prepare(event, target string) {
	m.event = event
	m.target = target
}

// Bind payload object related to current event
func (m Message) BindLatest(v any) error {
	return json.Unmarshal(m.payload[m.event], v)
}

// Bind payload object
func (m Message) Bind(k string, v any) error {
	return json.Unmarshal(m.payload[k], v)
}

// Export fields
func (m Message) exportFields() message {
	err := ""
	if m.err != nil {
		err = m.err.Error()
	}
	return message{ID: m.id, Err: err, Origin: m.origin, Event: m.event, Payload: m.payload}
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
