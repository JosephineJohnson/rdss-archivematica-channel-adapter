package s3

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/spf13/afero"
)

var fs = afero.Afero{Fs: afero.NewMemMapFs()}

func tempFile(t *testing.T) afero.File {
	file, err := fs.TempFile("", "")
	t.Logf("Created temporary file: %s", file.Name())
	if err != nil {
		t.Error(err)
	}
	return file
}

type mockS3Client struct {
	s3iface.S3API
	t *testing.T
	f afero.File
}

func (c *mockS3Client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{
		Body:         c.f,
		ContentRange: aws.String("1"),
	}, nil
}

func TestObjectStorageNew(t *testing.T) {
	tests := []struct {
		opts    []ClientOpt
		wantErr bool
	}{
		{[]ClientOpt{}, false},
		{[]ClientOpt{SetRegion("foobar")}, false},
		{[]ClientOpt{func(*ObjectStorageImpl) error { return errors.New("error") }}, true},
	}
	for _, tt := range tests {
		s, err := New(tt.opts...)
		if tt.wantErr {
			if s != nil || err == nil {
				t.Error()
			}
		} else {
			if s == nil || err != nil {
				t.Error()
			}
		}
	}
}

func TestObjectStorageImpl_Download(t *testing.T) {
	const want = "Hello world!"

	// Input file S3 mock is to read from
	fi := tempFile(t)
	defer fi.Close()
	fmt.Fprintf(fi, want) // Write the contents
	fi.Seek(0, 0)

	// Output file we want to validate
	fo := tempFile(t)
	defer fo.Close()

	s3c := &mockS3Client{t: t, f: fi}
	s3d := s3manager.NewDownloaderWithClient(s3c)
	client := &ObjectStorageImpl{S3: s3c, S3Downloader: s3d}

	var err error

	_, err = client.Download(context.TODO(), fo, "[invalid-url]:12345")
	if err == nil {
		t.Error("Download() should have returned an error but didn't")
	}

	_, err = client.Download(context.TODO(), fo, "s3://foo/bar")
	if err != nil {
		t.Error(err)
	}

	fo.Seek(0, 0)
	data, err := ioutil.ReadAll(fo)
	if err != nil {
		t.Error(err)
	}
	have := string(data)
	if want != have {
		t.Errorf("want %s, got %s", want, have)
	}
}

func TestCustomEndpoint(t *testing.T) {
	c, err := New(SetEndpoint("http://my.end-point.tld"))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	ci := c.(*ObjectStorageImpl)
	want := "http://my.end-point.tld"
	if got := *ci.S3Session.Config.Endpoint; got != want {
		t.Errorf("New() Endpoint = %s; want %s", got, want)
	}
}

func TestCustomKeys(t *testing.T) {
	var (
		aKey = "foo"
		sKey = "bar"
	)
	c, err := New(SetKeys(aKey, sKey))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	ci := c.(*ObjectStorageImpl)
	v, err := ci.S3Session.Config.Credentials.Get()
	if err != nil {
		t.Fatalf("ci.S3Session.Config.Credentials.Get() unexpected error: %v", err)
	}

	if v.AccessKeyID != aKey {
		t.Errorf("AccessKeyID = %s; want %s", v.AccessKeyID, aKey)
	}
	if v.SecretAccessKey != sKey {
		t.Errorf("SecretAccessKey = %s; want %s", v.SecretAccessKey, sKey)
	}
}

func Test_getBucketAndKey(t *testing.T) {
	testCases := []struct {
		url     string
		bucket  string
		key     string
		wantErr bool
	}{
		{"s3://rdss-bucker-2344/filename.jpg", "rdss-bucker-2344", "filename.jpg", false},
		{"s3://a-different-bucket/wqefqwef/cert.pem", "a-different-bucket", "wqefqwef/cert.pem", false},
		{"[invalid-url]:12345", "", "", true},
	}
	for _, tc := range testCases {
		bucket, key, err := getBucketAndKey(tc.url)
		if tc.wantErr {
			if bucket != "" || key != "" || err == nil {
				t.Errorf("getBucketAndKey() was expected to fail but didn't")
			}
			return
		}
		if err != nil {
			t.Errorf("Unexpected error in getBucketAndKey: %s", err)
		}
		if bucket != tc.bucket {
			t.Errorf("Unexpected bucket - got: %s, want: %s", bucket, tc.bucket)
		}
		if key != tc.key {
			t.Errorf("Unexpected key - got: %s, want: %s", key, tc.key)
		}
	}
}

func TestSetRegion(t *testing.T) {
	s := &ObjectStorageImpl{S3Session: session.Must(session.NewSession())}
	tests := []string{"region1", "region2", ""}
	for _, tt := range tests {
		want := tt
		SetRegion(want)(s)
		if got := s.S3Session.Config.Region; *got != want {
			t.Errorf("SetRegion() = %v, want %v", *got, want)
		}
	}
}

func TestSetForcePathStyle(t *testing.T) {
	s := &ObjectStorageImpl{S3Session: session.Must(session.NewSession())}
	tests := []bool{true, false}
	for _, tt := range tests {
		want := tt
		SetForcePathStyle(want)(s)
		if got := s.S3Session.Config.S3ForcePathStyle; *got != want {
			t.Errorf("SetForcePathStyle() = %v, want %v", *got, want)
		}
	}
}

func TestSetInsecureSkipVerify(t *testing.T) {
	s := &ObjectStorageImpl{S3Session: session.Must(session.NewSession())}
	tests := []bool{true, false}
	for _, skip := range tests {
		want := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		SetInsecureSkipVerify(skip)(s)
		if !reflect.DeepEqual(want, s.S3Session.Config.HTTPClient.Transport) {
			t.Error("SetInsecureSkipVerify() hasn't updated HTTPClient properly")
		}
	}
}
