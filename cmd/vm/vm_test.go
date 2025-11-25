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

package vm

import (
	"testing"
)

func TestVMPackage(t *testing.T) {
	// Test that VM command is defined
	if vmCmd == nil {
		t.Error("vmCmd should not be nil")
	}
	
	// Test VM command name
	if vmCmd.Use != "vm" {
		t.Errorf("Expected vmCmd.Use to be 'vm', got '%s'", vmCmd.Use)
	}
	
	// Test that subcommands exist
	if cloneCmd == nil {
		t.Error("cloneCmd should not be nil")
	}
	if migrateCmd == nil {
		t.Error("migrateCmd should not be nil")
	}
	if diskCmd == nil {
		t.Error("diskCmd should not be nil")
	}
	if templateCmd == nil {
		t.Error("templateCmd should not be nil")
	}
	if suspendCmd == nil {
		t.Error("suspendCmd should not be nil")
	}
	if resumeCmd == nil {
		t.Error("resumeCmd should not be nil")
	}
}

func TestVMCommandNames(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want string
	}{
		{"VM command", vmCmd.Use, "vm"},
		{"Clone command", cloneCmd.Use, "clone <template-vmid>"},
		{"Migrate command", migrateCmd.Use, "migrate <vmid/ctid> --target <node>"},
		{"Disk command", diskCmd.Use, "disk"},
		{"Template command", templateCmd.Use, "template <vmid>"},
		{"Suspend command", suspendCmd.Use, "suspend <vmid>"},
		{"Resume command", resumeCmd.Use, "resume <vmid>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd != tt.want {
				t.Errorf("Command use = %v, want %v", tt.cmd, tt.want)
			}
		})
	}
}
