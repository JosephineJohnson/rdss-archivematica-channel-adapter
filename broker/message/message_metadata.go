package message

import (
	"fmt"
)

// Metadata Create

// MetadataCreateRequest represents the body of the message.
type MetadataCreateRequest struct {
	// TODO: Embed Dataset
	UUID        string          `json:"objectUuid"`
	Title       string          `json:"objectTitle"`
	Description string          `json:"objectDescription"`
	Files       []*MetadataFile `json:"objectFile"`
}

// MetadataCreateRequest returns the body of the message.
func (m Message) MetadataCreateRequest() (*MetadataCreateRequest, error) {
	body, ok := m.Body.(*MetadataCreateRequest)
	if !ok {
		return nil, fmt.Errorf("interface conversion error")
	}
	return body, nil
}

// Metadata Read

// MetadataReadRequest represents the body of the message.
type MetadataReadRequest struct {
	UUID string `json:"objectUuid"`
}

// MetadataReadRequest returns the body of the message.
func (m Message) MetadataReadRequest() (*MetadataReadRequest, error) {
	b, ok := m.Body.(*MetadataReadRequest)
	if !ok {
		return nil, fmt.Errorf("interface conversion error")
	}
	return b, nil
}

// MetadataReadResponse represents the body of the message.
type MetadataReadResponse struct {
	// TODO: Embed Dataset
	UUID        string          `json:"objectUuid"`
	Title       string          `json:"objectTitle"`
	Description string          `json:"objectDescription"`
	Files       []*MetadataFile `json:"objectFile"`
}

// MetadataReadResponse returns the body of the message.
func (m Message) MetadataReadResponse() (*MetadataReadResponse, error) {
	b, ok := m.Body.(*MetadataReadResponse)
	if !ok {
		return nil, fmt.Errorf("interface conversion error")
	}
	return b, nil
}

// Metadata Update

// MetadataUpdateRequest represents the body of the message.
type MetadataUpdateRequest struct {
	// TODO: Embed Dataset
	UUID        string          `json:"objectUuid"`
	Title       string          `json:"objectTitle"`
	Description string          `json:"objectDescription"`
	Files       []*MetadataFile `json:"objectFile"`
}

// MetadataUpdateRequest returns the body of the message.
func (m Message) MetadataUpdateRequest() (*MetadataUpdateRequest, error) {
	b, ok := m.Body.(*MetadataUpdateRequest)
	if !ok {
		return nil, fmt.Errorf("interface conversion error")
	}
	return b, nil
}

// Metadata Delete

// MetadataDeleteRequest represents the body of the message.
type MetadataDeleteRequest struct {
	UUID string `json:"objectUuid"`
}

// MetadataDeleteRequest returns the body of the message.
func (m Message) MetadataDeleteRequest() (*MetadataDeleteRequest, error) {
	b, ok := m.Body.(*MetadataDeleteRequest)
	if !ok {
		return nil, fmt.Errorf("interface conversion error")
	}
	return b, nil
}

// Subtypes

type MetadataFile struct {
	UUID            string                 `json:"fileUuid"`
	Identifier      string                 `json:"fileIdentifier"`
	Name            string                 `json:"fileName"`
	Size            int                    `json:"fileSize"`
	Checksums       []MetadataFileChecksum `json:"fileChecksum"`
	Label           string                 `json:"fileLabel"`
	HasMimeType     bool                   `json:"fileHasMimeType"`
	FormatType      string                 `json:"fileFormatType"`
	StorageLocation string                 `json:"fileStorageLocation"`
	StorageType     string                 `json:"fileStorageType"`
}

type MetadataFileChecksum struct {
	UUID  string `json:"checksumUuid"`
	Type  string `json:"checksumType"`
	Value string `json:"checksumValue"`
}
