package amclient

import (
	"context"
	"fmt"
	"os"
	"time"

	"path"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// Timeout is the longest that TransferSession is going to wait before
	// it ceases to wait for the transfer to be picked by MCP and be listed.
	Timeout = 10 * time.Second

	// Maximum number of attempts to list unapproved transfers.
	MaxAttempts = 10
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

// ProcessingConfig inclues a processing configuration (processingMCP.xml)
// given its name.
func (s *TransferSession) ProcessingConfig(name string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	config, _, err := s.c.ProcessingConfig.Get(ctx, name)
	if err != nil {
		return err
	}

	err = s.fs.SafeWriteReader("/processingMCP.xml", config)
	if err != nil {
		return err
	}

	return nil
}

// Start submits the transfer to Archivematica.
func (s *TransferSession) Start() error {
	var (
		ctx      = context.Background()
		req      = &TransferUnapprovedRequest{}
		basePath = path.Base(s.Path)
		attempts = 0
	)
	ctx, cancel := context.WithTimeout(ctx, Timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "context done")
		default:
		}

		attempts++
		payload, _, err := s.c.Transfer.Unapproved(ctx, req)
		if err != nil {
			return errors.Wrap(err, "unapproved request failed")
		}
		for _, item := range payload.Results {
			if item.Directory == basePath {
				_, _, err := s.c.Transfer.Approve(ctx, &TransferApproveRequest{
					Directory: s.Path,
					Type:      "standard",
				})
				if err != nil {
					return errors.Wrap(err, "approve request failed")
				}
				return nil
			}
		}

		if attempts == MaxAttempts {
			return errors.Errorf("maximum number of attempts reached: %d", MaxAttempts)
		}

		// Wait for an extra second before the next attempt
		time.Sleep(time.Second * 1)
	}
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
