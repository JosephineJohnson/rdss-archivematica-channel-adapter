package s3

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// ObjectStorage is a S3-compatible storage interface.
type ObjectStorage interface {
	Download(ctx context.Context, w io.WriterAt, URI string) (int64, error)
}

// ObjectStorageImpl is our implementatino of the ObjectStorage interface.
type ObjectStorageImpl struct {
	S3           s3iface.S3API
	S3Session    *session.Session
	S3Downloader *s3manager.Downloader
}

// ClientOpt are options for New.
type ClientOpt func(*ObjectStorageImpl) error

// New returns a pointer to a new ObjectStorageImpl.
func New(opts ...ClientOpt) (ObjectStorage, error) {
	s := &ObjectStorageImpl{S3Session: session.Must(session.NewSession())}
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	s.S3 = s3.New(s.S3Session)
	s.S3Downloader = s3manager.NewDownloaderWithClient(s.S3)

	return s, nil
}

// SetEndpoint is a client option for setting the S3 endpoint.
func SetEndpoint(ep string) ClientOpt {
	return func(c *ObjectStorageImpl) error {
		c.S3Session.Config.Endpoint = aws.String(ep)
		return nil
	}
}

// SetKeys is a client option for setting the S3 keys.
func SetKeys(accessKey string, secretKey string) ClientOpt {
	return func(c *ObjectStorageImpl) error {
		creds := credentials.Value{
			AccessKeyID:     accessKey,
			SecretAccessKey: secretKey,
		}
		c.S3Session.Config.Credentials = credentials.NewStaticCredentialsFromCreds(creds)
		return nil
	}
}

// SetForcePathStyle is a client option for setting the S3 path style.
func SetForcePathStyle(force bool) ClientOpt {
	return func(c *ObjectStorageImpl) error {
		c.S3Session.Config.S3ForcePathStyle = aws.Bool(force)
		return nil
	}
}

// SetInsecureSkipVerify is a client option for setting the S3 client to skip
// the TLS verification process. aws-sdk-go relies on http.DefaultClient, here
// we build our custom client with a custom transport.
func SetInsecureSkipVerify(skip bool) ClientOpt {
	return func(c *ObjectStorageImpl) error {
		if !skip {
			return nil
		}
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		c.S3Session.Config.HTTPClient = &http.Client{Transport: tr}
		return nil
	}
}

// SetRegion is a client option for setting the S3 region.
func SetRegion(region string) ClientOpt {
	return func(c *ObjectStorageImpl) error {
		c.S3Session.Config.Region = aws.String(region)
		return nil
	}
}

// Download writes the contents of a remote file into the given writer.
func (s *ObjectStorageImpl) Download(ctx context.Context, w io.WriterAt, URI string) (n int64, err error) {
	bucket, key, err := getBucketAndKey(URI)
	if err != nil {
		return -1, err
	}
	req := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	return s.S3Downloader.DownloadWithContext(ctx, w, req)
}

func getBucketAndKey(URI string) (bucket string, key string, err error) {
	u, err := url.Parse(URI)
	if err != nil {
		return "", "", err
	}
	return u.Hostname(), strings.TrimPrefix(u.Path, "/"), nil
}
