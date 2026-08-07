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

package output

import (
	"fmt"
	"strings"
	"testing"
)

func TestPrintJSON(t *testing.T) {
	data := map[string]interface{}{
		"name":   "test",
		"status": "success",
	}

	err := printJSON(data)
	if err != nil {
		t.Errorf("printJSON failed: %v", err)
	}
}

func TestPrintYAML(t *testing.T) {
	data := map[string]interface{}{
		"name":   "test",
		"status": "success",
	}

	err := printYAML(data)
	if err != nil {
		t.Errorf("printYAML failed: %v", err)
	}
}

func TestPrintText(t *testing.T) {
	// Test with string
	err := printText("test message")
	if err != nil {
		t.Errorf("printText with string failed: %v", err)
	}

	// Structured values must render as human-readable text, not JSON.
	var output strings.Builder
	data := map[string]interface{}{
		"status": "running",
		"items":  []interface{}{map[string]interface{}{"id": 100, "name": "vm"}},
	}
	if err := renderText(&output, data, 0); err != nil {
		t.Fatalf("renderText failed: %v", err)
	}
	got := output.String()
	for _, expected := range []string{"status: running", "items:", "id: 100", "name: vm"} {
		if !strings.Contains(got, expected) {
			t.Errorf("renderText output %q does not contain %q", got, expected)
		}
	}
	if strings.HasPrefix(strings.TrimSpace(got), "{") {
		t.Errorf("renderText unexpectedly returned JSON: %s", got)
	}
}

func TestPrintError(t *testing.T) {
	err := fmt.Errorf("test error")

	// Should not panic
	PrintError(err)
}

func TestPrintSuccess(t *testing.T) {
	data := map[string]interface{}{
		"vmid": "100",
	}

	err := PrintSuccess("Operation completed", data)
	if err != nil {
		t.Errorf("PrintSuccess failed: %v", err)
	}
}
