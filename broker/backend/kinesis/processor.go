package kinesis

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/twitchscience/kinsumer"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
)

// processor processes the data records from a Kinesis stream. In Kinesis, each
// consumer reads from a particular shard, using a shard iterator. This
// processor attempts to process all the shards available.
type processor struct {
	logger  log.FieldLogger
	backend *BackendImpl
	stream  string
	quit    chan struct{}
	closed  bool

	handlers []backend.Handler
	mu       sync.RWMutex

	kinsumer *kinsumer.Kinsumer
}

const freq = 10 * time.Second

func newProcessor(backend *BackendImpl, stream string) (p *processor, err error) {
	p = &processor{
		logger:  backend.logger.WithField("stream", stream).WithField("app", backend.appName),
		backend: backend,
		stream:  stream,
		quit:    make(chan struct{}),
	}

	kcfg := kinsumer.NewConfig().WithShardCheckFrequency(freq).WithLeaderActionFrequency(freq)
	p.kinsumer, err = kinsumer.NewWithInterfaces(backend.Kinesis, backend.DynamoDB, stream, backend.appName, "rdss-archivematica-channel-adapter", kcfg)
	if err != nil {
		p.logger.Fatalln(err)
		return nil, err
	}

	if err := p.kinsumer.Run(); err != nil {
		p.logger.Fatalln(err)
	}

	go p.consumeRecords()

	return p, nil
}

func (p *processor) consumeRecords() {
	for {
		select {
		case <-p.quit:
			return
		default:
			record, err := p.kinsumer.Next()
			if err != nil {
				// TODO: this is currently unrecoverable state at the moment.
				// The backend should attempt to launch a new processor when
				// this is happening.
				panic(err)
			}
			if p.closed {
				return
			}
			p.route(record)
		}
	}
}

// route handles the message to all the handlers.
func (p *processor) route(data []byte) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, cb := range p.handlers {
		go cb(data)
	}
}

func (p *processor) addHandler(cb backend.Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers = append(p.handlers, cb)
}

func (p *processor) stop() {
	p.closed = true
	close(p.quit)
	p.kinsumer.Stop()
}
