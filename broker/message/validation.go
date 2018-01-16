package message

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonreference"
	"github.com/xeipuuv/gojsonschema"

	"github.com/JiscRDSS/rdss-archivematica-channel-adapter/broker/message/specdata"
)

type Validator interface {
	// Validate returns the results of validating a document against its schema.
	// The error returned is non-nil when the validation process failed. If nil,
	// the caller still needs to check the validation state given by Result.
	Validate(msg *Message) (*gojsonschema.Result, error)

	// Validators returns the validators available.
	Validators() map[string]*gojsonschema.Schema
}

// rdssValidator implements Validator. It is the default validation solution for
// the RDSS API and it depends on its schema files.
type rdssValidator struct {
	validators map[string]*gojsonschema.Schema
}

var _ Validator = rdssValidator{}

// rdssPrefix is used by our schema loder so we can differentiate local
// references from external references. When rdssPrefix is matched the schema
// loader will load our internal schemas persisted in the specdata package.
// It is not a constant becuase we change it in our tests.
var rdssPrefix = "https://www.jisc.ac.uk/rdss/schema/"

var rdssSchemas = map[string]string{
	"MetadataCreateRequest":  "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/create/request_schema.json",
	"MetadataDeleteRequest":  "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/delete/request_schema.json",
	"MetadataReadRequest":    "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/read/request_schema.json",
	"MetadataReadResponse":   "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/read/response_schema.json",
	"MetadataUpdateRequest":  "https://www.jisc.ac.uk/rdss/schema/messages/body/metadata/update/request_schema.json",
	"VocabularyPatchRequest": "https://www.jisc.ac.uk/rdss/schema/messages/body/vocabulary/patch/request_schema.json",
	"VocabularyReadRequest":  "https://www.jisc.ac.uk/rdss/schema/messages/body/vocabulary/read/request_schema.json",
	"VocabularyReadResponse": "https://www.jisc.ac.uk/rdss/schema/messages/body/vocabulary/read/response_schema.json",
}

// schemaDocFinder is a function that given a reference like the ones found in
// `rdssSchema` returns the corresponding stream of bytes of the schema document
// that corresponds.
type schemaDocFinder func(string) ([]byte, error)

// DefaultSchemaDocFinder is used by localSchemaLoader to read the schema
// documents from the local store. The default function depends on the
// `specdata` package which is populated at build time. The tests use a custom
// function so it can populate arbitrary documents.
var DefaultSchemaDocFinder schemaDocFinder = rdssSchemaDocFinder

// rdssSchemaDocFinder fetches the schema documents from the `specdata`
// package.
func rdssSchemaDocFinder(source string) ([]byte, error) {
	return specdata.Asset(resolveSchemaRef(source))
}

// resolveSchemaRef knows the path of the schema assets in the specdata pkg.
func resolveSchemaRef(source string) string {
	source = strings.TrimSuffix(source, "/")
	source = strings.TrimPrefix(source, rdssPrefix)
	if !strings.HasPrefix(source, "messages/") {
		source = fmt.Sprintf("schemas/%s", source)
	}
	return source
}

// NewValidator returns a Validator with all the RDSS API schemas loaded.
func NewValidator() (Validator, error) {
	v := &rdssValidator{
		validators: make(map[string]*gojsonschema.Schema),
	}

	// Initialize schema validators.
	loader := NewLocalSchemaLoaderFactory()
	for name, path := range rdssSchemas {
		schema, err := gojsonschema.NewSchema(loader.New(path))
		if err != nil {
			return nil, errors.Wrapf(err, "schema loader could not create a schema object from %s", path)
		}
		v.validators[name] = schema
	}

	return v, nil
}

// Validate implementes Validator.
func (v rdssValidator) Validate(msg *Message) (*gojsonschema.Result, error) {
	loader := gojsonschema.NewBytesLoader(msg.body)
	val, ok := v.validators[msg.Type()]
	if !ok {
		return nil, fmt.Errorf("validator for %s does not exist", msg.Type())
	}
	return val.Validate(loader)
}

// Validators implementes Validator.
func (v rdssValidator) Validators() map[string]*gojsonschema.Schema {
	return v.validators
}

type localSchemaLoaderFactory struct {
	gojsonschema.JSONLoaderFactory
}

type localSchemaLoader struct {
	gojsonschema.JSONLoader
	source  string
	factory localSchemaLoaderFactory
}

func NewLocalSchemaLoaderFactory() localSchemaLoaderFactory {
	return localSchemaLoaderFactory{}
}

func (f localSchemaLoaderFactory) New(source string) gojsonschema.JSONLoader {
	return &localSchemaLoader{
		JSONLoader: gojsonschema.NewReferenceLoader(source),
		source:     source,
		factory:    f,
	}
}

func (l *localSchemaLoader) JsonSource() interface{} {
	return l.source
}

func (l *localSchemaLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return gojsonreference.NewJsonReference(l.source)
}

func (l *localSchemaLoader) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return l.factory
}

func (l *localSchemaLoader) LoadJSON() (interface{}, error) {
	ref, err := l.JsonReference()
	if err != nil {
		return nil, err
	}
	url := ref.GetUrl().String()
	if !strings.HasPrefix(url, rdssPrefix) {
		return l.JSONLoader.LoadJSON()
	}
	blob, err := DefaultSchemaDocFinder(url)
	if err != nil {
		return nil, err
	}
	return decodeJson(bytes.NewReader(blob))
}

func decodeJson(r io.Reader) (interface{}, error) {
	var document interface{}
	decoder := json.NewDecoder(r)
	decoder.UseNumber()
	if err := decoder.Decode(&document); err != nil {
		return nil, err
	}
	return document, nil
}

// NoOpValidator is a validator that recognizes all messages as valid.
type NoOpValidator struct{}

var _ Validator = NoOpValidator{}

// Validate implementes Validator.
func (v NoOpValidator) Validate(msg *Message) (*gojsonschema.Result, error) {
	return &gojsonschema.Result{}, nil
}

// Validators implementes Validator.
func (v NoOpValidator) Validators() map[string]*gojsonschema.Schema {
	return map[string]*gojsonschema.Schema{}
}
