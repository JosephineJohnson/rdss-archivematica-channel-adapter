package backendmock

import (
	"sync"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
)

// New returns a backendmock backend.
func New(opts *backend.Opts) (backend.Backend, error) {
	b := &BackendImpl{}

	return b, nil
}

func init() {
	backend.Register("backendmock", New)
}

// BackendImpl is a mock implementation of broker.Backend. It's not safe to use
// from multiple goroutines (concurrent map access going on right now).
type BackendImpl struct {
	Subscriptions []subscription
	mu            sync.RWMutex
}

type subscription struct {
	topic string
	cb    backend.Handler
}

var _ backend.Backend = (*BackendImpl)(nil)

func (b *BackendImpl) Publish(topic string, data []byte) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, s := range b.Subscriptions {
		if topic != "" && s.topic != topic {
			continue
		}
		s.cb(data)
	}
	return nil
}

func (b *BackendImpl) Subscribe(topic string, cb backend.Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subscription := subscription{topic: topic, cb: cb}
	b.Subscriptions = append(b.Subscriptions, subscription)
}

func (b *BackendImpl) Check(topic string) error {
	return nil
}

// Close implements broker.Backend
func (b *BackendImpl) Close() error {
	return nil
}
