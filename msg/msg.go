package msg

import (
	"encoding/json"
)

type PayloadObject []byte

type EmptyPayloadObject struct{}

// Unmarshal Payload Object
func NewPayloadObject(v any) (PayloadObject, error) {
	if content, errMarshal := json.Marshal(v); errMarshal != nil {
		return make([]byte, 0), errMarshal
	} else {
		return content, nil
	}
}

// Unmarshal Payload Object
func (pO PayloadObject) Bind(v any) error {
	return json.Unmarshal(pO, v)
}

// Message struct
type Message struct {
	id      string
	origin  string
	event   string
	Payload map[string]PayloadObject
}

// Internal message struct
type message struct {
	ID      string
	Origin  string
	Event   string
	Payload map[string]PayloadObject
}

// Bind Payload Object related to current event
func (m Message) CurrentPayload() (PayloadObject, bool) {
	pO, ok := m.Payload[m.event]
	return pO, ok
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

// Export fields
func (m Message) exportFields() message {
	if m.Payload == nil {
		m.Payload = make(map[string]PayloadObject)
	}
	return message{ID: m.id, Origin: m.origin, Event: m.event, Payload: m.Payload}
}

// Import fields
func (m *Message) importFields(M message) {
	m.id = M.ID
	m.origin = M.Origin
	m.event = M.Event
	if M.Payload == nil {
		m.Payload = make(map[string]PayloadObject)
	} else {
		m.Payload = M.Payload
	}
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
func ResponseEvent(event string) string {
	return event + "-" + "response"
}
