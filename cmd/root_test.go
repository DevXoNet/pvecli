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

package cmd

import (
	"testing"
)

func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	if rootCmd.Use != "pvecli" {
		t.Errorf("Expected Use 'pvecli', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short == "" {
		t.Error("Expected Short description to be set")
	}

	if rootCmd.Long == "" {
		t.Error("Expected Long description to be set")
	}
}

func TestDebugFlag(t *testing.T) {
	// Test that debug flag exists
	flag := rootCmd.PersistentFlags().Lookup("debug")
	if flag == nil {
		t.Fatal("Debug flag not found")
	}

	if flag.Shorthand != "d" {
		t.Errorf("Expected debug flag shorthand 'd', got '%s'", flag.Shorthand)
	}
}
