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

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// OutputFormat defines the output format type
type OutputFormat string

const (
	OutputFormatJSON OutputFormat = "json"
	OutputFormatText OutputFormat = "text"
	OutputFormatYAML OutputFormat = "yaml"
)

// Debug returns whether debug mode is enabled
func Debug() bool {
	if cfg != nil {
		return debug || cfg.Debug
	}
	return debug
}

// SetDebug sets the debug mode for the current session
func SetDebug(d bool) {
	debug = d
}

// GetOutputFormat returns the configured output format (defaults to JSON if not set)
func GetOutputFormat() OutputFormat {
	if cfg != nil && cfg.OutputFormat != "" {
		return cfg.OutputFormat
	}
	return OutputFormatJSON
}

// Global config instance
var cfg *Config
var debug bool

type Config struct {
	APIURL             string       `yaml:"api_url"`
	TokenID            string       `yaml:"token_id"`
	TokenSecret        string       `yaml:"token_secret"`
	InsecureSkipVerify bool         `yaml:"insecure_skip_verify"`
	OutputFormat       OutputFormat `yaml:"output_format"`
	Debug              bool         `yaml:"debug"`
}

func LoadConfig() (*Config, error) {
	if cfg != nil {
		return cfg, nil
	}

	// Get user home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home dir: %w", err)
	}

	// Read configuration file from ~/.devxo/pve.yaml
	path := filepath.Join(home, ".devxo", "pve.yaml")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	// Initialize config with default values
	cfg = &Config{
		OutputFormat: OutputFormatJSON,
		Debug:        false,
	}

	// Override with values from config file
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	// Validate required fields
	if cfg.APIURL == "" || cfg.TokenID == "" || cfg.TokenSecret == "" {
		return nil, fmt.Errorf("missing required fields in %s (api_url, token_id, token_secret)", path)
	}

	return cfg, nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	return filepath.Join(home, ".devxo", "pve.yaml"), nil
}

// SaveConfig saves the configuration to ~/.devxo/pve.yaml
func SaveConfig(cfg *Config) error {
	// Get user home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("home dir: %w", err)
	}

	// Create config directory if it doesn't exist
	configDir := filepath.Join(home, ".devxo")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	// Get the path to the configuration file
	path, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Save the configuration
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}
