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

	// Test with map (fallback to JSON)
	data := map[string]string{"test": "value"}
	err = printText(data)
	if err != nil {
		t.Errorf("printText with map failed: %v", err)
	}
}

func TestPrintError(t *testing.T) {
	err := fmt.Errorf("test error")
	
	// Should not panic
	PrintError(err)
}

func TestPrintTable(t *testing.T) {
	table := &Table{
		Headers: []string{"Name", "Status", "ID"},
		Rows: [][]string{
			{"test1", "running", "100"},
			{"test2", "stopped", "101"},
		},
	}

	err := PrintTable(table)
	if err != nil {
		t.Errorf("PrintTable failed: %v", err)
	}
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
