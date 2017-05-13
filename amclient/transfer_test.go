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

	_, err := client.Transfer.Start(ctx, &TransferStartRequest{
		Name:  "foobar",
		Paths: []string{"a.jpg", "b.jpg"},
		Type:  "standard",
	})
	if err != nil {
		t.Errorf("Transfer.Start returned error: %v", err)
	}

	// TODO: Response
}

func TestTransfer_Approve(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/api/transfer/approve/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"message": "Copy successful", "path": "/var/foobar"}`)
	})

	_, err := client.Transfer.Approve(ctx, &TransferApproveRequest{
		Directory: "/var/foobar",
		Type:      "standard",
	})
	if err != nil {
		t.Errorf("Transfer.Approve returned error: %v", err)
	}

	// TODO: Response
}
