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

package instance

import (
	"testing"
)

// Basic test to ensure package compiles
func TestInstancePackage(t *testing.T) {
	// Test that commands are defined
	if startCmd == nil {
		t.Error("startCmd should not be nil")
	}
	if stopCmd == nil {
		t.Error("stopCmd should not be nil")
	}
	if statusCmd == nil {
		t.Error("statusCmd should not be nil")
	}
	if infoCmd == nil {
		t.Error("infoCmd should not be nil")
	}
	if novncCmd == nil {
		t.Error("novncCmd should not be nil")
	}
}

// Test command names
func TestCommandNames(t *testing.T) {
	tests := []struct {
		name    string
		cmd     string
		want    string
	}{
		{"Start command", startCmd.Use, "start <vmid/ctid>"},
		{"Stop command", stopCmd.Use, "stop <vmid/ctid>"},
		{"Status command", statusCmd.Use, "status <vmid/ctid>"},
		{"Info command", infoCmd.Use, "info <vmid/ctid>"},
		{"NoVNC command", novncCmd.Use, "novnc <vmid>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd != tt.want {
				t.Errorf("Command use = %v, want %v", tt.cmd, tt.want)
			}
		})
	}
}
