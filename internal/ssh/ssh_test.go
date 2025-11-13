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

package ssh

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetDefaultSSHKeyPath(t *testing.T) {
	keyPath := GetDefaultSSHKeyPath()
	
	if keyPath == "" {
		t.Error("GetDefaultSSHKeyPath returned empty string")
	}
	
	// Should contain .ssh directory
	if !filepath.IsAbs(keyPath) {
		t.Errorf("Expected absolute path, got: %s", keyPath)
	}
	
	// Should end with a key name
	base := filepath.Base(keyPath)
	validKeys := []string{"id_rsa", "id_ed25519", "id_ecdsa"}
	found := false
	for _, key := range validKeys {
		if base == key {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Expected key name to be one of %v, got: %s", validKeys, base)
	}
}

func TestGetDefaultSSHKeyPath_WithExistingKey(t *testing.T) {
	// Create temporary home directory
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}
	
	// Create a test key file
	keyFile := filepath.Join(sshDir, "id_rsa")
	err = os.WriteFile(keyFile, []byte("test-key-content"), 0600)
	if err != nil {
		t.Fatalf("Failed to create test key: %v", err)
	}
	
	// Temporarily override HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	keyPath := GetDefaultSSHKeyPath()
	
	if keyPath != keyFile {
		t.Errorf("Expected key path %s, got %s", keyFile, keyPath)
	}
}

func TestGetDefaultSSHKeyPath_PreferenceOrder(t *testing.T) {
	// Create temporary home directory
	tempDir := t.TempDir()
	sshDir := filepath.Join(tempDir, ".ssh")
	err := os.MkdirAll(sshDir, 0700)
	if err != nil {
		t.Fatalf("Failed to create .ssh directory: %v", err)
	}
	
	// Create multiple key files
	keys := []string{"id_rsa", "id_ed25519", "id_ecdsa"}
	for _, key := range keys {
		keyFile := filepath.Join(sshDir, key)
		err = os.WriteFile(keyFile, []byte("test-key"), 0600)
		if err != nil {
			t.Fatalf("Failed to create test key %s: %v", key, err)
		}
	}
	
	// Temporarily override HOME
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)
	
	keyPath := GetDefaultSSHKeyPath()
	
	// Should prefer id_rsa (first in the list)
	expectedPath := filepath.Join(sshDir, "id_rsa")
	if keyPath != expectedPath {
		t.Errorf("Expected key path %s, got %s", expectedPath, keyPath)
	}
}

func TestNewClient_NoAuthMethods(t *testing.T) {
	// Try to create client with non-existent key
	_, err := NewClient("localhost", "testuser", "/nonexistent/key", 22)
	
	if err == nil {
		t.Error("Expected error when no auth methods available, got nil")
	}
	
	if err != nil && err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestNewClient_InvalidKeyPath(t *testing.T) {
	// Create a temporary invalid key file
	tempDir := t.TempDir()
	invalidKey := filepath.Join(tempDir, "invalid_key")
	err := os.WriteFile(invalidKey, []byte("not-a-valid-ssh-key"), 0600)
	if err != nil {
		t.Fatalf("Failed to create invalid key: %v", err)
	}
	
	// Try to create client with invalid key
	_, err = NewClient("localhost", "testuser", invalidKey, 22)
	
	if err == nil {
		t.Error("Expected error with invalid key, got nil")
	}
}

func TestClient_Structure(t *testing.T) {
	// Test that Client struct has expected fields
	client := &Client{
		client: nil,
		host:   "test.example.com",
	}
	
	if client.host != "test.example.com" {
		t.Errorf("Expected host 'test.example.com', got '%s'", client.host)
	}
}
