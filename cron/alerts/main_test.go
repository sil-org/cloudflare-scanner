package main

import (
	"log"
	"testing"
)

func TestGetCFRecords(t *testing.T) {
	t.Skip("Only run this in local development")

	config, err := newScanner()
	if err != nil {
		t.Errorf("Failed initializing config for test: %v", err)
		return
	}

	for _, alertConfig := range config.Alerts {
		cfRecords := getCFRecords(*config, alertConfig)
		if len(cfRecords) < 1 {
			log.Printf("\n No records found in Cloudflare containing any of these: %v", alertConfig.CFContainsStrings)
		} else {
			log.Printf("\n %d records found containing: %v", len(cfRecords), alertConfig.CFContainsStrings)
		}
	}
}
