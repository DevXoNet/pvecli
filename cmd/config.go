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

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Please enter the following configuration data:")

	cfg := &config.Config{}

	fmt.Print("API URL (e.g. https://proxmox.example.com:8006/api2/json): ")
	cfg.APIURL, _ = reader.ReadString('\n')
	cfg.APIURL = strings.TrimSpace(cfg.APIURL)

	fmt.Print("Token ID: ")
	cfg.TokenID, _ = reader.ReadString('\n')
	cfg.TokenID = strings.TrimSpace(cfg.TokenID)

	fmt.Print("Token Secret: ")
	cfg.TokenSecret, _ = reader.ReadString('\n')
	cfg.TokenSecret = strings.TrimSpace(cfg.TokenSecret)

	fmt.Print("Skip SSL verification (yes/no, default: no): ")
	insecure, _ := reader.ReadString('\n')
	cfg.InsecureSkipVerify = strings.ToLower(strings.TrimSpace(insecure)) == "yes"

	fmt.Print("Output format (json/text/yaml, default: json): ")
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

	// Save the configuration
	if err := config.SaveConfig(cfg); err != nil {
		return fmt.Errorf("error saving configuration: %w", err)
	}

	configPath, _ := config.GetConfigPath()
	fmt.Printf("\nConfiguration saved successfully to %s\n", configPath)
	return nil
}
