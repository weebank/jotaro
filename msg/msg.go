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
	payload map[string]any
	content []byte
}

// Internal message struct
type message struct {
	ID      string
	Error   string
	Origin  string
	Event   string
	Payload map[string]any
	Content []byte
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

// Get Payload
func (m Message) Payload() map[string]any {
	M := make(map[string]any)
	for k, v := range m.payload {
		M[k] = v
	}
	return M
}

// Get Content
func (m Message) Content() []byte {
	return m.content
}

// Unwrap JSON Payload
func (m Message) Bind(v any) error {
	return json.Unmarshal(m.content, v)
}

// Export fields
func (m Message) exportFields() message {
	err := ""
	if m.err != nil {
		err = m.err.Error()
	}
	return message{ID: m.id, Error: err, Origin: m.origin, Event: m.event, Payload: m.payload}
}

// Import fields
func (m *Message) importFields(M message) {
	m.id = M.ID
	m.err = errors.New(M.Error)
	m.origin = M.Origin
	m.event = M.Event
	m.content = M.Content
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
	m.importFields(M)
	return
}
