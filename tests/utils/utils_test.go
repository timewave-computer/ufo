package utils

import (
	"testing"
)

func TestDefaultTestConfig(t *testing.T) {
	config := DefaultTestConfig("fauxmosis-comet")

	// Check that the config has the expected values
	if config.BinaryType != "fauxmosis-comet" {
		t.Errorf("Expected BinaryType to be 'fauxmosis-comet', got '%s'", config.BinaryType)
	}

	if config.RPCAddress != "tcp://localhost:26657" {
		t.Errorf("Expected RPCAddress to be 'tcp://localhost:26657', got '%s'", config.RPCAddress)
	}

	if config.BlockTimeMS != 1000 {
		t.Errorf("Expected BlockTimeMS to be 1000, got %d", config.BlockTimeMS)
	}
}

func TestHTTPClientCreation(t *testing.T) {
	client := NewHTTPClient("http://localhost:1317")

	if client == nil {
		t.Error("Expected non-nil HTTP client")
	}

	if client.baseURL != "http://localhost:1317" {
		t.Errorf("Expected baseURL to be 'http://localhost:1317', got '%s'", client.baseURL)
	}
}
