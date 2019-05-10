package leif

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/qri-io/jsonschema"
	"regexp"
)

// Versioning of the schema files (semvar)
const major = 0
const minor = 1
const patch = 0

// Store version
var Version = fmt.Sprintf("%d.%d.%d", major, minor, patch)

// Validate an input json against the defined json schema. Json schema's are also
// tagged based on the version of sherpa they apply too. Should you want to pin to a
// version of sherpa, simply use SetValidationVersion('x.y.z').
func Validate(input []byte) error {

	var schema string
	var ok bool
	if schema, ok = schemas[Version]; !ok {
		return errors.New("Can't load schema")
	}

	rs := &jsonschema.RootSchema{}
	if err := json.Unmarshal([]byte(schema), rs); err != nil {
		return err
	}

	if errors, _ := rs.ValidateBytes(input); len(errors) > 0 {
		return errors[0]
	}
	return nil
}

// Set the specific version of validation JSON to be used. Should be given in form
// 'x.y.z'. Must be given valid version.
func SetValidationVersion(v string) error {
	_, err := regexp.MatchString(`^[0-9]+\.[0-9]+\.[0-9]+$`, v)
	if err != nil {
		return errors.New("Version string not in form x.y.z")
	}

	if _, ok := schemas[v]; !ok {
		return errors.New("Invalid version given")
	}

	Version = v
	return nil
}
