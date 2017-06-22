package message

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
)

// Message represents RDSS messages.
type Message struct {
	Header Headers

	// Body carries the message payload. It uses ``
	// Find them in message_metadata.go, message_term.go, etc...
	Body interface{}
}

// Headers contains the message headers.
type Headers struct {
	ID            string `json:"messageId"`
	Class         Class  `json:"messageClass"`
	Type          Type   `json:"messageType"`
	CorrelationID string `json:"correlationId,omitempty"`
}

// New returns a pointer to a new message with a new ID.
func New(t Type, c Class) *Message {
	return &Message{
		Header: Headers{
			ID:    genID(),
			Type:  t,
			Class: c,
		},
		Body: typedBody(t, ""),
	}
}

// messageAlias is proxy type for Message. Using json.RawMessgae in order to:
// - Delay JSON decoding.
// - Precompute JSON encoding.
type messageAlias struct {
	Header Headers         `json:"messageHeader"`
	Body   json.RawMessage `json:"messageBody"`
}

// MarshalJSON implements Marshaler.
func (m *Message) MarshalJSON() ([]byte, error) {
	body, err := json.Marshal(m.Body)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&messageAlias{m.Header, json.RawMessage(body)})
}

// UnmarshalJSON implements Unmarshaler.
func (m *Message) UnmarshalJSON(data []byte) error {
	msg := messageAlias{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	m.Header = msg.Header
	m.Body = typedBody(m.Header.Type, m.Header.CorrelationID)
	return json.Unmarshal(msg.Body, m.Body)
}

// typedBody returns an interface{} type where the type of the underlying value
// is chosen after the header message type.
func typedBody(t Type, correlationID string) interface{} {
	var body interface{}
	switch {
	case t == TypeMetadataCreate:
		body = new(MetadataCreateRequest)
	case t == TypeMetadataRead:
		if correlationID == "" {
			body = new(MetadataReadRequest)
		} else {
			body = new(MetadataReadResponse)
		}
	case t == TypeMetadataUpdate:
		body = new(MetadataUpdateRequest)
	case t == TypeMetadataDelete:
		body = new(MetadataDeleteRequest)
	}
	return body
}

func genID() string {
	id := make([]byte, 16)
	io.ReadFull(rand.Reader, id)
	return fmt.Sprintf("%x", id)
}
