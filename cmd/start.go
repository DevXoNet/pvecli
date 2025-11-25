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
	"pvecli/internal/proxmox"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	SilenceUsage:  true, // Don't show usage on error
	SilenceErrors: true, // Don't show error: prefix
	Use:           "start <vmid/ctid>",
	Short:         "Start VM or container by ID (node auto-detected)",
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		client := proxmox.NewClient(cfg)
		node, vmType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}
		err = client.VMAction(node, vmid, "start")
		if err != nil {
			return err
		}
		
		// Output success message
		data := map[string]interface{}{
			"vmid": vmid,
			"node": node,
			"type": vmType,
		}
		return output.PrintSuccess(fmt.Sprintf("VM/CT %s started successfully", vmid), data)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
