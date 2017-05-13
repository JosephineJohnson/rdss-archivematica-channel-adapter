package amclient

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"golang.org/x/net/context"
)

// TransferSession lets you prepare a new transfer and submit it to
// Archivematica. It is a convenience tool around the transfer service.
type TransferSession struct {
	c    *Client
	fs   *afero.Afero
	Path string
}

// NewTransferSession returns a pointer to a new TransferSession.
func NewTransferSession(c *Client, name string, depositFs afero.Fs) (*TransferSession, error) {
	var (
		// The transfer folder is going to be created under depositFs after the
		// name provided. If the directory already exists we'll attempt to add
		// a numeric suffix corresponding to the number of attempts, e.g.:
		// Test, Test-1, Test-2... up to `maxTries` attempts - when the maximum
		// is reached, an error is reaturned instead.
		counter  = 1
		maxTries = 20
		nName    = name
	)
	for {
		err := depositFs.Mkdir(nName, os.FileMode(0755))
		if err == nil {
			fs := afero.NewBasePathFs(depositFs, nName)
			return &TransferSession{
				c:    c,
				fs:   &afero.Afero{Fs: fs},
				Path: afero.FullBaseFsPath(fs.(*afero.BasePathFs), "/"),
			}, nil
		}
		nName = fmt.Sprintf("%s-%d", name, counter)
		counter++
		if counter == maxTries {
			return nil, errors.Wrap(err, "Too many attempts. Last error")
		}
	}
}

// TransferSession returns a new transfer session.
func (c *Client) TransferSession(name string, depositFs afero.Fs) (*TransferSession, error) {
	return NewTransferSession(c, name, depositFs)
}

// Create creates a file in the filesystem, returning the file and an error, if
// any happens.
func (s *TransferSession) Create(name string) (afero.File, error) {
	return s.fs.Create(name)
}

// Start submits the transfer to Archivematica.
func (s *TransferSession) Start() error {
	ctx := context.TODO()
	_, err := s.c.Transfer.Approve(ctx, &TransferApproveRequest{
		Directory: s.Path,
		Type:      "standard",
	})
	return err
}

// Contents returns a list with all the files currently available in the
// temporary transfer filesystem.
func (s *TransferSession) Contents() []string {
	var paths []string
	s.fs.Walk("", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		paths = append(paths, path)
		return nil
	})
	return paths
}
