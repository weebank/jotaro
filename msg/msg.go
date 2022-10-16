package msg

import (
	"encoding/json"
)

// Payload Object struct
type PayloadObject struct {
	content []byte
	isError bool
}

type EmptyPayloadObject struct{}

// Internal Payload Object struct
type payloadObject struct {
	Content []byte `json:"content"`
	IsError bool   `json:"isError"`
}

// Export fields
func (pO PayloadObject) exportFields() payloadObject {
	return payloadObject{Content: pO.content, IsError: pO.isError}
}

// Import fields
func (pO *PayloadObject) importFields(p payloadObject) {
	pO.content = p.Content
	pO.isError = p.IsError
}

// Unmarshal Payload Object
func NewPayloadObject(v any) (PayloadObject, error) {
	if content, errMarshal := json.Marshal(v); errMarshal != nil {
		return PayloadObject{}, errMarshal
	} else {
		_, isError := v.(error)
		return PayloadObject{content, isError}, nil
	}
}

// Extract Error From Payload Object
func (pO PayloadObject) AsError() error {
	if pO.isError {
		var err error
		pO.Bind(&err)
		return err
	}
	return nil
}

// Unmarshal Payload Object
func (pO PayloadObject) Bind(v any) error {
	return json.Unmarshal(pO.content, v)
}

// Message struct
type Message struct {
	origin  string
	event   string
	Payload map[string]PayloadObject
}

// Internal message struct
type message struct {
	Origin  string
	Event   string
	Payload map[string]payloadObject
}

// Bind Payload Object related to current event
func (m Message) CurrentPayload() (PayloadObject, bool) {
	pO, ok := m.Payload[m.event]
	return pO, ok
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
	// Ensure Payload is not nil
	if m.Payload == nil {
		m.Payload = make(map[string]PayloadObject)
	}

	// Export Payload Objects
	exportedPOs := make(map[string]payloadObject)
	for k, v := range m.Payload {
		exportedPOs[k] = v.exportFields()
	}

	// Export Message
	return message{Origin: m.origin, Event: m.event, Payload: exportedPOs}
}

// Import fields
func (m *Message) importFields(M message) {
	m.origin = M.Origin
	m.event = M.Event
	if M.Payload == nil {
		m.Payload = make(map[string]PayloadObject)
	} else {
		// Import Payload Objects
		importedPOs := make(map[string]PayloadObject)
		for k, v := range M.Payload {
			pO := PayloadObject{}
			pO.importFields(v)
			importedPOs[k] = pO
		}
		m.Payload = importedPOs
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
