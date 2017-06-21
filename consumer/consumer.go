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
		name := getFilename(file.StorageLocation)
		if name == "" {
			err = ErrInvalidFile
			break
		}
		for _, c := range file.Checksums {
			switch c.Type {
			case "md5":
				t.ChecksumMD5(name, c.Value)
			case "sha1":
				t.ChecksumSHA1(name, c.Value)
			case "sha256":
				t.ChecksumSHA256(name, c.Value)
			}
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
			c.logger.Debugf("Saving %s into %s", file.StorageLocation, f.Name())
			n, err = c.s3.Download(c.ctx, f, file.StorageLocation)
			if err != nil {
				iErr = err
				c.logger.Errorf("Error downloading %s: %v", file.StorageLocation, err)
				return
			}
			c.logger.Debugf("Downloaded %s - %d bytes written", file.StorageLocation, n)
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

// datasetMetadata maps properties from a research object into a CSV entry
// (`amclient.FileMetadata`) in the `metadata.csv` file used in `amclient`.
func datasetMetadata(f *message.MetadataCreateRequest) *amclient.FileMetadata {
	entry := &amclient.FileMetadata{}

	// dc.title
	entry.DcTitle = f.Title

	// dc.type
	entry.DcType = f.Type

	// dc.identifier
	for _, item := range f.Identifiers {
		entry.DcIdentifier = item.Value
		break // TODO: we can't have multiple values for the same property atm
	}

	// dcterms.issued
	for _, item := range f.Dates {
		if item.Type != "published" {
			continue
		}
		entry.DcTermsIssued = item.Value
		break // TODO: we can't have multiple values for the same property atm
	}

	// dc.publisher
	for _, item := range f.Publishers {
		for _, itemOrgRoles := range item.OrganisationRole {
			entry.DcPublisher = itemOrgRoles.Organisation.Name
			break // TODO: we can't have multiple values for the same property atm
		}
	}

	return entry
}

// fileMetadata maps properties from an intellectual asset into a CSV entry
// (`amclient.FileMetadata`) in the `metadata.csv` file used in `amclient`.
func fileMetadata(name string, f *message.MetadataFile) *amclient.FileMetadata {
	return &amclient.FileMetadata{
		Filename:     "objects/" + name,
		DcIdentifier: f.Identifier,
		DcTitle:      f.Name,
	}
}
