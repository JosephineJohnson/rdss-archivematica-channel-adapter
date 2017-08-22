// +build !race

// Race detector disabled explicitly because MockDynamo is not safe.

package kinesis

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
	logt "github.com/sirupsen/logrus/hooks/test"
	"github.com/twitchscience/kinsumer/mocks"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/backend"
)

func Test_processor_route(t *testing.T) {
	var (
		count  int
		stopCh chan struct{}
	)
	tc := []struct {
		handlers []backend.Handler
		want     int
	}{
		{
			handlers: []backend.Handler{
				func([]byte) error { count++; stopCh <- struct{}{}; return nil },
				func([]byte) error { count++; stopCh <- struct{}{}; return nil },
			},
			want: 2,
		},
		{
			handlers: []backend.Handler{
				func([]byte) error { count++; stopCh <- struct{}{}; return errors.New("something really bad happened") },
				func([]byte) error { count++; stopCh <- struct{}{}; return nil },
				func([]byte) error { count++; stopCh <- struct{}{}; return nil },
			},
			want: 3,
		},
	}
	for _, tt := range tc {
		t.Run("", func(t *testing.T) {
			backend := getBackend(t)
			defer backend.Close()
			p := getProcessor(t, backend, "stream")

			count = 0
			stopCh = make(chan struct{}, len(tt.handlers))

			p.handlers = tt.handlers
			p.route([]byte("{}"))

			// Wait until delivered
			for i := 1; i <= len(tt.handlers); i++ {
				<-stopCh
			}

			if count != tt.want {
				t.Errorf("%d handlers were executed, wanted %d", count, tt.want)
			}
		})
	}
}

func Test_processor_addHandler(t *testing.T) {
	p := &processor{handlers: []backend.Handler{}}
	p.addHandler(func(data []byte) error { return nil })
	if got := len(p.handlers); got != 1 {
		t.Fatalf("addHandler(); unexpected number of handlers, got %d, want 1", got)
	}
}

func Test_processor_addHandler_safe(t *testing.T) {
	p := &processor{handlers: []backend.Handler{}}
	for i := 0; i < 10; i++ {
		go func() {
			p.addHandler(func(data []byte) error { return nil })
		}()
	}
}

func getBackend(t *testing.T) *BackendImpl {
	lt, _ := logt.NewNullLogger()

	// Create backend with non-default app name and table names
	backend := &BackendImpl{
		procs:   make(map[string]*processor),
		logger:  lt,
		Kinesis: myKinesis{},
		appName: "test_app_0",
		DynamoDB: myDynamo{mocks.NewMockDynamo([]string{
			"test_app_0_clients",
			"test_app_0_checkpoints",
			"test_app_0_metadata",
		})},
	}
	return backend
}

func getProcessor(t *testing.T, backend *BackendImpl, stream string) *processor {
	p, err := newProcessor(backend, stream)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

type myKinesis struct {
	kinesisiface.KinesisAPI
}

func (d myKinesis) DescribeStream(input *kinesis.DescribeStreamInput) (*kinesis.DescribeStreamOutput, error) {
	output := &kinesis.DescribeStreamOutput{}
	output.StreamDescription = &kinesis.StreamDescription{StreamStatus: aws.String("ACTIVE")}
	return output, nil
}

func (d myKinesis) DescribeStreamPages(*kinesis.DescribeStreamInput, func(*kinesis.DescribeStreamOutput, bool) bool) error {
	return nil
}

type myDynamo struct {
	dynamodbiface.DynamoDBAPI
}

func (d myDynamo) DescribeTable(input *dynamodb.DescribeTableInput) (*dynamodb.DescribeTableOutput, error) {
	output := &dynamodb.DescribeTableOutput{}
	output.Table = &dynamodb.TableDescription{TableStatus: aws.String("ACTIVE")}
	return output, nil
}

func (d myDynamo) DeleteItem(*dynamodb.DeleteItemInput) (*dynamodb.DeleteItemOutput, error) {
	output := &dynamodb.DeleteItemOutput{}
	return output, nil
}
