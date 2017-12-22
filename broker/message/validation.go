package message

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonreference"
	"github.com/xeipuuv/gojsonschema"
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

var rdssSchemas = map[string]string{
	"MetadataCreateRequest":  "messages/body/metadata/create/request_schema.json",
	"MetadataDeleteRequest":  "messages/body/metadata/delete/request_schema.json",
	"MetadataReadRequest":    "messages/body/metadata/read/request_schema.json",
	"MetadataReadResponse":   "messages/body/metadata/read/response_schema.json",
	"MetadataUpdateRequest":  "messages/body/metadata/update/request_schema.json",
	"VocabularyPatchRequest": "messages/body/vocabulary/patch/request_schema.json",
	"VocabularyReadRequest":  "messages/body/vocabulary/read/request_schema.json",
	"VocabularyReadResponse": "messages/body/vocabulary/read/response_schema.json",
}

// NewValidator returns a Validator designed for the RDSS API.
func NewValidator(schemasDir string) (Validator, error) {
	v := &rdssValidator{
		validators: make(map[string]*gojsonschema.Schema),
	}

	// Initialize schema validators.
	for name, path := range rdssSchemas {
		loader := newReferenceLoader(schemasDir, path)
		schema, err := gojsonschema.NewSchema(loader)
		if err != nil {
			return nil, err
		}
		v.validators[name] = schema
	}

	return v, nil
}

func (v rdssValidator) Validate(msg *Message) (*gojsonschema.Result, error) {
	loader := gojsonschema.NewBytesLoader(msg.body)
	val, ok := v.validators[msg.Type()]
	if !ok {
		return nil, fmt.Errorf("validator for %s does not exist", msg.Type())
	}
	return val.Validate(loader)
}

func (v rdssValidator) Validators() map[string]*gojsonschema.Schema {
	return v.validators
}

// loaderFactory is a schema loader factory for JiscRDSS API.
type loaderFactory struct {
	baseDir string

	// Default JSONLoader provided by gojsonschema
	defaultLoader gojsonschema.JSONLoader
}

func (l loaderFactory) New(source string) gojsonschema.JSONLoader {
	source = convertJiscRDSSURI(l.baseDir, source)
	return l.defaultLoader.LoaderFactory().New(source)
}

// convertJiscRDSSURI resolves JSON References used in the RDSS message API so
// we can load the documents from disk. It is simlar to the
// jsonchema.RefResolver created in the API test suite (https://git.io/vNe16)
// but it avoids to list all the pairs explicitly.
func convertJiscRDSSURI(baseDir, source string) string {
	const prefix = "https://www.jisc.ac.uk/rdss/schema"
	if !strings.HasPrefix(source, prefix) {
		return source
	}
	source = strings.TrimSuffix(source, "/")
	source = strings.TrimPrefix(source, prefix)
	var path string
	if strings.HasPrefix(source, "/messages/") {
		path = filepath.Join(baseDir, source)
	} else {
		path = filepath.Join(baseDir, "schemas", source)
	}
	source = fmt.Sprintf("file://%s", path)
	return source
}

// referenceLoader is our custom JSON reference loader that relies in our custom
// loader factory.
type referenceLoader struct {
	baseDir string
	loader  gojsonschema.JSONLoader
}

func newReferenceLoader(baseDir string, path string) *referenceLoader {
	path = fmt.Sprintf("file://%s", filepath.Join(baseDir, path))
	return &referenceLoader{
		baseDir: baseDir,
		loader:  gojsonschema.NewReferenceLoader(path),
	}
}

func (l *referenceLoader) JsonSource() interface{} {
	return l.loader.JsonSource()
}

func (l *referenceLoader) JsonReference() (gojsonreference.JsonReference, error) {
	return l.loader.JsonReference()
}

func (l *referenceLoader) LoaderFactory() gojsonschema.JSONLoaderFactory {
	return &loaderFactory{l.baseDir, l.loader}
}

func (l *referenceLoader) LoadJSON() (interface{}, error) {
	return l.loader.LoadJSON()
}

// NoOpValidator is an implementation of Validator that validates all the
// messages.
type NoOpValidator struct{}

func (v NoOpValidator) Validate(*Message) (*gojsonschema.Result, error) {
	return &gojsonschema.Result{}, nil
}

func (v NoOpValidator) Validators() map[string]*gojsonschema.Schema {
	return map[string]*gojsonschema.Schema{}
}
