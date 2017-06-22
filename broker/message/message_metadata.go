package message

import (
	"fmt"
)

// Metadata Create

// MetadataCreateRequest represents the body of the message.
type MetadataCreateRequest struct {
	// TODO: Embed Dataset
	UUID         string              `json:"objectUUID"`
	Title        string              `json:"objectTitle"`
	Contributors []*PersonRole       `json:"objectContributor"`
	Description  string              `json:"objectDescription"`
	Dates        []*Date             `json:"objectDate"`
	ResourceType string              `json:"objectResourceType"`
	Identifiers  []*Identifier       `json:"objectIdentifier"`
	Publishers   []*OrganisationRole `json:"objectPublisher"`
	Files        []*File             `json:"objectFile,omitempty"`
}

type Identifier struct {
	Value string `json:"identifierValue"`
	Type  string `json:"identifierType"`
}

type Date struct {
	Value string `json:"dateValue"`
	Type  string `json:"dateType"`
}

type OrganisationRole struct {
	Organisation *Organisation `json:"Organisation"`
	Role         string        `json:"Role"`
}

type Organisation struct {
	Name    string `json:"organisationName"`
	Address string `json:"organisationAddress"`
}

type PersonRole struct {
	Person *Person `json:"Person"`
	Role   string  `json:"Role"`
}

type Person struct {
	UUID      string `json:"personUUID"`
	GivenName string `json:"personGivenName"`
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
	UUID        string  `json:"objectUuid"`
	Title       string  `json:"objectTitle"`
	Description string  `json:"objectDescription"`
	Files       []*File `json:"objectFile"`
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
	UUID        string  `json:"objectUuid"`
	Title       string  `json:"objectTitle"`
	Description string  `json:"objectDescription"`
	Files       []*File `json:"objectFile"`
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

type File struct {
	UUID            string     `json:"fileUUID"`
	Identifier      string     `json:"fileIdentifier"`
	Name            string     `json:"fileName"`
	Size            int        `json:"fileSize"`
	Label           string     `json:"fileLabel,omitempty"`
	Checksums       []Checksum `json:"fileChecksum"`
	FormatType      string     `json:"fileFormatType,omitempty"`
	HasMimeType     bool       `json:"filehasMimeType,omitempty"`
	StorageLocation string     `json:"fileStorageLocation"`
	StorageType     string     `json:"fileStorageType"`
}

type Checksum struct {
	Type  string `json:"checksumType"`
	Value string `json:"checksumValue"`
}
