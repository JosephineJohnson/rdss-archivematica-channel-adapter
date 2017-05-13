package amclient

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestNewTransferSession(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, err := NewTransferSession(c, "Test", fs)
	if err != nil {
		t.Error(err)
	}
	want := "/Test"
	if sess.Path != want {
		t.Errorf("Want %s, have %s", want, sess.Path)
	}
}

func TestNewTransferSession_FolderAlreadyExists(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	fs.Mkdir("/Test", os.FileMode(755))
	fs.Mkdir("/Test-1", os.FileMode(755))
	sess, err := NewTransferSession(c, "Test", fs)
	if err != nil {
		t.Fatal(err)
	}

	want := "/Test-2"
	if sess.Path != want {
		t.Fatalf("Want %s, have %s", want, sess.Path)
	}
}

func TestNewTransferSession_MaxAttemptsError(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	fs.Mkdir("/Test", os.FileMode(755))
	for i := 1; i <= 20; i++ {
		fs.Mkdir(fmt.Sprintf("/Test-%d", i), os.FileMode(755))
	}
	sess, err := NewTransferSession(c, "Test", fs)
	if sess != nil {
		t.Fatal("TransferSession returned should be nil")
	}
	if err == nil {
		t.Fatal("An error was expected!")
	}
}

func TestTransferSession_Create(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, err := NewTransferSession(c, "Test", fs)
	f, err := sess.Create("foobar.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	want := "/Test/foobar.jpg"
	have := f.Name()
	if want != have {
		t.Fatalf("Want %s, have %s", want, have)
	}
}

func TestTransferSession_Contents(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)
	afero.TempFile(sess.fs, "/", "uno")
	afero.TempFile(sess.fs, "/", "dos")

	want := 2
	have := len(sess.Contents())
	if want != have {
		t.Fatalf("Created %d files, only %d found", want, have)
	}
}

func TestTransferSession_Start(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)
	sess.Start()

	ts := c.Transfer.(*ts)
	if ts.approveReq.Directory != sess.Path {
		t.Errorf("Have %s, want %s", ts.approveReq.Directory, sess.Path)
	}
	if ts.approveReq.Type != "standard" {
		t.Errorf("Have %s, want \"standard\"", ts.approveReq.Type)
	}
}

func getClient(t *testing.T) *Client {
	url, _ := url.Parse("http://localhost")
	c := NewClient(nil, url.String(), "", "")
	c.Transfer = &ts{t: t}
	return c
}

type ts struct {
	t          *testing.T
	startReq   *TransferStartRequest
	approveReq *TransferApproveRequest
}

func (ts *ts) Start(ctx context.Context, req *TransferStartRequest) (*Response, error) {
	ts.startReq = req
	return &Response{}, nil
}

func (ts *ts) Approve(ctx context.Context, req *TransferApproveRequest) (*Response, error) {
	ts.approveReq = req
	return &Response{}, nil
}
