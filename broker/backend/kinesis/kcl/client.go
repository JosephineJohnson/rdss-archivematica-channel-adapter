package kcl

import (
	"errors"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/kinesis/kcl/checkpointer"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/kinesis/kcl/locker"
	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend/kinesis/kcl/snitcher"
)

var (
	ErrMissingLocker       = errors.New("Missing locker")
	ErrMissingCheckpointer = errors.New("Missing checkpointer")
	ErrMissingSnitcher     = errors.New("Missing snitcher")
	ErrShardLocked         = errors.New("Shard locked")
)

type Client struct {
	kinesis    kinesisiface.KinesisAPI
	distlock   locker.Locker
	checkpoint checkpointer.Checkpointer
	snitch     snitcher.Snitcher
	logger     log.FieldLogger
}

func New(kinesis kinesisiface.KinesisAPI, distlock locker.Locker, checkpoint checkpointer.Checkpointer, snitch snitcher.Snitcher, logger log.FieldLogger) *Client {
	return &Client{
		kinesis:    kinesis,
		distlock:   distlock,
		checkpoint: checkpoint,
		snitch:     snitch,
		logger:     logger,
	}
}

func (c *Client) StreamDescription(streamName string) (*kinesis.StreamDescription, error) {
	out, err := c.kinesis.DescribeStream(&kinesis.DescribeStreamInput{
		StreamName: aws.String(streamName),
	})
	if err != nil {
		return nil, err
	}

	return out.StreamDescription, nil
}
