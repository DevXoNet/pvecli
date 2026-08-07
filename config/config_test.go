// Copyright 2025 DevXo part of vByte Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_MissingFile(t *testing.T) {
	// Set a non-existent config path
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)

	_, err := LoadConfig()
	if err == nil {
		t.Error("Expected error when config file doesn't exist, got nil")
	}
}

func TestLoadConfig_ValidConfig(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, ".devxo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	configFile := filepath.Join(configDir, "pve.yaml")
	configContent := `api_url: https://test.example.com:8006/api2/json
token_id: test@pam!test
token_secret: test-secret-123
insecure_skip_verify: true
`
	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Set HOME to temp dir
	t.Setenv("HOME", tempDir)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.APIURL != "https://test.example.com:8006/api2/json" {
		t.Errorf("Expected APIURL 'https://test.example.com:8006/api2/json', got '%s'", cfg.APIURL)
	}
	if cfg.TokenID != "test@pam!test" {
		t.Errorf("Expected TokenID 'test@pam!test', got '%s'", cfg.TokenID)
	}
	if cfg.TokenSecret != "test-secret-123" {
		t.Errorf("Expected TokenSecret 'test-secret-123', got '%s'", cfg.TokenSecret)
	}
	if !cfg.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be true")
	}
}

func TestDebug(t *testing.T) {
	// Test default (should be false)
	if Debug() {
		t.Error("Expected Debug to be false by default")
	}

	// Test with debug flag set
	SetDebug(true)
	if !Debug() {
		t.Error("Expected Debug to be true after SetDebug(true)")
	}

	// Reset
	SetDebug(false)
	if Debug() {
		t.Error("Expected Debug to be false after SetDebug(false)")
	}
}
