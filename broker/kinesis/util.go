package kinesis

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
)

func checkStream(k kinesisiface.KinesisAPI, stream string) error {
	params := &kinesis.DescribeStreamInput{
		StreamName: aws.String(stream),
		Limit:      aws.Int64(1),
	}
	_, err := k.DescribeStream(params)
	if err != nil {
		return err
	}
	return nil
}
