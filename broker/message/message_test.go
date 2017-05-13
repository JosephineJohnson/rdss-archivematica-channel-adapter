package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

const (
	metadataCreateRequest = `{
	"messageHeader": {
		"messageId": "e4bdaea0-4712-4682-8d4b-e6d92b1fe1ac",
		"messageType": "MetadataCreate",
		"messageClass": "Command"
	},
	"messageBody": {
		"datasetUuid": "f4fe4ff2-d9a3-11e6-bf26-cec0c932ce01",
		"datasetTitle": "A Sample Dataset",
		"files": [
			{
				"id": "baf0f6bc-3588-4f14-b24e-55a1a1d91930",
				"path": "s3://bucketfoo/keybar.awe"
			}
		]
	}
}`

	metadataReadRequest = `{
	"messageHeader": {
		"messageId": "8c1aec3b-426a-4a51-ad78-0213f093ac1c",
		"messageType": "MetadataRead",
		"messageClass": "Command"
	},
	"messageBody": {
		"datasetUuid": "a7e83002-29c1-11e7-93ae-92361f002671"
	}
}`

	metadataReadResponse = `{
	"messageHeader": {
		"messageId": "9e8f3cfc-29c2-11e7-93ae-92361f002671",
		"messageType": "MetadataRead",
		"messageClass": "Command",
		"correlationId": "8c1aec3b-426a-4a51-ad78-0213f093ac1c"
	},
	"messageBody": {
		"datasetUuid": "a7e83002-29c1-11e7-93ae-92361f002671",
		"datasetTitle": "Research about birds in the UK.",
		"files": [
			{
				"id": "ec2d4928-29c1-11e7-93ae-92361f002671",
				"path": "s3://rdss-prod-figshare-0132/bird-sounds.mp3"
			},
			{
				"id": "0dc88052-29c2-11e7-93ae-92361f002671",
				"path": "s3://rdss-prod-figshare-0132/woodpigeon-pic.jpg"
			}
		]
	}
}`
)

func TestFromJSON(t *testing.T) {
	testCases := []struct {
		data []byte // Message
		et   Type   // Expected message type
		ec   Class  // Expected message class
		ebt  string // Expected underlying type of body
	}{
		{[]byte(metadataCreateRequest), TypeMetadataCreate, ClassCommand, "*message.MetadataCreateRequest"},
		{[]byte(metadataReadRequest), TypeMetadataRead, ClassCommand, "*message.MetadataReadRequest"},
		{[]byte(metadataReadResponse), TypeMetadataRead, ClassCommand, "*message.MetadataReadResponse"},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.et, tc.ec), func(t *testing.T) {
			msg := &Message{}
			err := json.Unmarshal(tc.data, msg)
			if err != nil {
				t.Errorf("error unmarshalling the message: %s", err)
			}
			if msg.Header.Type != tc.et {
				t.Errorf("expected=%s received=%s", tc.et, msg.Header.Type)
			}
			if msg.Header.Class != tc.ec {
				t.Errorf("expected=%s received=%s", tc.ec, msg.Header.Class)
			}
			it := fmt.Sprintf("%s", reflect.TypeOf(msg.Body))
			if it != tc.ebt {
				t.Errorf("expected=%s received=%s", it, tc.ebt)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	testCases := []struct {
		data    []byte   // Expected JSON-encoded message
		message *Message // The message we're encoding
	}{
		{
			[]byte(metadataCreateRequest),
			&Message{
				Header: Headers{
					ID:    "e4bdaea0-4712-4682-8d4b-e6d92b1fe1ac",
					Type:  TypeMetadataCreate,
					Class: ClassCommand,
				},
				Body: &MetadataCreateRequest{
					UUID:  "f4fe4ff2-d9a3-11e6-bf26-cec0c932ce01",
					Title: "A Sample Dataset",
					Files: []*MetadataFile{
						{ID: "baf0f6bc-3588-4f14-b24e-55a1a1d91930", Path: "s3://bucketfoo/keybar.awe"},
					},
				},
			},
		},
		{
			[]byte(metadataReadRequest),
			&Message{
				Header: Headers{
					ID:    "8c1aec3b-426a-4a51-ad78-0213f093ac1c",
					Type:  TypeMetadataRead,
					Class: ClassCommand,
				},
				Body: &MetadataReadRequest{UUID: "a7e83002-29c1-11e7-93ae-92361f002671"},
			},
		},
		{
			[]byte(metadataReadResponse),
			&Message{
				Header: Headers{
					ID:            "9e8f3cfc-29c2-11e7-93ae-92361f002671",
					Type:          TypeMetadataRead,
					Class:         ClassCommand,
					CorrelationID: "8c1aec3b-426a-4a51-ad78-0213f093ac1c",
				},
				Body: &MetadataReadResponse{
					UUID:  "a7e83002-29c1-11e7-93ae-92361f002671",
					Title: "Research about birds in the UK.",
					Files: []*MetadataFile{
						{ID: "ec2d4928-29c1-11e7-93ae-92361f002671", Path: "s3://rdss-prod-figshare-0132/bird-sounds.mp3"},
						{ID: "0dc88052-29c2-11e7-93ae-92361f002671", Path: "s3://rdss-prod-figshare-0132/woodpigeon-pic.jpg"},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", tc.message.Header.Type, tc.message.Header.Class), func(t *testing.T) {
			// Encode it and indent it
			data, err := json.Marshal(tc.message)
			if err != nil {
				t.Errorf("ToJSON failed %s", err)
			}
			var out bytes.Buffer
			json.Indent(&out, data, "", "\t")

			// Compare them
			have := out.Bytes()
			if !bytes.Equal(have, tc.data) {
				t.Errorf("Unexpected result:\nHAVE: `%s`\nEXPECTED: `%s`", have, tc.data)
			}
		})
	}
}
