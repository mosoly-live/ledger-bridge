package kinesis

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"gitlab.com/p-invent/mosoly-ledger-bridge/mth-core/data/kinesis/producer"
	"go.uber.org/zap"
)

// PutterConfig is a configuration for Kinesis putter
type PutterConfig struct {
	AwsAccessKey string
	AwsSecretKey string
	AwsRegion    string
	StreamName   string
}

// StreamPutter simplifies putting records to Amazon Kinesis
type StreamPutter struct {
	p *producer.Producer
}

// NewStreamPutter creates a new instance of StreamPutter using provided `cfg` (configuration).
func NewStreamPutter(cfg PutterConfig) (*StreamPutter, error) {
	config := &aws.Config{Region: aws.String(cfg.AwsRegion)}

	if cfg.AwsAccessKey != "" {
		c := credentials.NewStaticCredentials(cfg.AwsAccessKey, cfg.AwsSecretKey, "")
		config.Credentials = c
	}

	s, err := session.NewSession(config)
	if err != nil {
		return nil, err
	}
	client := kinesis.New(s)

	p := producer.New(&producer.Config{
		StreamName:    cfg.StreamName,
		Client:        client,
		BacklogCount:  2000,
		FlushInterval: 5 * time.Second,
	})
	p.Start()

	go func() {
		for r := range p.NotifyFailures() {
			zap.L().
				With(zap.String("partition_key", r.PartitionKey)).
				Error(r.Error())
		}
	}()

	return &StreamPutter{p}, nil
}

// Put puts event asynchronously to stream. This method is thread-safe.
func (p *StreamPutter) Put(e *Event) (err error) {
	if e == nil {
		return nil
	}

	var b bytes.Buffer

	je := json.NewEncoder(&b)
	err = je.Encode(e)
	if err != nil {
		return
	}

	data := b.Bytes()
	partitionKey := getPartitionKey(e)

	return p.p.Put(data, partitionKey)
}

func getPartitionKey(e *Event) string {
	userUUID := e.UserUUID
	if len(userUUID) > 0 {
		return userUUID
	}

	return strconv.FormatInt(int64(rand.Uint32()), 10)
}

// Close implements io.Closer interface
func (p *StreamPutter) Close() error {
	p.p.Stop()

	return nil
}
