package main

import (
	"encoding/json"
	"log/slog"
	"testing"
)

func TestGetCFRecords(t *testing.T) {
	t.Skip("Only run this in local development")

	scanner, err := newScanner()
	if err != nil {
		t.Errorf("Failed initializing scanner for test: %v", err)
		return
	}

	for _, alert := range scanner.Alerts {
		cfRecords := alert.getCFRecords()
		if len(cfRecords) < 1 {
			slog.Info("no records found in Cloudflare", "strings", alert.CFContainsStrings)
		} else {
			slog.Info("records found", "count", len(cfRecords), "strings", alert.CFContainsStrings)
		}
	}
}

func TestHandler(t *testing.T) {
	t.Skip("Only run this in local development")

	err := handler()
	if err != nil {
		t.Errorf("%s", err)
	}
}

func TestReadFromParameterStore(t *testing.T) {
	t.Skip("Only run this in local development")

	cfg := readFromParameterStore("/cloudflare-scanner/prod/config")
	if cfg == "" {
		t.Errorf("No configuration file found")
	}

	var scanner Scanner
	err := json.Unmarshal([]byte(cfg), &scanner)
	if err != nil {
		t.Errorf("could not unmarshal the configuration into Scanner: %v", err)
	}

	if len(scanner.Alerts) == 0 {
		t.Errorf("No alerts found")
	}
}
