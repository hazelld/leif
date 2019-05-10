package leif_test

import (
	. "leif"
	"testing"
)

func TestSetInvalidVersion(t *testing.T) {

	// Set non-version string
	err := SetValidationVersion("notversionstring")
	if err == nil {
		t.Errorf("SetValidationVersion() allowed non-version string")
	}

	// Set version that doesn't exist
	err = SetValidationVersion("25.0.0")
	if err == nil {
		t.Errorf("SetValidationVersion() allowed non-existant version")
	}
}

func TestSetValidVersion(t *testing.T) {
	err := SetValidationVersion("0.1.0")
	if err != nil {
		t.Errorf("SetValidationVersion() didn't allow valid version")
	}
}
