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

package vm

import (
	"fmt"

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"

	"github.com/spf13/cobra"
)

var suspendCmd = &cobra.Command{
	Use:           "suspend <vmid>",
	Short:         "Suspend VM",
	Long:          "Suspend (pause) a running VM",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]

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
			return fmt.Errorf("only VMs can be suspended")
		}

		// Suspend
		err = client.SuspendVM(node, vmid)
		if err != nil {
			return fmt.Errorf("failed to suspend: %w", err)
		}

		return output.PrintSuccess(fmt.Sprintf("VM %s suspended successfully", vmid), map[string]interface{}{
			"vmid":   vmid,
			"node":   node,
			"action": "suspend",
		})
	},
}

var resumeCmd = &cobra.Command{
	Use:           "resume <vmid>",
	Short:         "Resume VM",
	Long:          "Resume a suspended VM",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]

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
			return fmt.Errorf("only VMs can be resumed")
		}

		// Resume
		err = client.ResumeVM(node, vmid)
		if err != nil {
			return fmt.Errorf("failed to resume: %w", err)
		}

		return output.PrintSuccess(fmt.Sprintf("VM %s resumed successfully", vmid), map[string]interface{}{
			"vmid":   vmid,
			"node":   node,
			"action": "resume",
		})
	},
}

