package message

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Class (messageClass), one of Command, Event or Document.
type Class int

const (
	ClassCommand Class = iota
	ClassEvent
	ClassDocument
)

var (
	classMap    map[string]Class
	classMapInv map[Class]string
)

func init() {
	classMap = map[string]Class{
		"Command":  ClassCommand,
		"Event":    ClassEvent,
		"Document": ClassDocument,
	}

	classMapInv = make(map[Class]string)
	for k, v := range classMap {
		classMapInv[v] = k
	}
}

func (c Class) String() string {
	v, ok := classMapInv[c]
	if !ok {
		return "Unknown"
	}
	return v
}

// MarshalJSON implements Marshaler
func (c Class) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%s", c)
	return json.Marshal(s)
}

// UnmarshalJSON implements unmarshaler
func (c *Class) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	have, ok := classMap[s]
	if !ok {
		return fmt.Errorf("invalid message class %q", s)
	}
	*c = have
	return nil
}

// Type (messageType) describes the type of Message.
type Type int

const (
	TypeMetadataCreate Type = iota
	TypeMetadataRead
	TypeMetadataUpdate
	TypeMetadataDelete
)

var (
	typeMap    map[string]Type
	typeMapInv map[Type]string
)

func init() {
	typeMap = map[string]Type{
		"MetadataCreate": TypeMetadataCreate,
		"MetadataRead":   TypeMetadataRead,
		"MetadataUpdate": TypeMetadataUpdate,
		"MetadataDelete": TypeMetadataDelete,
	}

	typeMapInv = make(map[Type]string)
	for k, v := range typeMap {
		typeMapInv[v] = k
	}
}

func (t Type) String() string {
	v, ok := typeMapInv[t]
	if !ok {
		return "Unknown"
	}
	return v
}

// MarshalJSON implements Marshaler
func (t Type) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("%s", t)
	s = strings.TrimSuffix(s, "Request")
	s = strings.TrimSuffix(s, "Response")
	return json.Marshal(s)
}

// UnmarshalJSON implements Unmarshaler
func (t *Type) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	v, ok := typeMap[s]
	if !ok {
		return fmt.Errorf("invalid message type %q", s)
	}
	*t = v
	return nil
}
