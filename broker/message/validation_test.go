package message

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var sampleSchema = []byte(`{
	"$schema": "http://json-schema.org/draft-06/schema#",
	"type": "object",
	"properties": {
		"age": {
			"type": "integer",
			"minimum": 0
		}
	},
	"required": ["age"]
}`)

func createSchemaFile(t *testing.T, dir, path string, data []byte) {
	filename := filepath.Join(dir, path)
	if err := os.MkdirAll(filepath.Dir(filename), 0750); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(filename, data, 0640); err != nil {
		t.Fatal(err)
	}
}

func TestRDSSValidator(t *testing.T) {
	// rdssValidator depends on a bunch of schema files that are only made
	// available during the build process so we're going to populate our own.
	dir, err := ioutil.TempDir("", "schemasDir")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	for _, filename := range rdssSchemas {
		createSchemaFile(t, dir, filename, sampleSchema)
	}

	v, err := NewValidator(dir)
	if err != nil {
		t.Fatal(err)
	}

	msg := New(MessageTypeMetadataCreate, MessageClassCommand)
	testCases := []struct {
		data       []byte
		shouldFail bool
	}{
		{[]byte(`{}`), true},
		{[]byte(`{"age": -1}`), true},
		{[]byte(`{"age": "foobar"}`), true},
		{[]byte(`{"age": 10}`), false},
	}
	for _, tc := range testCases {
		msg.body = tc.data
		res, err := v.Validate(msg)
		if err != nil {
			t.Error(err)
		}
		if tc.shouldFail && res.Valid() {
			t.Error("validator failed to recognize invalid document")
		}
	}
}

func Test_convertJiscRDSSURI(t *testing.T) {
	tests := []struct {
		baseDir string
		source  string
		want    string
	}{
		{
			"/var",
			"https://www.jisc.ac.uk/rdss/schema/types.json",
			"file:///var/schemas/types.json",
		},
		{
			"/home/foobar/rdss-message-api-docs/hack/schemas",
			"https://www.jisc.ac.uk/rdss/schema/messages/header/header_schema.json",
			"file:///home/foobar/rdss-message-api-docs/hack/schemas/messages/header/header_schema.json",
		},
		{
			"/tmp",
			"https://www.jisc.ac.uk/rdss/schema/intellectual_asset.json/",
			"file:///tmp/schemas/intellectual_asset.json",
		},
	}
	for _, tt := range tests {
		if got := convertJiscRDSSURI(tt.baseDir, tt.source); got != tt.want {
			t.Errorf("convertJiscRDSSURI() = %v, want %v", got, tt.want)
		}
	}
}
