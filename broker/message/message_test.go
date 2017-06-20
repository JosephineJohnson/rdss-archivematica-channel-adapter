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
		"objectUuid": "f4fe4ff2-d9a3-11e6-bf26-cec0c932ce01",
		"objectTitle": "A Sample Dataset",
		"objectDescription": "Description of the research object",
		"objectFile": [
			{
				"fileUuid": "baf0f6bc-3588-4f14-b24e-55a1a1d91930",
				"fileIdentifier": "keybar",
				"fileName": "keybar.awe",
				"fileSize": 1024,
				"fileChecksum": [
					{
						"checksumUuid": "cd7cb691-ac23-4a58-a875-c9988676448b",
						"checksumType": "md5",
						"checksumValue": "92c8ab01cecceb3bf0789c2cd8c7415a"
					}
				],
				"fileLabel": "kb",
				"fileHasMimeType": true,
				"fileFormatType": "text",
				"fileStorageLocation": "s3://bucketfoo/keybar.awe",
				"fileStorageType": "s3"
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
		"objectUuid": "a7e83002-29c1-11e7-93ae-92361f002671"
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
		"objectUuid": "a7e83002-29c1-11e7-93ae-92361f002671",
		"objectTitle": "Research about birds in the UK.",
		"objectDescription": "Description of the research object",
		"objectFile": [
			{
				"fileUuid": "ec2d4928-29c1-11e7-93ae-92361f002671",
				"fileIdentifier": "bird-sounds",
				"fileName": "bird-sounds.mp3",
				"fileSize": 910616,
				"fileChecksum": [
					{
						"checksumUuid": "cd7cb691-ac23-4a58-a875-c9988676448b",
						"checksumType": "md5",
						"checksumValue": "92c8ab01cecceb3bf0789c2cd8c7415a"
					},
					{
						"checksumUuid": "739c6bcc-cdeb-4f40-a095-6d0d11f3df2b",
						"checksumType": "sha256",
						"checksumValue": "18de3c56ff1084b6234e4fe609af475db1fddd56d11975de041d95ea094dbcf4"
					}
				],
				"fileLabel": "Bird sounds",
				"fileHasMimeType": true,
				"fileFormatType": "audio",
				"fileStorageLocation": "s3://rdss-prod-figshare-0132/bird-sounds.mp3",
				"fileStorageType": "s3"
			},
			{
				"fileUuid": "0dc88052-29c2-11e7-93ae-92361f002671",
				"fileIdentifier": "woodpigeon-pic",
				"fileName": "woodpigeon-pic.jpg",
				"fileSize": 147004,
				"fileChecksum": [
					{
						"checksumUuid": "9b2a5f85-d945-47e8-a410-02ee3d40b742",
						"checksumType": "md5",
						"checksumValue": "53a64110e067b14394c142c09571bea0"
					},
					{
						"checksumUuid": "18068198-470a-46d6-9aeb-857c33ca54ef",
						"checksumType": "sha256",
						"checksumValue": "476fa2fbd34bc96e5ec86b7c5ad81a071f4cd35c59a9be5e21a528ac9e04f66e"
					}
				],
				"fileLabel": "Woodpigeon pic",
				"fileHasMimeType": true,
				"fileFormatType": "image",
				"fileStorageLocation": "s3://rdss-prod-figshare-0132/woodpigeon-pic.jpg",
				"fileStorageType": "s3"
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
					UUID:        "f4fe4ff2-d9a3-11e6-bf26-cec0c932ce01",
					Title:       "A Sample Dataset",
					Description: "Description of the research object",
					Files: []*MetadataFile{
						{
							UUID:            "baf0f6bc-3588-4f14-b24e-55a1a1d91930",
							Identifier:      "keybar",
							Name:            "keybar.awe",
							Size:            1024,
							Label:           "kb",
							HasMimeType:     true,
							FormatType:      "text",
							StorageLocation: "s3://bucketfoo/keybar.awe",
							StorageType:     "s3",
							Checksums: []MetadataFileChecksum{
								{UUID: "cd7cb691-ac23-4a58-a875-c9988676448b", Type: "md5", Value: "92c8ab01cecceb3bf0789c2cd8c7415a"},
							},
						},
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
				Body: &MetadataReadRequest{
					UUID: "a7e83002-29c1-11e7-93ae-92361f002671",
				},
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
					UUID:        "a7e83002-29c1-11e7-93ae-92361f002671",
					Title:       "Research about birds in the UK.",
					Description: "Description of the research object",
					Files: []*MetadataFile{
						{
							UUID:            "ec2d4928-29c1-11e7-93ae-92361f002671",
							Identifier:      "bird-sounds",
							Name:            "bird-sounds.mp3",
							Size:            910616,
							Label:           "Bird sounds",
							HasMimeType:     true,
							FormatType:      "audio",
							StorageLocation: "s3://rdss-prod-figshare-0132/bird-sounds.mp3",
							StorageType:     "s3",
							Checksums: []MetadataFileChecksum{
								{UUID: "cd7cb691-ac23-4a58-a875-c9988676448b", Type: "md5", Value: "92c8ab01cecceb3bf0789c2cd8c7415a"},
								{UUID: "739c6bcc-cdeb-4f40-a095-6d0d11f3df2b", Type: "sha256", Value: "18de3c56ff1084b6234e4fe609af475db1fddd56d11975de041d95ea094dbcf4"},
							},
						},
						{
							UUID:            "0dc88052-29c2-11e7-93ae-92361f002671",
							Identifier:      "woodpigeon-pic",
							Name:            "woodpigeon-pic.jpg",
							Size:            147004,
							Label:           "Woodpigeon pic",
							HasMimeType:     true,
							FormatType:      "image",
							StorageLocation: "s3://rdss-prod-figshare-0132/woodpigeon-pic.jpg",
							StorageType:     "s3",
							Checksums: []MetadataFileChecksum{
								{UUID: "9b2a5f85-d945-47e8-a410-02ee3d40b742", Type: "md5", Value: "53a64110e067b14394c142c09571bea0"},
								{UUID: "18068198-470a-46d6-9aeb-857c33ca54ef", Type: "sha256", Value: "476fa2fbd34bc96e5ec86b7c5ad81a071f4cd35c59a9be5e21a528ac9e04f66e"},
							},
						},
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
