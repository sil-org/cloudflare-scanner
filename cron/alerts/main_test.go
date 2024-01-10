package main

import (
	"testing"
)

func TestGetCFRecordsOK(t *testing.T) {
	config := AlertsConfig{
		CFApiToken:        "abc123",
		CFZoneNames:       []string{"domain1.org", "domain2.org"},
		CFContainsStrings: []string{"outdated"},
		SESAWSRegion:      "",
		SESCharSet:        "",
		SESReturnToAddr:   "no_reply@domain1.org",
		SESSubjectText:    "Outdated Cloudflare records",
		RecipientEmails:   []string{"cio@domain1.org", "it-guy@domain1.org"},
	}

	if err := config.init(); err != nil {
		t.Errorf("Did not expect an error, but got one: %v", err.Error())
		return
	}

	results := config.SESCharSet
	expected := SESCharSet

	if results != expected {
		t.Errorf("Did not get default SESCharSet. Expected: %s, but got: %s", expected, results)
	}
}

func TestGetCFRecords(t *testing.T) {
	t.Skip("Only run this in local development")

	// Just initialize config from .env file
	config := AlertsConfig{}

	if err := config.init(); err != nil {
		t.Errorf("Failed initializing config for test: %v", err)
		return
	}

	_ = getCFRecords(config)
}
