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

	"pvecli/config"
	"pvecli/internal/proxmox"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var statusCmd = &cobra.Command{
	Use:   "status <vmid/ctid>",
	Short: "Show current status of VM or container",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := proxmox.NewClient(cfg)
		vmid := args[0]

		// Find which node has this VM/CT and its type
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		// Get status
		status, err := client.GetInstanceStatus(node, vmid, instanceType)
		if err != nil {
			return err
		}

		// Extract key information
		name := status["name"]
		statusStr := status["status"]
		vmType := instanceType
		if vmType == "ct" {
			vmType = "LXC"
		} else {
			vmType = "VM"
		}

		// Output based on configured format
		outputFormat := config.GetOutputFormat()

		switch outputFormat {
		case config.OutputFormatJSON:
			output := map[string]interface{}{
				"vmid":   vmid,
				"name":   name,
				"status": statusStr,
				"type":   vmType,
				"node":   node,
			}
			b, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			output := map[string]interface{}{
				"vmid":   vmid,
				"name":   name,
				"status": statusStr,
				"type":   vmType,
				"node":   node,
			}
			b, err := yaml.Marshal(output)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
			fmt.Print(string(b))

		case config.OutputFormatText:
			// Print compact status (original format)
			fmt.Printf("%s %s: %s - %s\n", vmType, vmid, name, statusStr)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
