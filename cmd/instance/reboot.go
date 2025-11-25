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

package instance

import (
	"fmt"

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"

	"github.com/spf13/cobra"
)

var rebootCmd = &cobra.Command{
	Use:           "reboot <vmid>",
	Short:         "Reboot VM or container",
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
		node, instanceType, err := FindInstance(client, vmid)
		if err != nil {
			return err
		}

		// Reboot
		err = client.RebootInstance(node, vmid, instanceType)
		if err != nil {
			return fmt.Errorf("failed to reboot: %w", err)
		}

		return output.PrintSuccess(fmt.Sprintf("Instance %s rebooted successfully", vmid), map[string]interface{}{
			"vmid":   vmid,
			"node":   node,
			"type":   instanceType,
			"action": "reboot",
		})
	},
}

