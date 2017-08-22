package broker

import (
	"context"
	"errors"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

type VocabularyService interface {
	Read(context.Context, *message.VocabularyReadRequest) (*message.VocabularyReadResponse, error)
	Patch(context.Context, *message.VocabularyPatchRequest) error
}

type VocabularyServiceOp struct {
	broker *Broker
}

// Read implements VocabularyService
func (s *VocabularyServiceOp) Read(ctx context.Context, req *message.VocabularyReadRequest) (*message.VocabularyReadResponse, error) {
	msg := message.New(message.MessageTypeVocabularyRead, message.MessageClassCommand)
	msg.MessageBody = req

	resp, err := s.broker.RequestResponse(ctx, msg)
	r, ok := resp.MessageBody.(*message.VocabularyReadResponse)
	if !ok {
		return nil, errors.New("unexpected")
	}

	return r, err
}

// Patch implements VocabularyService
func (s *VocabularyServiceOp) Patch(ctx context.Context, req *message.VocabularyPatchRequest) error {
	msg := message.New(message.MessageTypeVocabularyPatch, message.MessageClassCommand)
	msg.MessageBody = req

	return s.broker.Request(ctx, msg)
}
