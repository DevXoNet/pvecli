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

package lxc

import (
	"testing"
)

func TestLXCPackage(t *testing.T) {
	// Test that LXC command is defined
	if lxcCmd == nil {
		t.Error("lxcCmd should not be nil")
	}

	// Test LXC command name
	if lxcCmd.Use != "lxc" {
		t.Errorf("Expected lxcCmd.Use to be 'lxc', got '%s'", lxcCmd.Use)
	}

	// Test that templates command exists
	if lxcTemplatesCmd == nil {
		t.Error("lxcTemplatesCmd should not be nil")
	}
}

func TestLXCCommandNames(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want string
	}{
		{"LXC command", lxcCmd.Use, "lxc"},
		{"Templates command", lxcTemplatesCmd.Use, "templates [storage]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd != tt.want {
				t.Errorf("Command use = %v, want %v", tt.cmd, tt.want)
			}
		})
	}
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{"Contains substring", "backup,vztmpl,iso", "vztmpl", true},
		{"Does not contain", "backup,iso", "vztmpl", false},
		{"Empty string", "", "test", false},
		{"Empty substring", "test", "", false},
		{"Exact match", "vztmpl", "vztmpl", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}
