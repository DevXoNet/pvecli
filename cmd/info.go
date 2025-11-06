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
	"encoding/json"
	"fmt"
	"os"

	"pvecli/config"
	"pvecli/internal/proxmox"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var infoCmd = &cobra.Command{
	Use:   "info <vmid/ctid>",
	Short: "Show configuration info for VM or container by ID (node auto-detected)",
	Args:  cobra.ExactArgs(1),
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

		// Output based on configured format
		outputFormat := config.GetOutputFormat()
		
		switch outputFormat {
		case config.OutputFormatJSON:
			b, err := json.MarshalIndent(conf, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			fmt.Println(string(b))
			
		case config.OutputFormatYAML:
			b, err := yaml.Marshal(conf)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
			fmt.Print(string(b))
			
		case config.OutputFormatText:
			// Text format: key-value pairs
			for k, v := range conf {
				fmt.Printf("%s: %v\n", k, v)
			}
			
		default:
			// Fallback to table view
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"FIELD", "VALUE"})

			// Select fields based on instance type
			var fields []string
			if instanceType == "vm" {
				fields = []string{"name", "ostype", "cores", "sockets", "memory", "balloon", "scsihw", "boot", "bootdisk", "onboot", "agent", "net0", "scsi0", "ide2"}
			} else {
				fields = []string{"hostname", "ostype", "cores", "memory", "swap", "onboot", "net0", "rootfs", "arch", "features", "unprivileged"}
			}
			for _, k := range fields {
				if v, ok := conf[k]; ok {
					table.Append([]string{k, fmt.Sprintf("%v", v)})
				}
			}
			table.Render()

			// Dump full JSON for reference
			b, _ := json.MarshalIndent(conf, "", "  ")
			fmt.Printf("\nRaw config JSON for %s %s:\n%s", instanceType, vmid, string(b))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
