package consumer

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/amclient"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/s3"
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
	c.broker.SubscribeType(message.MessageTypeMetadataCreate, c.handleMetadataCreateRequest)

	<-c.ctx.Done()
	c.broker.Close()
	c.logger.Info("Consumer says good-bye!")
}

// handleMetadataCreateRequest handles the reception of a Metadata Create
// messages.
func (c *ConsumerImpl) handleMetadataCreateRequest(msg *message.Message) error {
	body, err := msg.MetadataCreateRequest()
	if err != nil {
		return err
	}

	// Ignore messages with no files listed
	if len(body.ObjectFile) == 0 {
		return nil
	}

	t, err := c.amc.TransferSession(body.ObjectTitle, c.depositFs)
	if err != nil {
		return err
	}
	err = t.ProcessingConfig("automated") // Automated workflow
	if err != nil {
		c.logger.Warningf("Failed to download `automated` processing configuration: %s", err)
	}
	describeDataset(t, body)
	for i, file := range body.ObjectFile {
		name := getFilename(file.FileStorageLocation)
		if name == "" {
			err = fmt.Errorf("malformed file storage location: %s (position %d)", file.FileStorageLocation, i)
			break
		}
		for _, c := range file.FileChecksum {
			switch c.ChecksumType {
			case message.ChecksumTypeEnum_md5:
				t.ChecksumMD5(name, c.ChecksumValue)
			case message.ChecksumTypeEnum_sha256:
				t.ChecksumSHA256(name, c.ChecksumValue)
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
			c.logger.Debugf("Saving %s into %s", file.FileStorageLocation, f.Name())
			n, err = c.s3.Download(c.ctx, f, file.FileStorageLocation)
			if err != nil {
				iErr = err
				c.logger.Errorf("Error downloading %s: %v", file.FileStorageLocation, err)
				return
			}
			c.logger.Debugf("Downloaded %s - %d bytes written", file.FileStorageLocation, n)
			describeFile(t, name, &file)
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

// describeDataset maps properties from a research object into a CSV entry
// in the `metadata.csv` file used in `amclient`.
func describeDataset(t *amclient.TransferSession, f *message.MetadataCreateRequest) {
	t.Describe("dc.title", f.ObjectTitle)
	t.Describe("dc.type", f.ObjectResourceType.String())

	for _, item := range f.ObjectIdentifier {
		t.Describe("dc.identifier", item.IdentifierValue)
	}

	for _, item := range f.ObjectDate {
		if item.DateType != message.DateTypeEnum_published {
			continue
		}
		t.Describe("dcterms.issued", item.DateValue)
	}

	for _, item := range f.ObjectOrganisationRole {
		t.Describe("dc.publisher", item.Organisation.OrganisationName)
	}

	for _, item := range f.ObjectPersonRole {
		if item.Role != message.PersonRoleEnum_dataCreator {
			continue
		}
		t.Describe("dc.contributor", item.Person.PersonGivenName)
	}
}

// describeFile maps properties from an intellectual asset into a CSV entry
// in the `metadata.csv` file used in `amclient`.
func describeFile(t *amclient.TransferSession, name string, f *message.File) {
	n := fmt.Sprintf("objects/%s", name)
	t.DescribeFile(n, "dc.identifier", f.FileIdentifier)
	t.DescribeFile(n, "dc.title", f.FileName)
}
