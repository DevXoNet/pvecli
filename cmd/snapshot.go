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
	"fmt"
	"time"

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"

	"github.com/spf13/cobra"
)

var (
	snapshotName        string
	snapshotDescription string
	snapshotNoRAM       bool
)

var snapshotCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "snapshot <vmid>",
	Short:         "Create a snapshot of a VM or container",
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
		client := pve.NewClient(cfg)

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
		if err != nil {
			return err
		}

		// Output using output package
		data := map[string]interface{}{
			"vmid":     vmid,
			"node":     node,
			"type":     instanceType,
			"snapshot": snapName,
			"task_id":  taskID,
		}
		if snapshotDescription != "" {
			data["description"] = snapshotDescription
		}
		return output.PrintSuccess(fmt.Sprintf("Snapshot '%s' created successfully", snapName), data)
	},
}

var snapshotDelCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "del <vmid>",
	Short:         "Delete a snapshot of a VM or container",
	Long: `Delete a snapshot of a VM or container by name.

Example:
  pvecli snapshot del 230 --name my-snapshot
  pvecli snapshot del 114 --name my-snapshot`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]

		// Require snapshot name
		if snapshotName == "" {
			return fmt.Errorf("--name flag is required")
		}

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := pve.NewClient(cfg)

		// Find which node hosts this VM/container
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		// Delete snapshot
		taskID, err := client.DeleteSnapshot(node, vmid, instanceType, snapshotName)
		if err != nil {
			return err
		}

		// Output using output package
		data := map[string]interface{}{
			"vmid":     vmid,
			"node":     node,
			"type":     instanceType,
			"snapshot": snapshotName,
			"task_id":  taskID,
		}
		return output.PrintSuccess(fmt.Sprintf("Snapshot '%s' deleted successfully", snapshotName), data)
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
