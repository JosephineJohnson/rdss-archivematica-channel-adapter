package kinesis

import (
	"errors"
	"testing"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	log "github.com/Sirupsen/logrus"
)

func Test_processor_route_ErrorHandling(t *testing.T) {
	tc := []struct {
		handlers []backend.Handler
		expErr   bool
	}{
		{
			handlers: []backend.Handler{
				func([]byte) error { return nil },
				func([]byte) error { return nil },
			},
			expErr: false,
		},
		{
			handlers: []backend.Handler{
				func([]byte) error { return errors.New("something really bad happened") },
				func([]byte) error { return nil },
			},
			expErr: true,
		},
	}
	for _, tt := range tc {
		p, _ := newProcessor(&BackendImpl{logger: log.New()}, "stream")
		p.handlers = tt.handlers
		err := p.route([]byte("{}"))
		if tt.expErr && err == nil {
			t.Fatalf("processor.route() returned unexpected nil error value")
		}
	}
}
