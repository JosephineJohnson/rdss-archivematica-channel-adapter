package kinesis

import (
	"sync"

	log "github.com/Sirupsen/logrus"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/kinesis/kcl"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/kinesis/kcl/checkpointer"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/kinesis/kcl/locker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/kinesis/kcl/snitcher"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/errors"
)

// processor processes the data records from a Kinesis stream. In Kinesis, each
// consumer reads from a particular shard, using a shard iterator. This
// processor attempts to process all the shards available.
type processor struct {
	logger  log.FieldLogger
	backend *BackendImpl
	stream  string

	kcl    *kcl.Client
	reader *kcl.SharedReader

	handlers []backend.Handler
	mu       sync.RWMutex
}

func newProcessor(backend *BackendImpl, stream string) (*processor, error) {
	p := &processor{
		logger:  backend.logger.WithField("stream", stream),
		backend: backend,
		stream:  stream,
	}

	l := locker.NewDummyLocker()
	c := checkpointer.NewDummyCheckpointer()
	s := snitcher.NewDummySnitcher()
	p.kcl = kcl.New(p.backend.kinesis, l, c, s, p.logger)

	var err error
	p.reader, err = p.kcl.NewSharedReader(p.stream, p.backend.clientName)
	if err != nil {
		return nil, err
	}

	go p.loop()

	return p, nil
}

func (p *processor) loop() {
	var err error
	for m := range p.reader.Records() {
		err = p.route(m.Data)
		if err != nil {
			p.backend.putError(m, err)
		}
		// Regard the message as consumed by uptaing the checkpoint.
		err = p.reader.UpdateCheckpoint()
		if err != nil {
			p.logger.Errorln("The processor failed when it attempted to update the checkpoint!", err)
		}
	}
	if err = p.reader.Close(); err != nil {
		p.logger.Errorln("Reader failed when it was closing:", err)
	}
}

// route handles the message to all the handlers and returns the latest error
// captured in order to signal the caller that an error was produced.
func (p *processor) route(data []byte) error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	var err error
	for _, cb := range p.handlers {
		rErr := cb(data)
		if rErr != nil {
			err = rErr
		}
	}
	if err != nil {
		return &errors.Error{Err: err, Kind: errors.GENERR006}
	}
	return nil
}

func (p *processor) addHandler(cb backend.Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers = append(p.handlers, cb)
}

func (p *processor) stop() {
	p.reader.Close() // TODO
}
