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
	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:           "status <vmid/ctid>",
	Short:         "Show current status of VM or container",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := pve.NewClient(cfg)
		vmid := args[0]

		// Find which node has this VM/CT and its type
		node, instanceType, err := FindInstance(client, vmid)
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

		// Prepare output data
		data := map[string]interface{}{
			"vmid":   vmid,
			"name":   name,
			"status": statusStr,
			"type":   vmType,
			"node":   node,
		}

		return output.Print(data)
	},
}

