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

package oci

import (
	"testing"
)

func TestOCIPackage(t *testing.T) {
	// Test that OCI command is defined
	if ociCmd == nil {
		t.Error("ociCmd should not be nil")
	}
	
	// Test OCI command name
	if ociCmd.Use != "oci" {
		t.Errorf("Expected ociCmd.Use to be 'oci', got '%s'", ociCmd.Use)
	}
	
	// Test that subcommands exist
	if ociTagsCmd == nil {
		t.Error("ociTagsCmd should not be nil")
	}
	if ociImagesCmd == nil {
		t.Error("ociImagesCmd should not be nil")
	}
}

func TestOCICommandNames(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want string
	}{
		{"OCI command", ociCmd.Use, "oci"},
		{"Tags command", ociTagsCmd.Use, "tags <node> <repository>"},
		{"Images command", ociImagesCmd.Use, "images [storage]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cmd != tt.want {
				t.Errorf("Command use = %v, want %v", tt.cmd, tt.want)
			}
		})
	}
}

func TestContainsSubstr(t *testing.T) {
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
			got := containsSubstr(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("containsSubstr(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}
