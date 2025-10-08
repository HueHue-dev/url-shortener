package models

import (
	"testing"
)

func TestGetShortCode(t *testing.T) {
	code := GetShortUrl()

	if code == "" {
		t.Error("GetShortCode() returned an empty string, expected a short code.")
	}

	if len(code) == 0 {
		t.Errorf("GetShortCode() returned empty code")
	}
}
