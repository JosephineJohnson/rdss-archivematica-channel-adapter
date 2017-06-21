package amclient

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"path"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// Timeout is the longest that TransferSession is going to wait before
	// it ceases to wait for the transfer to be picked by MCP and be listed.
	Timeout = 10 * time.Second

	// MaxAttempts is the number of attempts to list unapproved transfers.
	MaxAttempts = 10

	// Objects prefix
	objectsDirPrefix = "objects/"
)

// TransferSession lets you prepare a new transfer and submit it to
// Archivematica. It is a convenience tool around the transfer service.
type TransferSession struct {
	c    *Client
	fs   *afero.Afero
	Path string

	FileMetadata map[string]*FileMetadata

	ChecksumsMD5    *ChecksumSet
	ChecksumsSHA1   *ChecksumSet
	ChecksumsSHA256 *ChecksumSet
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
			ts := &TransferSession{
				c:            c,
				fs:           &afero.Afero{Fs: fs},
				Path:         afero.FullBaseFsPath(fs.(*afero.BasePathFs), "/"),
				FileMetadata: make(map[string]*FileMetadata),
			}
			ts.ChecksumsMD5 = NewChecksumSet("md5", ts.fs)
			ts.ChecksumsSHA1 = NewChecksumSet("sha1", ts.fs)
			ts.ChecksumsSHA256 = NewChecksumSet("sha256", ts.fs)
			return ts, nil
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
	err := s.fs.MkdirAll(filepath.Dir(name), os.FileMode(0755))
	if err != nil {
		return nil, err
	}
	f, err := s.fs.Create(name)
	if err != nil {
		return nil, err
	}
	return f, nil
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

	if err := s.createMetadataDir(); err != nil {
		return err
	}

	if err := s.createMetadataFile(); err != nil {
		return err
	}

	if err := s.createChecksumsFiles(); err != nil {
		return err
	}

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

// DescribeFile registers metadata of a file. It causes the transfer to include
// a `metadata.json` file with the metadata of each file described.
func (s *TransferSession) DescribeFile(name string, m *FileMetadata) {
	s.FileMetadata[name] = m
}

// Describe registers metadata of the whole dataset/transfer. It causes the
// transfer to include a `metadata.json` file with the metadata included.
func (s *TransferSession) Describe(m *FileMetadata) {
	m.Filename = objectsDirPrefix
	s.FileMetadata[objectsDirPrefix] = m
}

// ChecksumMD5 registers a MD5 checksum for a file.
func (s *TransferSession) ChecksumMD5(name, sum string) {
	s.ChecksumsMD5.Add(name, sum)
}

// ChecksumSHA1 registers a SHA1 checksum for a file.
func (s *TransferSession) ChecksumSHA1(name, sum string) {
	s.ChecksumsSHA1.Add(name, sum)
}

// ChecksumSHA256 registers a SHA256 checksum for a file.
func (s *TransferSession) ChecksumSHA256(name, sum string) {
	s.ChecksumsSHA256.Add(name, sum)
}

func (s *TransferSession) createMetadataFile() error {
	if len(s.FileMetadata) == 0 {
		return nil
	}
	const path = "/metadata/metadata.json"
	fd, err := s.fs.Create(path)
	defer fd.Close()
	if err != nil {
		return fmt.Errorf("error creating metadata.json: %s", err)
	}
	entries := make([]*FileMetadata, 0, len(s.FileMetadata))
	// Let's try to add the main "objects/" entry first which is what the reader
	// would probably expect when inspecting the `metadata.json` file.
	if entry, ok := s.FileMetadata[objectsDirPrefix]; ok {
		entries = append(entries, entry)
		delete(s.FileMetadata, objectsDirPrefix)
	}
	for _, entry := range s.FileMetadata {
		entries = append(entries, entry)
	}
	enc := json.NewEncoder(fd)
	enc.SetIndent("", "\t")
	if err := enc.Encode(entries); err != nil {
		return err
	}
	return nil
}

func (s *TransferSession) createMetadataDir() error {
	const path = "/metadata"
	if _, err := s.fs.Stat(path); err != nil {
		return s.fs.Mkdir(path, os.FileMode(0755))
	}
	return nil
}

func (s *TransferSession) createChecksumsFiles() error {
	if err := s.ChecksumsMD5.Write(); err != nil {
		return err
	}
	if err := s.ChecksumsSHA1.Write(); err != nil {
		return err
	}
	if err := s.ChecksumsSHA256.Write(); err != nil {
		return err
	}
	return nil
}

// FileMetadata represents the metadata entry of a file (see `metadata.json`).
type FileMetadata struct {
	Filename      string `json:"filename"`
	DcIdentifier  string `json:"dc.identifier,omitempty"`
	DcTitle       string `json:"dc.title,omitempty"`
	DcContributor string `json:"dc.contributor,omitempty"`
	DcPublisher   string `json:"dc.publisher,omitempty"`
	DcTermsIssued string `json:"dcterms.issued,omitempty"`
	DcFormat      string `json:"dc.format,omitempty"`
	DcType        string `json:"dc.type,omitempty"`
}

// ChecksumSet holds the checksums of the files for a sum algorithm.
type ChecksumSet struct {
	sumType string
	values  map[string]string
	fs      afero.Fs
}

func NewChecksumSet(sumType string, fs afero.Fs) *ChecksumSet {
	return &ChecksumSet{
		sumType: sumType,
		values:  make(map[string]string),
		fs:      fs,
	}
}

func (c *ChecksumSet) Add(name, sum string) {
	c.values[name] = sum
}

func (c *ChecksumSet) Write() error {
	const (
		path = "/metadata/checksum.%s"
		sep  = ' '
	)
	if len(c.values) == 0 {
		return nil
	}
	f, err := c.fs.Create(fmt.Sprintf(path, c.sumType))
	if err != nil {
		return err
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	writer.Comma = sep
	writer.UseCRLF = false
	defer writer.Flush()
	for name, sum := range c.values {
		if err := writer.Write([]string{sum, name}); err != nil {
			return err
		}
	}
	return nil
}
