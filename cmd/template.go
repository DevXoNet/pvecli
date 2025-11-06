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

var templateCmd = &cobra.Command{
	Use:   "template <vmid>",
	Short: "Convert VM to template",
	Long:  "Convert a VM into a template (cannot be undone easily)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

		// Find node
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		if instanceType != "vm" {
			return fmt.Errorf("only VMs can be converted to templates")
		}

		// Convert to template
		err = client.CreateTemplate(node, vmid)
		if err != nil {
			return fmt.Errorf("failed to create template: %w", err)
		}

		// Output based on configured format
		outputFormat := config.GetOutputFormat()

		switch outputFormat {
		case config.OutputFormatJSON:
			output := map[string]interface{}{
				"vmid":   vmid,
				"node":   node,
				"action": "template",
				"status": "success",
			}
			b, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			output := map[string]interface{}{
				"vmid":   vmid,
				"node":   node,
				"action": "template",
				"status": "success",
			}
			b, _ := yaml.Marshal(output)
			fmt.Print(string(b))

		case config.OutputFormatText:
			fmt.Printf("VM %s converted to template successfully\n", vmid)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(templateCmd)
}
