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
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

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
	_, err := config.LoadConfig()
	if err != nil {
		return config.GetOutputFormat()
	}
	return config.GetOutputFormat()
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

// printText outputs strings directly and renders structured values as readable,
// deterministic key/value lists instead of falling back to JSON.
func printText(data interface{}) error {
	switch v := data.(type) {
	case string:
		fmt.Println(v)
		return nil
	case fmt.Stringer:
		fmt.Println(v.String())
		return nil
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("prepare text output: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var normalized interface{}
	if err := decoder.Decode(&normalized); err != nil {
		return fmt.Errorf("prepare text output: %w", err)
	}
	return renderText(os.Stdout, normalized, 0)
}

func renderText(writer io.Writer, value interface{}, indent int) error {
	prefix := strings.Repeat(" ", indent)
	switch value := value.(type) {
	case map[string]interface{}:
		keys := make([]string, 0, len(value))
		for key := range value {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		if len(keys) == 0 {
			_, err := fmt.Fprintf(writer, "%s(none)\n", prefix)
			return err
		}
		for _, key := range keys {
			item := value[key]
			if isScalar(item) {
				if _, err := fmt.Fprintf(writer, "%s%s: %s\n", prefix, key, scalarText(item)); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(writer, "%s%s:\n", prefix, key); err != nil {
				return err
			}
			if err := renderText(writer, item, indent+2); err != nil {
				return err
			}
		}
	case []interface{}:
		if len(value) == 0 {
			_, err := fmt.Fprintf(writer, "%s(none)\n", prefix)
			return err
		}
		for _, item := range value {
			if isScalar(item) {
				if _, err := fmt.Fprintf(writer, "%s- %s\n", prefix, scalarText(item)); err != nil {
					return err
				}
				continue
			}
			if _, err := fmt.Fprintf(writer, "%s-\n", prefix); err != nil {
				return err
			}
			if err := renderText(writer, item, indent+2); err != nil {
				return err
			}
		}
	default:
		_, err := fmt.Fprintf(writer, "%s%s\n", prefix, scalarText(value))
		return err
	}
	return nil
}

func isScalar(value interface{}) bool {
	switch value.(type) {
	case nil, string, bool, json.Number,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}

func scalarText(value interface{}) string {
	if value == nil {
		return "null"
	}
	return fmt.Sprint(value)
}
