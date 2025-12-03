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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"pvecli/cmd/agent"
	"pvecli/cmd/cluster"
	"pvecli/cmd/instance"
	"pvecli/cmd/lxc"
	"pvecli/cmd/monitoring"
	"pvecli/cmd/oci"
	"pvecli/cmd/storage"
	"pvecli/cmd/task"
	"pvecli/cmd/vm"
	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"

	"github.com/spf13/cobra"
)

var debugFlag bool
var envFlag string

var rootCmd = &cobra.Command{
	Use:   "pvecli",
	Short: "Proxmox CLI tool",
	Long:  "pvecli - Simple CLI tool for managing Proxmox Cluster\n\nDeveloped by DevXo part of vByte Ltd",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.SetDebug(debugFlag)
		if envFlag != "" {
			config.SetEnvironment(envFlag)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugFlag, "debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().StringVarP(&envFlag, "env", "e", "", "Environment/cluster to use (prod, dev, staging, etc.)")

	// Initialize all command packages
	instance.Init(rootCmd)
	vm.Init(rootCmd)
	lxc.Init(rootCmd)
	agent.Init(rootCmd)
	cluster.Init(rootCmd)
	storage.Init(rootCmd)
	oci.Init(rootCmd)
	monitoring.Init(rootCmd)
	task.Init(rootCmd)

	// Config command is registered in config.go init() function
}

func Execute() {
	// Set debug mode before executing command
	config.SetDebug(debugFlag)

	// Hide fish completion command
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "completion" {
			for _, subCmd := range cmd.Commands() {
				if subCmd.Name() == "fish" {
					subCmd.Hidden = true
					break
				}
			}
			break
		}
	}

	if err := rootCmd.Execute(); err != nil {
		// Check if it's a friendly error (informational message)
		if friendlyErr, ok := err.(*pve.FriendlyError); ok {
			// Load config to get output format
			cfg, cfgErr := config.LoadConfig()
			outputFormat := config.OutputFormatText
			if cfgErr == nil {
				outputFormat = cfg.OutputFormat
			}

			// Parse the message to extract structured data
			msg := friendlyErr.Msg

			switch outputFormat {
			case config.OutputFormatJSON:
				// Parse message to extract vmid and action
				output := map[string]interface{}{
					"message": msg,
				}

				// Try to extract vmid and status from message
				if strings.Contains(msg, "instance") {
					parts := strings.Fields(msg)
					if len(parts) >= 2 {
						output["vmid"] = parts[1]
					}
					if strings.Contains(msg, "started") {
						output["action"] = "start"
						output["status"] = "success"
					} else if strings.Contains(msg, "stopped") {
						output["action"] = "stop"
						output["status"] = "success"
					} else if strings.Contains(msg, "already running") {
						output["status"] = "already_running"
					} else if strings.Contains(msg, "not running") {
						output["status"] = "not_running"
					}
				}

				b, _ := json.MarshalIndent(output, "", "  ")
				fmt.Println(string(b))

			case config.OutputFormatYAML:
				output := map[string]interface{}{
					"message": msg,
				}

				if strings.Contains(msg, "instance") {
					parts := strings.Fields(msg)
					if len(parts) >= 2 {
						output["vmid"] = parts[1]
					}
					if strings.Contains(msg, "started") {
						output["action"] = "start"
						output["status"] = "success"
					} else if strings.Contains(msg, "stopped") {
						output["action"] = "stop"
						output["status"] = "success"
					} else if strings.Contains(msg, "already running") {
						output["status"] = "already_running"
					} else if strings.Contains(msg, "not running") {
						output["status"] = "not_running"
					}
				}

				b, _ := yaml.Marshal(output)
				fmt.Print(string(b))

			case config.OutputFormatText:
				fmt.Println(msg)
			}
		} else {
			// Regular error - use output module
			output.PrintError(err)
		}
		os.Exit(1)
	}
}
