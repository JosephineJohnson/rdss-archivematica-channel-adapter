package amclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Bitstream struct {
	Path string
}

type IdentifyResult struct {
	ApplicationName    string `json:"applicationName,omitempty"`
	ApplicationVersion string `json:"applicationVersion,omitempty"`
	PUID               string `json:"puid"`
}

type FileIdentificationService interface {
	Identify(context.Context, *Bitstream) (*IdentifyResult, error)
}

func NewFileIdentificationService(bu string) FileIdentificationService {
	pur, _ := url.Parse(bu)
	return &SiegfriedService{
		BaseURL: pur,
	}
}

type SiegfriedService struct {
	BaseURL *url.URL
}

func (s *SiegfriedService) Identify(ctx context.Context, b *Bitstream) (*IdentifyResult, error) {
	if b.Path == "" {
		return nil, fmt.Errorf("unexpected path: %s", b.Path)
	}
	path64 := base64.StdEncoding.EncodeToString([]byte(b.Path))

	rel, err := url.Parse(fmt.Sprintf("/%s", path64))
	if err != nil {
		return nil, err
	}
	u := s.BaseURL.ResolveReference(rel)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rerr := resp.Body.Close(); err == nil {
			err = rerr
		}
	}()

	v := &IdentifyResult{}
	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return nil, err
	}

	return v, nil
}
