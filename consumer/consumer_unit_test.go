package consumer

import (
	"context"
	"testing"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	logt "github.com/sirupsen/logrus/hooks/test"
)

func Test_handleMetadataCreateRequest_errMessageType(t *testing.T) {
	c, _ := getConsumer(t)
	if err := c.handleMetadataCreateRequest(&message.Message{}); err == nil {
		t.Fatal("Expected non-nil error, got nil")
	}
}
func Test_handleMetadataCreateRequest_emptyFiles(t *testing.T) {
	c, _ := getConsumer(t)
	msg := message.New(message.MessageTypeMetadataCreate, message.MessageClassCommand)
	if err := c.handleMetadataCreateRequest(msg); err != nil {
		t.Fatalf("Expected nil error, got %s", err)
	}
}

func Test_getFilename(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"https://foo12345.com/foobar.jpg", "foobar.jpg"},
		{"https://foo12345.com/foo/bar/rrr", "foo/bar/rrr"},
		{":invalid-url:", ""},
	}
	for _, tt := range tests {
		if got := getFilename(tt.path); tt.want != got {
			t.Errorf("getFilename(); want %v, got %v", tt.want, got)
		}
	}
}

func getConsumer(t *testing.T) (*ConsumerImpl, *logt.Hook) {
	logger, hook := logt.NewNullLogger()
	c := &ConsumerImpl{
		ctx:    context.Background(),
		logger: logger,
	}
	return c, hook
}
