package broker

import (
	"context"
	"errors"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

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
	msg := message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand)
	msg.MessageBody = req

	return s.broker.Request(ctx, msg)
}

// Read implements MetadataService
func (s *MetadataServiceOp) Read(ctx context.Context, req *message.MetadataReadRequest) (*message.MetadataReadResponse, error) {
	msg := message.New(message.MessageTypeMetadataRead, message.MessageClassCommand)
	msg.MessageBody = req

	resp, err := s.broker.RequestResponse(ctx, msg)
	r, ok := resp.MessageBody.(*message.MetadataReadResponse)
	if !ok {
		return nil, errors.New("unexpected")
	}

	return r, err
}

// Update implements MetadataService
func (s *MetadataServiceOp) Update(ctx context.Context, req *message.MetadataUpdateRequest) error {
	msg := message.New(message.MessageTypeMetadataUpdate, message.MessageClassCommand)
	msg.MessageBody = req

	return s.broker.Request(ctx, msg)
}

// Delete implements MetadataService
func (s *MetadataServiceOp) Delete(ctx context.Context, req *message.MetadataDeleteRequest) error {
	msg := message.New(message.MessageTypeMetadataDelete, message.MessageClassCommand)
	msg.MessageBody = req

	return s.broker.Request(ctx, msg)
}
