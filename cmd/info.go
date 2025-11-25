// Copyright 2025 DevXo part of vByte Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/proxmox"

	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:           "info <vmid/ctid>",
	Short:         "Show configuration info for VM or container by ID (node auto-detected)",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		client := proxmox.NewClient(cfg)

		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		conf, err := client.GetInstanceConfig(node, vmid, instanceType)
		if err != nil {
			return err
		}

		return output.Print(conf)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
