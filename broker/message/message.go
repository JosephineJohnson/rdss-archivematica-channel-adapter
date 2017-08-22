package message

import (
	"encoding/json"

	"github.com/twinj/uuid"

	bErrors "github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/errors"
)

// Message represents RDSS messages.
type Message struct {
	MessageHeader MessageHeader

	// MessageBody carries the message payload.
	MessageBody interface{}
}

// New returns a pointer to a new message with a new ID.
func New(t MessageType, c MessageClass) *Message {
	return &Message{
		MessageHeader: MessageHeader{
			ID:           uuid.NewV4().String(),
			MessageType:  t,
			MessageClass: c,
			Version:      Version,
		},
		MessageBody: typedBody(t, ""),
	}
}

// messageAlias is proxy type for Message. Using json.RawMessage in order to:
// - Delay JSON decoding.
// - Precompute JSON encoding.
type messageAlias struct {
	MessageHeader MessageHeader   `json:"messageHeader"`
	MessageBody   json.RawMessage `json:"messageBody"`
}

func (m *Message) ID() string {
	return m.MessageHeader.ID
}

func (m *Message) TagError(err error) {
	if err == nil {
		return
	}
	e, ok := err.(*bErrors.Error)
	if ok && e != nil {
		m.MessageHeader.ErrorCode = e.Kind.String()
		m.MessageHeader.ErrorDescription = e.Err.Error()
	} else if !ok && err != nil {
		m.MessageHeader.ErrorCode = "Unknown"
		m.MessageHeader.ErrorDescription = err.Error()
	}
}

// MarshalJSON implements Marshaler.
func (m *Message) MarshalJSON() ([]byte, error) {
	body, err := json.Marshal(m.MessageBody)
	if err != nil {
		return nil, err
	}
	return json.Marshal(&messageAlias{m.MessageHeader, json.RawMessage(body)})
}

// UnmarshalJSON implements Unmarshaler.
func (m *Message) UnmarshalJSON(data []byte) error {
	msg := messageAlias{}
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}
	m.MessageHeader = msg.MessageHeader
	m.MessageBody = typedBody(m.MessageHeader.MessageType, m.MessageHeader.CorrelationID)
	return json.Unmarshal(msg.MessageBody, m.MessageBody)
}

// typedBody returns an interface{} type where the type of the underlying value
// is chosen after the header message type.
func typedBody(t MessageType, correlationID string) interface{} {
	var body interface{}
	switch {
	case t == MessageTypeMetadataCreate:
		body = new(MetadataCreateRequest)
	case t == MessageTypeMetadataRead:
		if correlationID == "" {
			body = new(MetadataReadRequest)
		} else {
			body = new(MetadataReadResponse)
		}
	case t == MessageTypeMetadataUpdate:
		body = new(MetadataUpdateRequest)
	case t == MessageTypeMetadataDelete:
		body = new(MetadataDeleteRequest)
	case t == MessageTypeVocabularyRead:
		if correlationID == "" {
			body = new(VocabularyReadRequest)
		} else {
			body = new(VocabularyReadResponse)
		}
	case t == MessageTypeVocabularyPatch:
		body = new(VocabularyPatchRequest)
	}
	return body
}
