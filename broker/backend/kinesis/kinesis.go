package kinesis

import (
	"context"
	"strconv"
	"sync"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	"github.com/pkg/errors"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/util"
)

// New returns a kinesis backend with the corresponding AWS config initialized.
// Certain configuratino attributes specific to aws-sdk-go must be defined by
// the user of the adapter via environment variables.
func New(opts *backend.Opts) (backend.Backend, error) {
	b := &BackendImpl{
		procs: make(map[string]*processor),

		// TODO: use existing logger intead. No way to pass it right now!
		logger: log.StandardLogger().WithField("cmd", "consumer").WithField("z-backend", "kinesis"),
	}

	sess := session.Must(session.NewSession())
	if ep, ok := opts.Opts["endpoint"]; ok {
		sess.Config.Endpoint = aws.String(ep)
	}
	if tls, ok := opts.Opts["tls"]; ok {
		tls, err := strconv.ParseBool(tls)
		if err == nil && !tls {
			sess.Config.DisableSSL = aws.Bool(tls)
		}
	}

	b.session = sess
	b.kinesis = kinesis.New(b.session)

	if cn, ok := opts.Opts["client-name"]; ok {
		b.clientName = cn
	} else {
		b.clientName = util.GenId(10)
	}

	return b, nil
}

func init() {
	backend.Register("kinesis", New)
}

type BackendImpl struct {
	session *session.Session
	kinesis kinesisiface.KinesisAPI
	logger  log.FieldLogger

	// Application name
	clientName string

	// Each stream is assigned a processor.
	procs map[string]*processor
	mu    sync.Mutex
}

var _ backend.Backend = (*BackendImpl)(nil)

func (b *BackendImpl) Publish(topic string, data []byte) error {
	return nil // TODO
}

func (b *BackendImpl) Subscribe(topic string, cb backend.Handler) {
	p := b.processor(topic)
	p.addHandler(cb)
}

// processor returns the current processor for a given topic. The processor will
// be created if it hasn't been created before.
func (b *BackendImpl) processor(topic string) *processor {
	b.mu.Lock()
	defer b.mu.Unlock()
	p, ok := b.procs[topic]
	if !ok {
		b.logger.Infoln("Creating new stream processor")
		var err error
		p, err = newProcessor(b, topic)
		if err != nil {
			panic(err) // TODO
		}
		b.procs[topic] = p
	}
	return p
}

func (b *BackendImpl) Check(topic string) error {
	req := &kinesis.DescribeStreamInput{
		StreamName: aws.String(topic),
	}
	resp, err := b.kinesis.DescribeStreamWithContext(context.TODO(), req)
	if err != nil {
		return errors.Wrapf(err, "kinesis failed describing stream %s", topic)
	}
	status := *resp.StreamDescription.StreamStatus
	if status != "ACTIVE" {
		return errors.Errorf("kinesis stream %s is not active but %s", topic, status)
	}
	return nil
}

func (b *BackendImpl) Close() error {
	b.logger.Infoln("Backend is shutting down!")
	for _, p := range b.procs {
		p.stop()
	}
	return nil
}

// mError puts an erroneous message to the Error Message Queue.
func (b *BackendImpl) putError(r *kinesis.Record, err error) {
	b.logger.WithField("code", err).Errorln("Moving message to Error Message Queue:", *r.SequenceNumber)
	return
}

// mInvalid puts an invalid message to the Invalid Message Queue.
func (b *BackendImpl) putInvalid(r *kinesis.Record, err error) {
	b.logger.WithField("code", err).Errorln("Moving message to Invalid Message Queue:", *r.SequenceNumber)
	return
}
