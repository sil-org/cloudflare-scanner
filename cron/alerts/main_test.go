package alerts

import (
	"testing"
	"strings"
)

func TestGetCFRecordsOK(t *testing.T) {
	config := AlertsConfig {
		CFApiKey: "abc123",
		CFApiEmail: "cio@domain1.org",
		CFZoneNames: []string{"domain1.org","domain2.org"},
		CFContainsStrings: []string{"outdated"},
		SESAWSRegion: "",
		SESCharSet: "",
		SESReturnToAddr: "no_reply@domain1.org",
		SESSubjectText: "Outdated Cloudflare records",
		RecipientEmails: []string{"cio@domain1.org", "it-guy@domain1.org"},
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

func TestGetCFRecordsMissingRequired(t *testing.T) {
	config := AlertsConfig {
		CFApiKey: "abc123",
	}

	var err error
	if err = config.init(); err == nil {
		t.Errorf("Expected an error, but didn't get one.")
		return
	}

	results := err.Error()
	expected := "required"
	if !strings.Contains(results, expected) {
		t.Errorf("The message of the error thrown was not correct.\nExpected it to contain: %s. Got '%s'", expected, results)
	}
}
