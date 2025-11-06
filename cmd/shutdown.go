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

var (
	shutdownForce bool
)

var shutdownCmd = &cobra.Command{
	Use:   "shutdown <vmid>",
	Short: "Shutdown VM or container gracefully",
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

		// Shutdown
		err = client.ShutdownInstance(node, vmid, instanceType, shutdownForce)
		if err != nil {
			return fmt.Errorf("failed to shutdown: %w", err)
		}

		// Output based on configured format
		outputFormat := config.GetOutputFormat()

		switch outputFormat {
		case config.OutputFormatJSON:
			output := map[string]interface{}{
				"vmid":   vmid,
				"node":   node,
				"type":   instanceType,
				"action": "shutdown",
				"force":  shutdownForce,
				"status": "success",
			}
			b, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			output := map[string]interface{}{
				"vmid":   vmid,
				"node":   node,
				"type":   instanceType,
				"action": "shutdown",
				"force":  shutdownForce,
				"status": "success",
			}
			b, _ := yaml.Marshal(output)
			fmt.Print(string(b))

		case config.OutputFormatText:
			if shutdownForce {
				fmt.Printf("Instance %s forcefully shutdown\n", vmid)
			} else {
				fmt.Printf("Instance %s shutdown gracefully\n", vmid)
			}
		}

		return nil
	},
}

func init() {
	shutdownCmd.Flags().BoolVar(&shutdownForce, "force", false, "Force shutdown (hard stop)")
	rootCmd.AddCommand(shutdownCmd)
}
