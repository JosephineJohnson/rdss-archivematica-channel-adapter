package publisher

import (
	"golang.org/x/net/context"

	log "github.com/sirupsen/logrus"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/publisher/pb"
)

func init() {
	// I believe this would be a great palce to replace the gRPC logger with
	// our own logger. But we need first a logger, one that meets their
	// interface, which should not be a big deal. See for example how this is
	// solved in github.com/docker/swarmkit/log, although they do something
	// I'm not sure I like: passing loggers via contexts - is that bad?
}

type RdssServer struct {
	broker *broker.Broker
	logger log.FieldLogger
}

func MakeRdssServer(b *broker.Broker, l log.FieldLogger) pb.RdssServer {
	return &RdssServer{broker: b, logger: l}
}

func (s *RdssServer) MetadataCreate(ctx context.Context, req *pb.MetadataCreateRequest) (*pb.MetadataCreateResponse, error) {
	if err := s.broker.Metadata.Create(ctx, &message.MetadataCreateRequest{}); err != nil {
		s.logger.Error(err)
	}

	return &pb.MetadataCreateResponse{}, nil
}

func (s *RdssServer) MetadataRead(ctx context.Context, req *pb.MetadataReadRequest) (*pb.MetadataReadResponse, error) {
	if _, err := s.broker.Metadata.Read(ctx, &message.MetadataReadRequest{}); err != nil {
		s.logger.Error(err)
	}

	return &pb.MetadataReadResponse{}, nil
}

func (s *RdssServer) MetadataUpdate(ctx context.Context, req *pb.MetadataUpdateRequest) (*pb.MetadataUpdateResponse, error) {
	if err := s.broker.Metadata.Update(ctx, &message.MetadataUpdateRequest{}); err != nil {
		s.logger.Error(err)
	}

	return &pb.MetadataUpdateResponse{}, nil
}

func (s *RdssServer) MetadataDelete(ctx context.Context, req *pb.MetadataDeleteRequest) (*pb.MetadataDeleteResponse, error) {
	if err := s.broker.Metadata.Delete(ctx, &message.MetadataDeleteRequest{}); err != nil {
		s.logger.Error(err)
	}

	return &pb.MetadataDeleteResponse{}, nil
}
