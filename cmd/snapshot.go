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
	"time"

	"pvecli/config"
	"pvecli/internal/proxmox"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	snapshotName        string
	snapshotDescription string
	snapshotNoRAM       bool
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot <vmid>",
	Short: "Create a snapshot of a VM or container",
	Long: `Create a snapshot of a VM or container.

A snapshot captures the current state of the VM/container and can be used to restore it later.

Example:
  pvecli snapshot 230
  pvecli snapshot 202 --name my-snapshot --description "Before upgrade"
  pvecli snapshot 114 --name my-snapshot --description "Before upgrade" --no-ram`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

		// Find which node hosts this VM/container
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		// Generate snapshot name if not provided
		snapName := snapshotName
		if snapName == "" {
			snapName = fmt.Sprintf("snap_%d", time.Now().Unix())
		}

		// Create snapshot (includeRAM is true by default, unless --no-ram is specified)
		includeRAM := !snapshotNoRAM
		taskID, err := client.CreateSnapshot(node, vmid, instanceType, snapName, snapshotDescription, includeRAM)

		// Get output format before handling error
		outputFormat := config.GetOutputFormat()

		if err != nil {
			// Format error according to output format
			switch outputFormat {
			case config.OutputFormatJSON:
				errorOutput := map[string]interface{}{
					"error":  err.Error(),
					"vmid":   vmid,
					"status": "failed",
				}
				b, _ := json.MarshalIndent(errorOutput, "", "  ")
				fmt.Println(string(b))
			case config.OutputFormatYAML:
				errorOutput := map[string]interface{}{
					"error":  err.Error(),
					"vmid":   vmid,
					"status": "failed",
				}
				b, _ := yaml.Marshal(errorOutput)
				fmt.Print(string(b))
			case config.OutputFormatText:
				fmt.Printf("Error: %s\n", err.Error())
			}
			os.Exit(1)
		}

		// Output based on configured format
		switch outputFormat {
		case config.OutputFormatJSON:
			output := map[string]interface{}{
				"vmid":        vmid,
				"node":        node,
				"type":        instanceType,
				"snapshot":    snapName,
				"description": snapshotDescription,
				"task_id":     taskID,
				"status":      "started",
			}
			b, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			output := map[string]interface{}{
				"vmid":        vmid,
				"node":        node,
				"type":        instanceType,
				"snapshot":    snapName,
				"description": snapshotDescription,
				"task_id":     taskID,
				"status":      "started",
			}
			b, err := yaml.Marshal(output)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
			fmt.Print(string(b))

		case config.OutputFormatText:
			fmt.Printf("Snapshot '%s' created successfully\n", snapName)
			if taskID != "" {
				fmt.Printf("Task ID: %s\n", taskID)
				fmt.Println("Use 'pvecli task <task_id>' to monitor progress")
			}
		}

		return nil
	},
}

var snapshotDelCmd = &cobra.Command{
	Use:   "del <vmid>",
	Short: "Delete a snapshot of a VM or container",
	Long: `Delete a snapshot of a VM or container by name.

Example:
  pvecli snapshot del 230 --name my-snapshot
  pvecli snapshot del 114 --name my-snapshot`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]

		// Require snapshot name
		if snapshotName == "" {
			outputFormat := config.GetOutputFormat()
			switch outputFormat {
			case config.OutputFormatJSON:
				errorOutput := map[string]interface{}{
					"error":  "--name flag is required",
					"vmid":   vmid,
					"status": "failed",
				}
				b, _ := json.MarshalIndent(errorOutput, "", "  ")
				fmt.Println(string(b))
			case config.OutputFormatYAML:
				errorOutput := map[string]interface{}{
					"error":  "--name flag is required",
					"vmid":   vmid,
					"status": "failed",
				}
				b, _ := yaml.Marshal(errorOutput)
				fmt.Print(string(b))
			case config.OutputFormatText:
				fmt.Println("Error: --name flag is required")
			}
			os.Exit(1)
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

		// Find which node hosts this VM/container
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		// Delete snapshot
		taskID, err := client.DeleteSnapshot(node, vmid, instanceType, snapshotName)

		// Get output format before handling error
		outputFormat := config.GetOutputFormat()

		if err != nil {
			// Format error according to output format
			switch outputFormat {
			case config.OutputFormatJSON:
				errorOutput := map[string]interface{}{
					"error":  err.Error(),
					"vmid":   vmid,
					"status": "failed",
				}
				b, _ := json.MarshalIndent(errorOutput, "", "  ")
				fmt.Println(string(b))
			case config.OutputFormatYAML:
				errorOutput := map[string]interface{}{
					"error":  err.Error(),
					"vmid":   vmid,
					"status": "failed",
				}
				b, _ := yaml.Marshal(errorOutput)
				fmt.Print(string(b))
			case config.OutputFormatText:
				fmt.Printf("Error: %s\n", err.Error())
			}
			os.Exit(1)
		}

		// Output based on configured format
		switch outputFormat {
		case config.OutputFormatJSON:
			output := map[string]interface{}{
				"vmid":     vmid,
				"node":     node,
				"type":     instanceType,
				"snapshot": snapshotName,
				"task_id":  taskID,
				"status":   "deleted",
			}
			b, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			output := map[string]interface{}{
				"vmid":     vmid,
				"node":     node,
				"type":     instanceType,
				"snapshot": snapshotName,
				"task_id":  taskID,
				"status":   "deleted",
			}
			b, err := yaml.Marshal(output)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
			fmt.Print(string(b))

		case config.OutputFormatText:
			fmt.Printf("Snapshot '%s' deleted successfully\n", snapshotName)
			if taskID != "" {
				fmt.Printf("Task ID: %s\n", taskID)
			}
		}

		return nil
	},
}

func init() {
	snapshotCmd.Flags().StringVar(&snapshotName, "name", "", "Snapshot name (auto-generated if not specified)")
	snapshotCmd.Flags().StringVar(&snapshotDescription, "description", "", "Snapshot description")
	snapshotCmd.Flags().BoolVar(&snapshotNoRAM, "no-ram", false, "Exclude RAM/VM state from snapshot (RAM included by default for QEMU VMs)")

	snapshotDelCmd.Flags().StringVar(&snapshotName, "name", "", "Snapshot name to delete (required)")
	snapshotCmd.AddCommand(snapshotDelCmd)

	rootCmd.AddCommand(snapshotCmd)
}
