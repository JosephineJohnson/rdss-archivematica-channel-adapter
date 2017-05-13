package kinesis

import (
	"testing"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
)

var (
	s *session.Session
	b *BackendImpl
)

func tearUp(t *testing.T) {
	s = session.New()
	bc, _ := New(&backend.Opts{})
	b = bc.(*BackendImpl)
	b.kinesis = &kmock{}
}

func tearDown(t *testing.T) {
	b.Close()
}

func Test1(t *testing.T) {
	tearUp(t)
	defer tearDown(t)

	var output []byte
	b.Subscribe("foobar", func(data []byte) error {
		output = data
		return nil
	})

	// Once we have the processor implemented (see processor.go WIP), we can
	// introduce our kmock (see below) to generate the record we want.

	t.Skip("TODO: not ready yet!")
}

type kmock struct {
	kinesisiface.KinesisAPI
}
