// Package broker implements a low-level interface for communicating with RDSS
// via message brokers such Amazon Kinesis Stream or RabbitMQ.
//
// The goal of this package is to be reusable by any user. It's not coupled to
// Archivematica.
package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
)

type Config struct {
	QueueMain    string
	QueueInvalid string
	QueueError   string
}

func (c *Config) Validate() error {
	if c.QueueMain == "" {
		return errors.New("main queue name is undefined")
	}
	if c.QueueInvalid == "" {
		return errors.New("invalid queue name is undefined")
	}
	if c.QueueError == "" {
		return errors.New("error queue name is undefined")
	}
	return nil
}

// Broker is a broker client conforming to the RDSS messaging API.
type Broker struct {
	backend backend.Backend
	logger  log.FieldLogger
	config  *Config

	Metadata MetadataService
	// TODO: Vocabulary VocabularyService
	// TODO: Term       TermService

	// Number of messages received.
	Count uint64

	// List of subscribers
	subs []subscription
	mu   sync.RWMutex
}

// subscription is a subscription with a handler to a particular topic. If the
// value of all is true then the handler will be invoked for every message.
type subscription struct {
	// Whether the subscriber wants to be subscribed to receive every message.
	all bool

	// The type of message that this subscriber is listening.
	mType message.Type

	// The callback associated to this particular subscriber.
	cb MessageHandler
}

// MessageHandler is a callback function supplied by subscribers.
type MessageHandler func(msg *message.Message) error

// New returns a new Broker.
func New(backend backend.Backend, logger log.FieldLogger, config *Config) (*Broker, error) {
	if logger == nil {
		logger = log.New()
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("Configuration error: %s", err)
	}

	b := &Broker{
		backend: backend,
		logger:  logger,
		config:  config,
	}
	b.Metadata = &MetadataServiceOp{broker: b}

	// Check queues
	if err := b.checkQueues(); err != nil {
		return nil, err
	}

	// Set up message router
	b.backend.Subscribe(config.QueueMain, b.messageHandler)

	return b, nil
}

// messageHandler is
func (b *Broker) messageHandler(data []byte) error {
	// Unmarshal the message.
	msg := &message.Message{}
	err := json.Unmarshal(data, msg)
	if err != nil {
		b.logger.Errorln("Message received but unmarshalling failed:", err)
		return err
	}

	atomic.AddUint64(&b.Count, 1)
	b.logger.WithFields(log.Fields{"count": b.Count, "type": msg.Header.Type}).Infoln("Message received")

	// Dispatch the message to the subscribers asynchronously.
	var wg sync.WaitGroup
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, s := range b.subs {
		if !s.all && msg.Header.Type != s.mType {
			continue
		}
		wg.Add(1)
		// Here is our goroutine that is going to dispatch the message to each
		// subscriber.
		go func(s subscription) {
			defer wg.Done()
			// We want to regain control of a panicking goroutine if that was
			// the case. The idea is not to let the application crash if the
			// user making use of this package implemented a faulty handler.
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("recovered from panic in subscriber: %s", err)
				}
			}()
			cbErr := s.cb(msg)
			if cbErr != nil {
				err = fmt.Errorf("handler returned error: %s", cbErr)
			}
		}(s)
	}

	// Wait until all the goroutines are done
	wg.Wait()

	return err
}

// checkQueues verifies access and availability of the queues being used.
func (b *Broker) checkQueues() error {
	var err error

	if err = b.backend.Check(b.config.QueueMain); err != nil {
		return err
	}
	b.logger.WithField("queue", "main").Debugln("Queue check succeeded:", b.config.QueueMain)

	if err = b.backend.Check(b.config.QueueError); err != nil {
		return err
	}
	b.logger.WithField("queue", "error").Debugln("Queue check succeeded:", b.config.QueueError)

	if err = b.backend.Check(b.config.QueueInvalid); err != nil {
		return err
	}
	b.logger.WithField("queue", "invalid").Debugln("Queue check succeeded:", b.config.QueueInvalid)

	return nil
}

// Subscribe creates a new subscription associated to every message received.
func (b *Broker) Subscribe(cb MessageHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, subscription{all: true, cb: cb})
}

// SubscribeType creates a new subscription associated to a particular message type.
func (b *Broker) SubscribeType(t message.Type, cb MessageHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subs = append(b.subs, subscription{mType: t, cb: cb})
}

// Request sends a fire-and-forget request to RDSS.
func (b *Broker) Request(ctx context.Context, msg *message.Message) error {
	return errors.New("not implemented yet")
}

// RequestResponse sends a request and waits until a response is received.
func (b *Broker) RequestResponse(context.Context, *message.Message) (*message.Message, error) {
	return nil, errors.New("not implemented yet")
}
