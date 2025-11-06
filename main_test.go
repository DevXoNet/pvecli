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

package main

import (
	"testing"
)

func TestMain(t *testing.T) {
	// Test that main package compiles without errors
	// This is a basic smoke test
	t.Log("Main package compiles successfully")
}

func TestVersionVariables(t *testing.T) {
	// Test that version variables exist and have default values

	if version == "" {
		t.Error("version should have a default value")
	}

	if build == "" {
		t.Error("build should have a default value")
	}

	// In tests, they should have default values
	if version != "dev" && version != "" {
		t.Logf("version is set to: %s", version)
	}

	if build != "unknown" && build != "" {
		t.Logf("build is set to: %s", build)
	}
}
