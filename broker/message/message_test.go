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
    "messageClass": "Command",
    "messageType": "MetadataCreate"
  },
  "messageBody": {
    "objectUUID": "06f75186-f9cb-4be3-8df7-64dd037eb54f",
    "objectTitle": "Non-uniform Mesh for Embroidered Microstrip Antennas - Simulation files",
    "objectContributor": [
      {
        "Person": {
          "personUUID": "479dada4-8650-421e-8480-63d58107a998",
          "personGivenName": "Zhang, Shiyu"
        },
        "Role": "dataCreator"
      },
      {
        "Person": {
          "personUUID": "b8102d3a-2c49-419c-aae4-312c2e08c3ce",
          "personGivenName": "Whittow, William"
        },
        "Role": "dataCreator"
      },
      {
        "Person": {
          "personUUID": "b65be0c5-e451-4237-bc1e-abb1e5222c4d",
          "personGivenName": "Seager, Rob"
        },
        "Role": "dataCreator"
      },
      {
        "Person": {
          "personUUID": "e51803fc-953b-4771-9e62-01a138d9303a",
          "personGivenName": "Chauraya, Alford"
        },
        "Role": "dataCreator"
      },
      {
        "Person": {
          "personUUID": "34c6e333-c02a-4965-b0a2-5a1f80dedf22",
          "personGivenName": "Vardaxoglou, Yiannis"
        },
        "Role": "dataCreator"
      }
    ],
    "objectDescription": "The fileset contains simulation files of the non-uniform meshed patch antennas (using FDTD Empire XCcel). The conductor thickness and conductivity can be adjusted according to the conductive threads.",
    "objectDate": [
      {
        "dateValue": "2017-03-17",
        "dateType": "published"
      }
    ],
    "objectResourceType": "Dataset",
    "objectIdentifier": [
      {
        "identifierValue": "10.17028/rd.lboro.4665448.v1",
        "identifierType": "DOI"
      }
    ],
    "objectPublisher": [
      {
        "Organisation": {
          "organisationName": "Loughborough University",
          "organisationAddress": "Epinal Way, Loughborough LE11 3TU, UK"
        },
        "Role": "publisher"
      }
    ],
    "objectFile": [
      {
        "fileUUID": "42143745-dba9-9c98-cdd7-0c521ae55118",
        "fileIdentifier": "1",
        "fileName": "woodpigeon-pic.jpg",
        "fileSize": 147004,
        "fileChecksum": [
          {
            "checksumType": "md5",
            "checksumValue": "53a64110e067b14394c142c09571bea0"
          }
        ],
        "fileStorageLocation": "s3://rdss-prod-figshare-0132/woodpigeon-pic.jpg",
        "fileStorageType": "S3"
      },
      {
        "fileUUID": "9102f719-2f78-d9c2-de12-15f27024f78b",
        "fileIdentifier": "2",
        "fileName": "bird-sounds.mp3",
        "fileSize": 910616,
        "fileChecksum": [
          {
            "checksumType": "md5",
            "checksumValue": "92c8ab01cecceb3bf0789c2cd8c7415a"
          }
        ],
        "fileStorageLocation": "s3://rdss-prod-figshare-0132/bird-sounds.mp3",
        "fileStorageType": "S3"
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
					Class: ClassCommand,
					Type:  TypeMetadataCreate,
				},
				Body: &MetadataCreateRequest{
					UUID:  "06f75186-f9cb-4be3-8df7-64dd037eb54f",
					Title: "Non-uniform Mesh for Embroidered Microstrip Antennas - Simulation files",
					Contributors: []*PersonRole{
						{
							Person: &Person{
								UUID:      "479dada4-8650-421e-8480-63d58107a998",
								GivenName: "Zhang, Shiyu",
							},
							Role: "dataCreator",
						},
						{
							Person: &Person{
								UUID:      "b8102d3a-2c49-419c-aae4-312c2e08c3ce",
								GivenName: "Whittow, William",
							},
							Role: "dataCreator",
						},
						{
							Person: &Person{
								UUID:      "b65be0c5-e451-4237-bc1e-abb1e5222c4d",
								GivenName: "Seager, Rob",
							},
							Role: "dataCreator",
						},
						{
							Person: &Person{
								UUID:      "e51803fc-953b-4771-9e62-01a138d9303a",
								GivenName: "Chauraya, Alford",
							},
							Role: "dataCreator",
						},
						{
							Person: &Person{
								UUID:      "34c6e333-c02a-4965-b0a2-5a1f80dedf22",
								GivenName: "Vardaxoglou, Yiannis",
							},
							Role: "dataCreator",
						},
					},
					Description: "The fileset contains simulation files of the non-uniform meshed patch antennas (using FDTD Empire XCcel). The conductor thickness and conductivity can be adjusted according to the conductive threads.",
					Dates: []*Date{
						{
							Type:  "published",
							Value: "2017-03-17",
						},
					},
					ResourceType: "Dataset",
					Identifiers: []*Identifier{
						{
							Value: "10.17028/rd.lboro.4665448.v1",
							Type:  "DOI",
						},
					},
					Publishers: []*OrganisationRole{
						{
							Role: "publisher",
							Organisation: &Organisation{
								Name:    "Loughborough University",
								Address: "Epinal Way, Loughborough LE11 3TU, UK",
							},
						},
					},
					Files: []*File{
						{
							UUID:       "42143745-dba9-9c98-cdd7-0c521ae55118",
							Identifier: "1",
							Name:       "woodpigeon-pic.jpg",
							Size:       147004,
							Checksums: []Checksum{
								{
									Type:  "md5",
									Value: "53a64110e067b14394c142c09571bea0",
								},
							},
							StorageLocation: "s3://rdss-prod-figshare-0132/woodpigeon-pic.jpg",
							StorageType:     "S3",
						},
						{
							UUID:       "9102f719-2f78-d9c2-de12-15f27024f78b",
							Identifier: "2",
							Name:       "bird-sounds.mp3",
							Size:       910616,
							Checksums: []Checksum{
								{
									Type:  "md5",
									Value: "92c8ab01cecceb3bf0789c2cd8c7415a",
								},
							},
							StorageLocation: "s3://rdss-prod-figshare-0132/bird-sounds.mp3",
							StorageType:     "S3",
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
			json.Indent(&out, data, "", "  ")

			// Compare them
			have := out.Bytes()
			if !bytes.Equal(have, tc.data) {
				t.Errorf("Unexpected result:\nHAVE: `%s`\nEXPECTED: `%s`", have, tc.data)
			}
		})
	}
}
