package broker

import (
	"context"
	"errors"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

// MetadataService TODO! These methods need to be defined.
type MetadataService interface {
	Create(context.Context, *message.MetadataCreateRequest) error
	Read(context.Context, *message.MetadataReadRequest) (*message.MetadataReadResponse, error)
	Update(context.Context, *message.MetadataUpdateRequest) error
	Delete(context.Context, *message.MetadataDeleteRequest) error
}

type MetadataServiceOp struct {
	broker *Broker
}

// Create implements MetadataService
func (s *MetadataServiceOp) Create(ctx context.Context, req *message.MetadataCreateRequest) error {
	msg := message.New(message.TypeMetadataCreate, message.ClassCommand)
	msg.Body = req

	return s.broker.Request(ctx, msg)
}

// Read implements MetadataService
func (s *MetadataServiceOp) Read(ctx context.Context, req *message.MetadataReadRequest) (*message.MetadataReadResponse, error) {
	msg := message.New(message.TypeMetadataCreate, message.ClassCommand)
	msg.Body = req

	resp, err := s.broker.RequestResponse(ctx, msg)
	r, ok := resp.Body.(*message.MetadataReadResponse)
	if !ok {
		return nil, errors.New("unexpected")
	}

	return r, err
}

// Update implements MetadataService
func (s *MetadataServiceOp) Update(ctx context.Context, req *message.MetadataUpdateRequest) error {
	msg := message.New(message.TypeMetadataCreate, message.ClassCommand)
	msg.Body = req

	return s.broker.Request(ctx, msg)
}

// Delete implements MetadataService
func (s *MetadataServiceOp) Delete(ctx context.Context, req *message.MetadataDeleteRequest) error {
	msg := message.New(message.TypeMetadataCreate, message.ClassCommand)
	msg.Body = req

	return s.broker.Request(ctx, msg)
}
