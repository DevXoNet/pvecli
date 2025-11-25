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
	"encoding/json"
	"fmt"
	"os"

	"pvecli/config"

	"gopkg.in/yaml.v3"
)

// Print outputs data according to the configured format
func Print(data interface{}) error {
	format := getOutputFormat()

	switch format {
	case config.OutputFormatJSON:
		return printJSON(data)
	case config.OutputFormatYAML:
		return printYAML(data)
	case config.OutputFormatText:
		return printText(data)
	default:
		return printJSON(data)
	}
}

// PrintError outputs an error according to the configured format
func PrintError(err error) {
	format := getOutputFormat()

	switch format {
	case config.OutputFormatJSON:
		output := map[string]interface{}{
			"error": err.Error(),
		}
		b, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(b))

	case config.OutputFormatYAML:
		output := map[string]interface{}{
			"error": err.Error(),
		}
		b, _ := yaml.Marshal(output)
		fmt.Print(string(b))

	case config.OutputFormatText:
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

// PrintSuccess outputs a success message according to the configured format
func PrintSuccess(message string, data map[string]interface{}) error {
	format := getOutputFormat()

	switch format {
	case config.OutputFormatJSON:
		if data == nil {
			data = make(map[string]interface{})
		}
		data["message"] = message
		data["status"] = "success"
		return printJSON(data)

	case config.OutputFormatYAML:
		if data == nil {
			data = make(map[string]interface{})
		}
		data["message"] = message
		data["status"] = "success"
		return printYAML(data)

	case config.OutputFormatText:
		fmt.Println(message)
		return nil
	}

	return nil
}

// getOutputFormat returns the configured output format
func getOutputFormat() config.OutputFormat {
	cfg, err := config.LoadConfig()
	if err != nil {
		return config.OutputFormatJSON
	}
	return cfg.OutputFormat
}

// printJSON outputs data as JSON
func printJSON(data interface{}) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

// printYAML outputs data as YAML
func printYAML(data interface{}) error {
	b, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal yaml: %w", err)
	}
	fmt.Print(string(b))
	return nil
}

// printText outputs data as text
// For text format, we expect data to be a string or have a String() method
func printText(data interface{}) error {
	switch v := data.(type) {
	case string:
		fmt.Println(v)
	case fmt.Stringer:
		fmt.Println(v.String())
	default:
		// Fallback to JSON for complex types
		return printJSON(data)
	}
	return nil
}

// Table represents tabular data for output
type Table struct {
	Headers []string
	Rows    [][]string
}

// PrintTable outputs a table according to the configured format
func PrintTable(table *Table) error {
	format := getOutputFormat()

	switch format {
	case config.OutputFormatJSON:
		// Convert table to JSON array of objects
		var data []map[string]interface{}
		for _, row := range table.Rows {
			rowData := make(map[string]interface{})
			for i, header := range table.Headers {
				if i < len(row) {
					rowData[header] = row[i]
				}
			}
			data = append(data, rowData)
		}
		return printJSON(data)

	case config.OutputFormatYAML:
		// Convert table to YAML array of objects
		var data []map[string]interface{}
		for _, row := range table.Rows {
			rowData := make(map[string]interface{})
			for i, header := range table.Headers {
				if i < len(row) {
					rowData[header] = row[i]
				}
			}
			data = append(data, rowData)
		}
		return printYAML(data)

	case config.OutputFormatText:
		// Print as simple text table
		// Print headers
		for i, header := range table.Headers {
			if i > 0 {
				fmt.Print("\t")
			}
			fmt.Print(header)
		}
		fmt.Println()

		// Print separator
		for i := range table.Headers {
			if i > 0 {
				fmt.Print("\t")
			}
			fmt.Print("---")
		}
		fmt.Println()

		// Print rows
		for _, row := range table.Rows {
			for i, cell := range row {
				if i > 0 {
					fmt.Print("\t")
				}
				fmt.Print(cell)
			}
			fmt.Println()
		}
	}

	return nil
}
