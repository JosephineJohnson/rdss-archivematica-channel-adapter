package amclient

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
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
	fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewMemMapFs(), "/")}
	sess, _ := NewTransferSession(c, "Test", fs)

	tests := []struct {
		path string
		want string
	}{
		{"foobar.jpg", "/Test/foobar.jpg"},
		{"foo/bar.jpg", "/Test/foo/bar.jpg"},
		{"/foo/bar.jpg", "/Test/foo/bar.jpg"},
		{"/f/o/o/b/a/r.jpg", "/Test/f/o/o/b/a/r.jpg"},
	}
	for _, tt := range tests {
		f, err := sess.Create(tt.path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		have := f.Name()
		if tt.want != have {
			t.Fatalf("Want %s, have %s", tt.want, have)
		}
		dirPath := filepath.Dir(tt.want)
		if exists, err := fs.DirExists(dirPath); err != nil {
			t.Fatalf("DirExists failed: %v", err)
		} else if !exists {
			t.Fatalf("Directory %s not created", dirPath)
		}
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

func TestTransferSession_DescribeFile(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)
	entry := &FileMetadata{DcTitle: "Title"}
	sess.DescribeFile("foobar", entry)

	e, ok := sess.FileMetadata["foobar"]
	if !ok {
		t.Fatalf("The metadata entry was not added to the internal store")
	}
	if e != entry {
		t.Fatalf("The metadata entry found in the internal store wasn't the expected")
	}
}

func TestTransferSession_Describe(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)
	entry := &FileMetadata{DcTitle: "Title"}
	sess.Describe(entry)

	e, ok := sess.FileMetadata[objectsDirPrefix]
	if !ok {
		t.Fatalf("The metadata entry was not added to the internal store")
	}
	if e != entry {
		t.Fatalf("The metadata entry found in the internal store wasn't the expected")
	}
}

func TestTransferSession_createMetadataFile(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)
	sess.DescribeFile("foobar", &FileMetadata{Filename: "objects/foobar.jpg", DcTitle: "Title"})
	sess.Describe(&FileMetadata{DcTitle: "Birds are in danger"})
	sess.Start()

	have, err := sess.fs.ReadFile("/metadata/metadata.json")
	if err != nil {
		t.Fatalf("Error reading /metadata/metadata.json")
	}

	want := []byte(`[
	{
		"filename": "objects/",
		"dc.title": "Birds are in danger"
	},
	{
		"filename": "objects/foobar.jpg",
		"dc.title": "Title"
	}
]
`)
	if bytes.Compare(have, want) != 0 {
		t.Fatalf("Unexpected contents found in metadata file; want: %s, have %s", want, have)
	}
}

func TestTransferSession_createMetadataDir(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)

	if err := sess.createMetadataDir(); err != nil {
		t.Fatalf("createMetadataDir failed: %v", err)
	}
	info, err := sess.fs.Stat("metadata")
	if info == nil {
		t.Fatal("Metadata directory was not created")
	}
	if err != nil {
		t.Fatalf("Metadata directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("/metadata is not a directory")
	}
}

func TestTransferSession_Start(t *testing.T) {
	c := getClient(t)
	fs := afero.NewBasePathFs(afero.NewMemMapFs(), "/")
	sess, _ := NewTransferSession(c, "Test", fs)
	if err := sess.Start(); err != nil {
		t.Fatal(err)
	}

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

func (ts *ts) Start(ctx context.Context, req *TransferStartRequest) (*TransferStartResponse, *Response, error) {
	return nil, nil, nil
}

func (ts *ts) Approve(ctx context.Context, req *TransferApproveRequest) (*TransferApproveResponse, *Response, error) {
	ts.approveReq = req
	return &TransferApproveResponse{}, &Response{}, nil
}

func (ts *ts) Unapproved(ctx context.Context, req *TransferUnapprovedRequest) (*TransferUnapprovedResponse, *Response, error) {
	return &TransferUnapprovedResponse{
		Message: "Fetched unapproved transfers successfully.",
		Results: []*TransferUnapprovedResponseResult{
			&TransferUnapprovedResponseResult{Directory: "Test"},
		},
	}, &Response{}, nil
}

func TestTransferSession_ChecksumSet(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewBasePathFs(afero.NewMemMapFs(), "/")}
	set := NewChecksumSet("md5", fs)
	set.Add("bird-sounds.mp3", "92c8ab01cecceb3bf0789c2cd8c7415a")
	set.Add("woodpigeon-pic.jpg", "53a64110e067b14394c142c09571bea0")

	want1 := `92c8ab01cecceb3bf0789c2cd8c7415a bird-sounds.mp3
53a64110e067b14394c142c09571bea0 woodpigeon-pic.jpg
`
	want2 := `92c8ab01cecceb3bf0789c2cd8c7415a bird-sounds.mp3
53a64110e067b14394c142c09571bea0 woodpigeon-pic.jpg
`

	if err := set.Write(); err != nil {
		t.Fatal(err)
	}
	c, err := fs.ReadFile("/metadata/checksum.md5")
	if err != nil {
		t.Fatal(err)
	}
	have := string(c)
	if have != want1 || have != want2 {
		t.Fatalf("Unexpected content:\n%s", have)
	}
}
