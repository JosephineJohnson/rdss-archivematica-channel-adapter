package amclient

import (
	"context"
	"fmt"
	"path/filepath"
)

const transferBasePath = "api/transfer"

// TransferService is an interface for interfacing with the Transfer endpoints
// of the Dashboard API.
type TransferService interface {
	Start(context.Context, *TransferStartRequest) (*TransferStartResponse, *Response, error)
	Approve(context.Context, *TransferApproveRequest) (*TransferApproveResponse, *Response, error)
	Unapproved(context.Context, *TransferUnapprovedRequest) (*TransferUnapprovedResponse, *Response, error)
}

// TransferServiceOp handles communication with the Tranfer related methods of
// the Archivematica API.
type TransferServiceOp struct {
	client *Client
}

var _ TransferService = &TransferServiceOp{}

// TransferStartRequest represents a request to start a transfer.
type TransferStartRequest struct {
	Name  string   `schema:"name"`
	Type  string   `schema:"type"`
	Paths []string `schema:"paths"`
}

// TransferStartResponse represents a response to TransferStartRequest.
type TransferStartResponse struct {
	Message string `schema:"message"`
	Path    string `schema:"path"`
}

// Start starts a new transfer.
func (s *TransferServiceOp) Start(ctx context.Context, r *TransferStartRequest) (*TransferStartResponse, *Response, error) {
	path := fmt.Sprintf("%s/start_transfer/", transferBasePath)

	req, err := s.client.NewRequest(ctx, "POST", path, r)
	if err != nil {
		return nil, nil, err
	}

	payload := &TransferStartResponse{}
	resp, err := s.client.Do(ctx, req, payload)

	return payload, resp, err
}

// TransferApproveRequest represents a request to approve a transfer.
type TransferApproveRequest struct {
	Type      string `schema:"type"`
	Directory string `schema:"directory"`
}

// TransferApproveResponse represents a response to TransferApproveRequest.
type TransferApproveResponse struct {
	Message string `json:"message"`
	UUID    string `json:"uuid"`
}

// Approve approves an existing transfer awaiting for approval.
func (s *TransferServiceOp) Approve(ctx context.Context, r *TransferApproveRequest) (*TransferApproveResponse, *Response, error) {
	path := fmt.Sprintf("%s/approve/", transferBasePath)

	r.Directory = filepath.Base(r.Directory) // We only need its base directory.
	req, err := s.client.NewRequest(ctx, "POST", path, r)
	if err != nil {
		return nil, nil, err
	}

	payload := &TransferApproveResponse{}
	resp, err := s.client.Do(ctx, req, payload)

	return payload, resp, err
}

// TransferUnapprovedRequest represents a request to list unapproved transfer.
type TransferUnapprovedRequest struct{}

// TransferUnapprovedResponse represents a response to TransferUnapprovedRequest.
type TransferUnapprovedResponse struct {
	Message string                              `json:"message"`
	Results []*TransferUnapprovedResponseResult `json:"results"`
}

// TransferUnapprovedResponseResult represents a result of
// TransferUnapprovedResponse.
type TransferUnapprovedResponseResult struct {
	Type      string `json:"type"`
	Directory string `json:"directory"`
	UUID      string `json:"uuid"`
}

// Unapproved lists existing transfers waiting for approval.
func (s *TransferServiceOp) Unapproved(ctx context.Context, r *TransferUnapprovedRequest) (*TransferUnapprovedResponse, *Response, error) {
	path := fmt.Sprintf("%s/unapproved/", transferBasePath)

	req, err := s.client.NewRequest(ctx, "GET", path, r)
	if err != nil {
		return nil, nil, err
	}

	payload := &TransferUnapprovedResponse{}
	resp, err := s.client.Do(ctx, req, payload)

	return payload, resp, err
}
