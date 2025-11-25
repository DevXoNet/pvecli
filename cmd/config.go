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
	"bufio"
	"fmt"
	"os"
	"strings"

	"pvecli/config"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure pvecli connection to Proxmox cluster",
	Long:  "Interactive setup wizard to configure API URL, authentication token, and other settings.",
	RunE:  runConfig,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured environments",
	RunE:  runConfigList,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get configuration value",
	Long:  "Get configuration value. If no key specified, shows all settings.\n\nAvailable keys: output_format, default_env, debug",
	RunE:  runConfigGet,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set configuration value",
	Long:  "Set configuration value.\n\nAvailable keys:\n  output_format (json|yaml|text)\n  default_env (environment name)\n  debug (true|false)",
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}

func runConfigList(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Println("=== Configured Environments ===")
	fmt.Println()

	if len(cfg.Environments) == 0 {
		fmt.Println("No environments configured.")
		fmt.Println("Run 'pvecli config' to add an environment.")
		return nil
	}

	for envName, cluster := range cfg.Environments {
		isDefault := ""
		if envName == cfg.DefaultEnv {
			isDefault = " (default)"
		}
		fmt.Printf("• %s%s\n", envName, isDefault)
		fmt.Printf("  API URL: %s\n", cluster.APIURL)
		fmt.Printf("  Token ID: %s\n", cluster.TokenID)
		fmt.Println()
	}

	fmt.Printf("Default environment: %s\n", cfg.DefaultEnv)
	fmt.Printf("Output format: %s\n", cfg.OutputFormat)

	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// If no key specified, show all settings
	if len(args) == 0 {
		fmt.Println("=== Configuration Settings ===")
		fmt.Printf("Output format: %s\n", cfg.OutputFormat)
		fmt.Printf("Default environment: %s\n", cfg.DefaultEnv)
		fmt.Printf("Debug: %v\n", cfg.Debug)
		return nil
	}

	key := args[0]
	switch key {
	case "output_format":
		fmt.Println(cfg.OutputFormat)
	case "default_env":
		fmt.Println(cfg.DefaultEnv)
	case "debug":
		fmt.Println(cfg.Debug)
	default:
		return fmt.Errorf("unknown key: %s\n\nAvailable keys: output_format, default_env, debug", key)
	}

	return nil
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	key := args[0]
	value := args[1]

	switch key {
	case "output_format":
		switch strings.ToLower(value) {
		case "json":
			cfg.OutputFormat = config.OutputFormatJSON
		case "yaml":
			cfg.OutputFormat = config.OutputFormatYAML
		case "text":
			cfg.OutputFormat = config.OutputFormatText
		default:
			return fmt.Errorf("invalid output format: %s (must be json, yaml, or text)", value)
		}
		fmt.Printf("✓ Output format set to: %s\n", cfg.OutputFormat)

	case "default_env":
		if _, exists := cfg.Environments[value]; !exists {
			return fmt.Errorf("environment '%s' does not exist\n\nRun 'pvecli config list' to see available environments", value)
		}
		cfg.DefaultEnv = value
		fmt.Printf("✓ Default environment set to: %s\n", value)

	case "debug":
		switch strings.ToLower(value) {
		case "true", "1", "yes":
			cfg.Debug = true
		case "false", "0", "no":
			cfg.Debug = false
		default:
			return fmt.Errorf("invalid debug value: %s (must be true or false)", value)
		}
		fmt.Printf("✓ Debug set to: %v\n", cfg.Debug)

	default:
		return fmt.Errorf("unknown key: %s\n\nAvailable keys: output_format, default_env, debug", key)
	}

	// Save the configuration
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	return nil
}

func runConfig(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	// Try to load existing config
	cfg, err := config.LoadConfig()
	if err != nil {
		// Create new config if it doesn't exist
		cfg = &config.Config{
			Environments: make(map[string]*config.ClusterConfig),
			OutputFormat: config.OutputFormatJSON,
			Debug:        false,
		}
	}

	fmt.Println("=== Proxmox CLI Configuration ===")
	fmt.Println()

	// Ask for environment name
	fmt.Print("Environment name (e.g., prod, dev, staging): ")
	envName, _ := reader.ReadString('\n')
	envName = strings.TrimSpace(envName)
	if envName == "" {
		envName = "default"
	}

	// Check if environment already exists
	if _, exists := cfg.Environments[envName]; exists {
		fmt.Printf("\nEnvironment '%s' already exists. Overwrite? (yes/no): ", envName)
		overwrite, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(overwrite)) != "yes" {
			fmt.Println("Configuration cancelled.")
			return nil
		}
	}

	cluster := &config.ClusterConfig{}

	fmt.Print("API URL (IP or hostname, e.g. 192.168.1.10 or pve.example.com): ")
	apiInput, _ := reader.ReadString('\n')
	apiInput = strings.TrimSpace(apiInput)

	// Auto-format API URL
	if !strings.HasPrefix(apiInput, "http://") && !strings.HasPrefix(apiInput, "https://") {
		// Remove any trailing slashes or paths
		apiInput = strings.TrimRight(apiInput, "/")
		// Add https:// prefix and standard Proxmox API path
		cluster.APIURL = fmt.Sprintf("https://%s:8006/api2/json", apiInput)
	} else {
		cluster.APIURL = apiInput
	}

	fmt.Print("Token ID (e.g. root@pam!mytoken): ")
	cluster.TokenID, _ = reader.ReadString('\n')
	cluster.TokenID = strings.TrimSpace(cluster.TokenID)

	fmt.Print("Token Secret: ")
	cluster.TokenSecret, _ = reader.ReadString('\n')
	cluster.TokenSecret = strings.TrimSpace(cluster.TokenSecret)

	fmt.Print("Skip SSL verification (yes/no, default: yes): ")
	insecure, _ := reader.ReadString('\n')
	insecure = strings.ToLower(strings.TrimSpace(insecure))
	// Default to yes (skip verification)
	if insecure == "" || insecure == "yes" || insecure == "y" {
		cluster.InsecureSkipVerify = true
	} else {
		cluster.InsecureSkipVerify = false
	}

	// SSH settings (optional)
	fmt.Println("\n--- SSH Settings (optional, for cluster update command) ---")

	fmt.Print("SSH User (default: root): ")
	sshUser, _ := reader.ReadString('\n')
	sshUser = strings.TrimSpace(sshUser)
	if sshUser != "" {
		cluster.SSHUser = sshUser
	}

	fmt.Print("SSH Key Path (default: ~/.ssh/id_rsa): ")
	sshKey, _ := reader.ReadString('\n')
	sshKey = strings.TrimSpace(sshKey)
	if sshKey != "" {
		cluster.SSHKeyPath = sshKey
	}

	fmt.Print("SSH Port (default: 22): ")
	sshPortStr, _ := reader.ReadString('\n')
	sshPortStr = strings.TrimSpace(sshPortStr)
	if sshPortStr != "" {
		var sshPort int
		fmt.Sscanf(sshPortStr, "%d", &sshPort)
		if sshPort > 0 {
			cluster.SSHPort = sshPort
		}
	}

	// Add cluster to environments
	cfg.Environments[envName] = cluster

	// Set as default if it's the first environment
	if len(cfg.Environments) == 1 || cfg.DefaultEnv == "" {
		cfg.DefaultEnv = envName
		fmt.Printf("\nSetting '%s' as default environment.\n", envName)
	} else {
		fmt.Printf("\nSet '%s' as default environment? (yes/no, current: %s): ", envName, cfg.DefaultEnv)
		setDefault, _ := reader.ReadString('\n')
		if strings.ToLower(strings.TrimSpace(setDefault)) == "yes" {
			cfg.DefaultEnv = envName
		}
	}

	// Ask about output format only if not set
	if cfg.OutputFormat == "" {
		fmt.Print("\nOutput format (json/text/yaml, default: json): ")
		format, _ := reader.ReadString('\n')
		format = strings.ToLower(strings.TrimSpace(format))
		switch format {
		case "text":
			cfg.OutputFormat = config.OutputFormatText
		case "yaml":
			cfg.OutputFormat = config.OutputFormatYAML
		case "", "json":
			cfg.OutputFormat = config.OutputFormatJSON
		default:
			fmt.Println("Invalid format, using default (json)")
			cfg.OutputFormat = config.OutputFormatJSON
		}
	}

	// Save the configuration
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("\n✓ Configuration saved successfully to %s\n", configPath)
	fmt.Printf("✓ Environment '%s' configured\n", envName)
	fmt.Printf("\nUse: pvecli --env %s <command>\n", envName)
	if cfg.DefaultEnv == envName {
		fmt.Println("Or simply: pvecli <command> (default environment)")
	}

	return nil
}
