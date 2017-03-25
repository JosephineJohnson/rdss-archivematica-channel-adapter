package outbound

import (
	"log"

	"golang.org/x/net/context"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/outbound/pb"
)

func init() {
	// I believe this would be a great palce to replace the gRPC logger with
	// our own logger. But we need first a logger, one that meets their
	// interface, which should not be a big deal. See for example how this is
	// solved in github.com/docker/swarmkit/log, although they do something
	// I'm not sure I like: passing loggers via contexts - is that bad?
}

type RdssServer struct {
	broker broker.Backend
}

func MakeRdssServer(broker broker.Backend) pb.RdssServer {
	return &RdssServer{broker: broker}
}

func (s *RdssServer) MetadataRead(context.Context, *pb.MetadataReadRequest) (*pb.MetadataReadResponse, error) {
	msg := &broker.Message{}
	err := s.broker.Request(context.TODO(), msg)
	log.Println(err)
	return &pb.MetadataReadResponse{}, nil
}

func (s *RdssServer) MetadataCreate(context.Context, *pb.MetadataCreateRequest) (*pb.MetadataCreateResponse, error) {
	msg := &broker.Message{}
	err := s.broker.Request(context.TODO(), msg)
	log.Println(err)
	return &pb.MetadataCreateResponse{}, nil
}

func (s *RdssServer) MetadataUpdate(context.Context, *pb.MetadataUpdateRequest) (*pb.MetadataUpdateResponse, error) {
	msg := &broker.Message{}
	err := s.broker.Request(context.TODO(), msg)
	log.Println(err)
	return &pb.MetadataUpdateResponse{}, nil
}

func (s *RdssServer) MetadataDelete(context.Context, *pb.MetadataDeleteRequest) (*pb.MetadataDeleteResponse, error) {
	msg := &broker.Message{}
	err := s.broker.Request(context.TODO(), msg)
	log.Println(err)
	return &pb.MetadataDeleteResponse{}, nil
}
