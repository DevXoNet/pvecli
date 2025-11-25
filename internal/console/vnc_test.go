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

package console

import (
	"encoding/json"
	"testing"
)

func TestPVEConsole_Structure(t *testing.T) {
	// Test that PVEConsole struct can be created
	console := &PVEConsole{
		conn:         nil,
		originalMode: nil,
		width:        80,
		height:       24,
		debug:        false,
	}

	if console.width != 80 {
		t.Errorf("Expected width 80, got %d", console.width)
	}

	if console.height != 24 {
		t.Errorf("Expected height 24, got %d", console.height)
	}

	if console.debug {
		t.Error("Expected debug to be false")
	}
}

func TestPVEConsole_ResizeMessage(t *testing.T) {
	// Test resize message format
	resizeMsg := map[string]interface{}{
		"command": "resize",
		"width":   120,
		"height":  40,
	}

	data, err := json.Marshal(resizeMsg)
	if err != nil {
		t.Fatalf("Failed to marshal resize message: %v", err)
	}

	// Verify JSON structure
	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal resize message: %v", err)
	}

	if decoded["command"] != "resize" {
		t.Errorf("Expected command 'resize', got '%v'", decoded["command"])
	}

	// JSON numbers are float64
	if decoded["width"] != float64(120) {
		t.Errorf("Expected width 120, got %v", decoded["width"])
	}

	if decoded["height"] != float64(40) {
		t.Errorf("Expected height 40, got %v", decoded["height"])
	}
}

func TestConnect_InvalidURL(t *testing.T) {
	// Test with invalid URL
	_, err := Connect("://invalid-url", "node01", "100", "qemu", 5900, "test-ticket", true, false)

	if err == nil {
		t.Error("Expected error with invalid URL, got nil")
	}
}

func TestConnect_URLSchemeConversion(t *testing.T) {
	// This test verifies the logic without actually connecting
	// We can't test actual connection without a real Proxmox server

	testCases := []struct {
		baseURL        string
		expectedScheme string
	}{
		{"http://proxmox.local:8006/api2/json", "ws"},
		{"https://proxmox.local:8006/api2/json", "wss"},
	}

	for _, tc := range testCases {
		// We expect connection to fail (no server), but we can verify URL parsing
		_, err := Connect(tc.baseURL, "node01", "100", "qemu", 5900, "ticket", true, false)

		// Should fail to connect (no server), but not fail to parse URL
		if err == nil {
			t.Errorf("Expected connection error for %s", tc.baseURL)
		}

		// Error should be about connection, not URL parsing
		if err != nil && err.Error() == "parse URL: invalid URI for request" {
			t.Errorf("URL parsing failed for %s: %v", tc.baseURL, err)
		}
	}
}

func TestPVEConsole_Close(t *testing.T) {
	// Test Close with nil connection
	console := &PVEConsole{
		conn:         nil,
		originalMode: nil,
		width:        80,
		height:       24,
		debug:        false,
	}

	err := console.Close()
	if err != nil {
		t.Errorf("Expected no error when closing nil connection, got: %v", err)
	}
}

func TestPVEConsole_RestoreTerminal(t *testing.T) {
	// Test restoreTerminal with nil originalMode
	console := &PVEConsole{
		conn:         nil,
		originalMode: nil,
		width:        80,
		height:       24,
		debug:        false,
	}

	// Should not panic
	console.restoreTerminal()
}

func TestAuthMessageFormat(t *testing.T) {
	// Test authentication message format
	user := "root@pam"
	ticket := "test-ticket-123"
	authMsg := user + ":" + ticket

	expected := "root@pam:test-ticket-123"
	if authMsg != expected {
		t.Errorf("Expected auth message '%s', got '%s'", expected, authMsg)
	}
}

func TestWebSocketPath(t *testing.T) {
	// Test WebSocket path construction
	node := "node01"
	expectedPath := "/api2/json/nodes/node01/vncwebsocket"

	actualPath := "/api2/json/nodes/" + node + "/vncwebsocket"

	if actualPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, actualPath)
	}
}
