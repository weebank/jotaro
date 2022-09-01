package msg

import (
	"encoding/json"
	"errors"
)

// Payload object
type PayloadObject struct {
	content []byte
	Err     error
}

// Unmarshal Payload Object
func BuildPayloadObject(v any, err error) (PayloadObject, error) {
	if content, errMarshal := json.Marshal(v); errMarshal != nil {
		return PayloadObject{}, errMarshal
	} else {
		return PayloadObject{content, err}, nil
	}
}

// Unmarshal Payload Object
func (pO PayloadObject) Bind(v any) error {
	return json.Unmarshal(pO.content, v)
}

// Message struct
type Message struct {
	id      string
	err     error
	origin  string
	event   string
	Payload map[string]PayloadObject
}

// Internal message struct
type message struct {
	ID      string
	Err     string
	Origin  string
	Event   string
	Payload map[string]PayloadObject
}

// Bind Payload Object related to current event
func (m Message) CurrentPayloadObject() (PayloadObject, bool) {
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
	m.err = errors.New(M.Err)
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
