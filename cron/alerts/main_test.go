package main

import (
	"log"
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
			log.Printf("No records found in Cloudflare containing any of these: %v", alert.CFContainsStrings)
		} else {
			log.Printf("%d records found containing: %v", len(cfRecords), alert.CFContainsStrings)
		}
	}
}

func TestHandler(t *testing.T) {
	err := handler()
	if err != nil {
		t.Errorf("%s", err)
	}
}
