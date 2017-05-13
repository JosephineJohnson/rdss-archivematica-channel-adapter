package amclient

import (
	"context"
	"fmt"
)

const transferBasePath = "api/transfer"

type TransferService interface {
	Start(context.Context, *TransferStartRequest) (*Response, error)
	Approve(context.Context, *TransferApproveRequest) (*Response, error)
}

// TransferServiceOp handles communication with the Tranfer related methods of
// the Archivematica API.
type TransferServiceOp struct {
	client *Client
}

var _ TransferService = &TransferServiceOp{}

// Start starts a new transfer.
func (s *TransferServiceOp) Start(ctx context.Context, r *TransferStartRequest) (*Response, error) {
	path := fmt.Sprintf("%s/start_transfer/", transferBasePath)

	req, err := s.client.NewRequest(ctx, "POST", path, r)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil) // TODO: don't ignore response

	return resp, err
}

// Approve approves an existing transfer awaiting for approval.
func (s *TransferServiceOp) Approve(ctx context.Context, r *TransferApproveRequest) (*Response, error) {
	path := fmt.Sprintf("%s/approve/", transferBasePath)

	req, err := s.client.NewRequest(ctx, "POST", path, r)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(ctx, req, nil) // TODO: don't ignore response

	return resp, err
}

// TransferStartRequest represents a request to start a transfer.
type TransferStartRequest struct {
	Name  string   `json:"name"`
	Type  string   `json:"type"`
	Paths []string `json:"paths"`
}

// TransferApproveRequest represents a request to approve a transfer.
type TransferApproveRequest struct {
	Type      string `json:"type"`
	Directory string `json:"directory"`
}
