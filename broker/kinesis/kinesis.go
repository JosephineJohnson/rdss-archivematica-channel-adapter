package kinesis

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker"
)

// New initializes a Broker implementation that connects to Amazon Kinesis
// Streams using Streams API.
func New(opts *broker.Opts) (broker.Backend, error) {
	b := &brokerImpl{partitionKey: aws.String("1")}

	s, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	stream, ok := opts.Opts["stream"]
	if !ok {
		return nil, fmt.Errorf("missing required parameter: stream")
	}
	b.stream = aws.String(stream)

	// Change default endpoint, e.g. when you want to use kinesalite.
	endpoint, ok := opts.Opts["endpoint"]
	if ok {
		s.Config.Endpoint = aws.String(endpoint)
	}

	b.kinesis = kinesis.New(s)

	if err := checkStream(b.kinesis, stream); err != nil {
		return nil, fmt.Errorf("stream not available: %s", err)
	}

	return b, nil
}

func init() {
	broker.Register("kinesis", New)
}

type brokerImpl struct {
	kinesis      kinesisiface.KinesisAPI
	stream       *string
	partitionKey *string
}

var _ broker.Backend = (*brokerImpl)(nil)

// Request implements Broker.
func (b *brokerImpl) Request(context.Context, *broker.Message) error {
	_, err := b.kinesis.PutRecord(&kinesis.PutRecordInput{
		Data:         []byte("fooo"),
		StreamName:   b.stream,
		PartitionKey: b.partitionKey,
	})

	return err
}

// RequestResponse implements Broker.
// 1. convertToMessage()
// 2. decorateMessageHeaders()
// 3. getShardIterator(ShardIteratorType=TRIM_HORIZON)
// 4. updateNextShardIterator()
// 5. putRecord()
// 6. logOutboundMessage()
// 7. LOOP
//   7.1. getRecords(limit=10000, shardIterator=....)
//   7.2. recordsContainCorrelationId()
//   7.3. updateNextShardIterator()
//   7.4. sleep()
// 8. logInboundMessage()
// 9. convertToJson()
func (b *brokerImpl) RequestResponse(context.Context, *broker.Message) (*broker.Message, error) {
	iteratorID, err := b.getShardIterator("TRIM_HORIZON")
	if err != nil {
		return nil, err
	}

	_, err = b.kinesis.PutRecord(&kinesis.PutRecordInput{
		Data:       []byte("fooo"),
		StreamName: b.stream,
	})
	if err != nil {
		return nil, err
	}

	// TODO: logOutboundMessage

	for {
		resp, err := b.kinesis.GetRecordsWithContext(context.TODO(), &kinesis.GetRecordsInput{
			Limit:         aws.Int64(1000),
			ShardIterator: iteratorID,
		})
		if err != nil {
			return nil, err
		}

		for _, record := range resp.Records {
			fmt.Println(record.Data)
		}

		if false {
			break // TODO: stop the loop if...
		}

		iteratorID = resp.NextShardIterator
	}

	return nil, err
}

// Subscribe implements Broker.
func (b *brokerImpl) Subscribe(context.Context) error {
	return nil
}

func (b *brokerImpl) getShardIterator(iteratorType string) (*string, error) {
	desc, err := b.kinesis.DescribeStreamWithContext(context.TODO(), &kinesis.DescribeStreamInput{StreamName: b.stream})
	if err != nil {
		return nil, err
	}

	// We only want the first shard available for now.
	var resp *kinesis.GetShardIteratorOutput
	for _, shard := range desc.StreamDescription.Shards {
		resp, err = b.kinesis.GetShardIterator(&kinesis.GetShardIteratorInput{
			ShardId:           shard.ShardId,
			ShardIteratorType: aws.String(iteratorType),
		})

		return resp.ShardIterator, nil
	}

	return nil, fmt.Errorf("shard not found")
}
