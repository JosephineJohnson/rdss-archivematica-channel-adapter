package consumer

import (
	"context"
	"errors"
	"net/url"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/s3"
	"github.com/spf13/afero"
)

// Consumer is the component that subscribes to the broker and interacts with
// Archivematica.
type Consumer interface {
	Start()
}

// ConsumerImpl is an implementation of Consumer.
type ConsumerImpl struct {
	broker    *broker.Broker
	ctx       context.Context
	logger    log.FieldLogger
	amc       *amclient.Client
	s3        s3.ObjectStorage
	depositFs afero.Fs
}

// MakeConsumer returns a new ConsumerImpl which implements Consumer
func MakeConsumer(
	ctx context.Context,
	logger log.FieldLogger,
	broker *broker.Broker,
	amc *amclient.Client,
	s3 s3.ObjectStorage,
	depositFs afero.Fs) *ConsumerImpl {
	return &ConsumerImpl{
		ctx:       ctx,
		logger:    logger,
		broker:    broker,
		amc:       amc,
		s3:        s3,
		depositFs: depositFs,
	}
}

// Start implements Consumer
func (c *ConsumerImpl) Start() {
	c.broker.SubscribeType(message.TypeMetadataCreate, c.handleMetadataCreateRequest)

	<-c.ctx.Done()
	c.broker.Close()
	c.logger.Info("Consumer says good-bye!")
}

var (
	ErrUnpexpectedPayloadType = errors.New("unexpected payload type")
	ErrInvalidFile            = errors.New("invalid file")
)

// handleMetadataCreateRequest handles the reception of a Metadata Create
// messages.
func (c *ConsumerImpl) handleMetadataCreateRequest(msg *message.Message) error {
	body, ok := msg.Body.(*message.MetadataCreateRequest)
	if !ok {
		return ErrUnpexpectedPayloadType
	}
	t, err := c.amc.TransferSession(body.Title, c.depositFs)
	if err != nil {
		return err
	}
	err = t.ProcessingConfig("automated") // Automated workflow
	if err != nil {
		c.logger.Warningf("Failed to download `automated` processing configuration: %s", err)
	}
	t.Describe(datasetMetadata(body))
	for _, file := range body.Files {
		name := getFilename(file.Path)
		if name == "" {
			err = ErrInvalidFile
			break
		}
		// Using an anonymous function so I can use defer inside this loop.
		var iErr error
		func() {
			var (
				f afero.File
				n int64
			)
			f, err = t.Create(name)
			if err != nil {
				iErr = err
				c.logger.Errorf("Error creating %s: %v", name, err)
				return
			}
			defer f.Close()
			c.logger.Debugf("Saving %s into %s", file.Path, f.Name())
			n, err = c.s3.Download(c.ctx, f, file.Path)
			if err != nil {
				iErr = err
				c.logger.Errorf("Error downloading %s: %v", file.Path, err)
				return
			}
			c.logger.Debugf("%d bytes written", n)
			t.DescribeFile(name, fileMetadata(name, file))
		}()
		if iErr != nil {
			return iErr
		}
	}
	return t.Start()
}

func getFilename(path string) string {
	u, err := url.Parse(path)
	if err != nil {
		return ""
	}
	return strings.TrimPrefix(u.Path, "/")
}

func datasetMetadata(f *message.MetadataCreateRequest) *amclient.FileMetadata {
	return &amclient.FileMetadata{
		DcTitle: f.Title,
	}
}

func fileMetadata(name string, f *message.MetadataFile) *amclient.FileMetadata {
	return &amclient.FileMetadata{
		Filename: "objects/" + name,
		DcTitle:  f.Title,
	}
}
