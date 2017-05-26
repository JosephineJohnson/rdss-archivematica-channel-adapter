package amclient

import (
	"fmt"
	"net/http"
	"testing"
)

func TestTransfer_Start(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/transfer/start_transfer/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"message": "Copy successful", "path": "/var/foobar"}`)
	})

	payload, _, err := client.Transfer.Start(ctx, &TransferStartRequest{
		Name:  "foobar",
		Paths: []string{"a.jpg", "b.jpg"},
		Type:  "standard",
	})
	if err != nil {
		t.Errorf("Transfer.Start returned error: %v", err)
	}
	if want, got := "Copy successful", payload.Message; want != got {
		t.Errorf("Transfer.Start(): Message = %v, want %v", got, want)
	}
}

func TestTransfer_Approve(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/transfer/approve/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"message": "Approval successful.", "uuid": "eaedbee3-2b02-4e40-baa0-3ef92c5fd17e"}`)
	})

	payload, _, err := client.Transfer.Approve(ctx, &TransferApproveRequest{
		Directory: "Foobar",
		Type:      "standard",
	})
	if err != nil {
		t.Errorf("Transfer.Approve returned error: %v", err)
	}
	if want, got := "Approval successful.", payload.Message; want != got {
		t.Errorf("Transfer.Approve(): Message = %v, want %v", got, want)
	}
	if want, got := "eaedbee3-2b02-4e40-baa0-3ef92c5fd17e", payload.UUID; want != got {
		t.Errorf("Transfer.Approve(): UUID = %v, want %v", got, want)
	}
}

func TestTransfer_Unapproved(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/transfer/unapproved/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{
			"message": "Fetched unapproved transfers successfully.",
			"results": [
				{
					"type": "standard",
					"directory": "/var/foobar1",
					"uuid": "eaedbee3-2b02-4e40-baa0-3ef92c5fd17e"
				},
				{
					"type": "standard",
					"directory": "/var/foobar2",
					"uuid": "433f20e4-a0e4-484b-8fb4-ec9b3cda4cfc"
				}
			]
		}`)
	})

	payload, _, err := client.Transfer.Unapproved(ctx, &TransferUnapprovedRequest{})
	if err != nil {
		t.Errorf("Transfer.Unapproved() returned error: %v", err)
	}
	if want, got := 2, len(payload.Results); want != got {
		t.Errorf("Transfer.Unapproved() len(Results) %v, want %v", got, want)
	}
}
