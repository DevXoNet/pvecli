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

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"

	"github.com/spf13/cobra"
)

var (
	diskName string
	diskSize string
)

var diskCmd = &cobra.Command{
	Use:   "disk",
	Short: "Disk management commands",
	Long:  "Manage VM disks (resize, move, etc.)",
}

var diskResizeCmd = &cobra.Command{
	Use:           "resize <vmid>",
	Short:         "Resize VM disk",
	Long:          "Resize a VM disk (can only increase size)",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]

		if diskName == "" || diskSize == "" {
			return fmt.Errorf("--disk and --size flags are required")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := pve.NewClient(cfg)

		// Find node
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		if instanceType != "qemu" {
			return fmt.Errorf("disk resize only supported for VMs, not containers")
		}

		// Resize disk
		err = client.ResizeDisk(node, vmid, diskName, diskSize)
		if err != nil {
			return fmt.Errorf("failed to resize disk: %w", err)
		}

		return output.PrintSuccess(fmt.Sprintf("Disk %s on VM %s resized to %s", diskName, vmid, diskSize), map[string]interface{}{
			"vmid":   vmid,
			"node":   node,
			"disk":   diskName,
			"size":   diskSize,
			"action": "resize",
		})
	},
}

func init() {
	diskResizeCmd.Flags().StringVar(&diskName, "disk", "", "Disk name (e.g., scsi0, virtio0)")
	diskResizeCmd.Flags().StringVar(&diskSize, "size", "", "New size (e.g., +10G, 50G)")
	diskResizeCmd.MarkFlagRequired("disk")
	diskResizeCmd.MarkFlagRequired("size")

	diskCmd.AddCommand(diskResizeCmd)
	rootCmd.AddCommand(diskCmd)
}
