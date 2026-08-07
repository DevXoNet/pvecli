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

// GetOutputFormat returns the command-line override or configured output format.
func GetOutputFormat() OutputFormat {
	if outputFormatOverride != "" {
		return outputFormatOverride
	}
	if cfg != nil && cfg.OutputFormat != "" {
		return cfg.OutputFormat
	}
	return OutputFormatText
}

// SetOutputFormat overrides the output format for the current command invocation.
func SetOutputFormat(format OutputFormat) error {
	switch format {
	case "":
		outputFormatOverride = ""
	case OutputFormatJSON, OutputFormatText, OutputFormatYAML:
		outputFormatOverride = format
	default:
		return fmt.Errorf("invalid output format %q (must be text, json, or yaml)", format)
	}
	return nil
}

// Global config instance
var cfg *Config
var debug bool
var currentEnv string
var outputFormatOverride OutputFormat

// ClusterConfig holds configuration for a single Proxmox cluster
type ClusterConfig struct {
	APIURL             string `yaml:"api_url"`
	TokenID            string `yaml:"token_id"`
	TokenSecret        string `yaml:"token_secret"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
	SSHUser            string `yaml:"ssh_user,omitempty"`
	SSHKeyPath         string `yaml:"ssh_key_path,omitempty"`
	SSHPort            int    `yaml:"ssh_port,omitempty"`
}

// Config holds the main configuration with multiple environments
type Config struct {
	Environments map[string]*ClusterConfig `yaml:"environments"`
	DefaultEnv   string                    `yaml:"default_env"`
	OutputFormat OutputFormat              `yaml:"output_format"`
	Debug        bool                      `yaml:"debug"`

	// Legacy fields for backward compatibility
	APIURL             string `yaml:"api_url,omitempty"`
	TokenID            string `yaml:"token_id,omitempty"`
	TokenSecret        string `yaml:"token_secret,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify,omitempty"`
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
		OutputFormat: OutputFormatText,
		Debug:        false,
		Environments: make(map[string]*ClusterConfig),
	}

	// Override with values from config file
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	// Handle legacy config format (backward compatibility)
	if cfg.APIURL != "" && cfg.TokenID != "" && cfg.TokenSecret != "" {
		// Convert legacy format to new format
		if len(cfg.Environments) == 0 {
			cfg.Environments["default"] = &ClusterConfig{
				APIURL:             cfg.APIURL,
				TokenID:            cfg.TokenID,
				TokenSecret:        cfg.TokenSecret,
				InsecureSkipVerify: cfg.InsecureSkipVerify,
			}
			cfg.DefaultEnv = "default"
		}
	}

	// Validate that we have at least one environment
	if len(cfg.Environments) == 0 {
		return nil, fmt.Errorf("no environments configured in %s", path)
	}

	// Set default environment if not specified
	if cfg.DefaultEnv == "" {
		// Use the first environment as default
		for env := range cfg.Environments {
			cfg.DefaultEnv = env
			break
		}
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

	// Clear legacy fields before saving
	cfg.APIURL = ""
	cfg.TokenID = ""
	cfg.TokenSecret = ""
	cfg.InsecureSkipVerify = false

	// API tokens are stored in this file, so it must only be readable by the owner.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer func() { _ = f.Close() }()
	if err := f.Chmod(0600); err != nil {
		return fmt.Errorf("secure config file permissions: %w", err)
	}

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// SetEnvironment sets the current environment to use
func SetEnvironment(env string) {
	currentEnv = env
}

// GetCurrentEnvironment returns the current environment name
func GetCurrentEnvironment() string {
	if currentEnv != "" {
		return currentEnv
	}
	if cfg != nil && cfg.DefaultEnv != "" {
		return cfg.DefaultEnv
	}
	return "default"
}

// GetCurrentCluster returns the ClusterConfig for the current environment
func GetCurrentCluster() (*ClusterConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config not loaded")
	}

	env := GetCurrentEnvironment()
	cluster, ok := cfg.Environments[env]
	if !ok {
		return nil, fmt.Errorf("environment '%s' not found in config", env)
	}

	return cluster, nil
}
