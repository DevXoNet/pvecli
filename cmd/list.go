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
	"sort"
	"strconv"

	"pvecli/config"
	"pvecli/internal/pve"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	nodeFlag     string
	runningOnly  bool
	instanceType string
	tableFormat  bool
)

type vmEntry struct {
	Node       string
	VMID       int
	Name       string
	Status     string
	MaxMem     int64
	MaxDisk    int64
	CPUs       int
	Type       string
	Template   bool
	CloudInit  bool
	GuestAgent bool
}

var listCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "list",
	Short:         "List all VMs across the cluster or only from a specific node",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}

		client := pve.NewClient(cfg)

		nodes, err := client.GetNodes()
		if err != nil {
			return err
		}

		var entries []vmEntry

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"NODE", "VMID", "TYPE", "NAME", "STATUS", "CPU", "MEM(MB)", "DISK(GB)", "CLOUDINIT", "AGENT"})

		for _, n := range nodes {
			if nodeFlag != "" && n.Node != nodeFlag {
				continue
			}

			instances, err := client.GetAllInstances(n.Node)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warn: fetch %s failed: %v\n", n.Node, err)
				continue
			}
			for _, inst := range instances {
				if (!runningOnly || inst.Status == "running") &&
					(instanceType == "" || inst.Type == instanceType) {
					entries = append(entries, vmEntry{
						Node:       n.Node,
						VMID:       inst.VMID,
						Name:       inst.Name,
						Status:     inst.Status,
						MaxMem:     inst.MaxMem,
						MaxDisk:    inst.MaxDisk,
						CPUs:       inst.CPUs,
						Type:       inst.Type,
						Template:   inst.Template,
						CloudInit:  inst.CloudInit,
						GuestAgent: inst.GuestAgent,
					})
				}
			}
		}
		// Split entries into regular and template VMs
		var regularVMs, templateVMs []vmEntry
		for _, e := range entries {
			if e.Template {
				templateVMs = append(templateVMs, e)
			} else {
				regularVMs = append(regularVMs, e)
			}
		}

		// Sort both slices
		sortEntries := func(entries []vmEntry) {
			sort.Slice(entries, func(i, j int) bool {
				if entries[i].Node != entries[j].Node {
					return entries[i].Node < entries[j].Node
				}
				return entries[i].VMID < entries[j].VMID
			})
		}
		sortEntries(regularVMs)
		sortEntries(templateVMs)

		// Determine output format: --table flag overrides config
		outputFormat := config.GetOutputFormat()
		if tableFormat {
			outputFormat = "table" // Force table format
		}

		switch outputFormat {
		case config.OutputFormatJSON:
			// Output as JSON
			type jsonOutput struct {
				VMs       []vmEntry `json:"vms"`
				Templates []vmEntry `json:"templates"`
			}
			output := jsonOutput{
				VMs:       regularVMs,
				Templates: templateVMs,
			}
			b, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			// Output as YAML
			type yamlOutput struct {
				VMs       []vmEntry `yaml:"vms"`
				Templates []vmEntry `yaml:"templates"`
			}
			output := yamlOutput{
				VMs:       regularVMs,
				Templates: templateVMs,
			}
			b, err := yaml.Marshal(output)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
			fmt.Print(string(b))

		case config.OutputFormatText:
			// Output as plain text
			if len(regularVMs) > 0 {
				fmt.Println("VMs:")
				for _, e := range regularVMs {
					fmt.Printf("Node: %s, VMID: %d, Type: %s, Name: %s, Status: %s, CPU: %d, Mem: %dMB, Disk: %dGB, CloudInit: %s, Agent: %s\n",
						e.Node, e.VMID, e.Type, e.Name, e.Status, e.CPUs,
						e.MaxMem/1024/1024, e.MaxDisk/1024/1024/1024,
						map[bool]string{true: "YES", false: "NO"}[e.CloudInit],
						map[bool]string{true: "YES", false: "NO"}[e.GuestAgent])
				}
			}
			if len(templateVMs) > 0 {
				fmt.Println("\nTemplates:")
				for _, e := range templateVMs {
					fmt.Printf("Node: %s, VMID: %d, Type: %s, Name: %s, Status: %s, CPU: %d, Mem: %dMB, Disk: %dGB, CloudInit: %s, Agent: %s\n",
						e.Node, e.VMID, e.Type, e.Name, e.Status, e.CPUs,
						e.MaxMem/1024/1024, e.MaxDisk/1024/1024/1024,
						map[bool]string{true: "YES", false: "NO"}[e.CloudInit],
						map[bool]string{true: "YES", false: "NO"}[e.GuestAgent])
				}
			}

		default:
			// Table format (default or when --table is used)
			// Render regular VMs
			for _, e := range regularVMs {
				table.Append([]string{
					e.Node,
					strconv.Itoa(e.VMID),
					e.Type,
					e.Name,
					e.Status,
					fmt.Sprintf("%d", e.CPUs),
					fmt.Sprintf("%d", e.MaxMem/1024/1024),
					fmt.Sprintf("%d", e.MaxDisk/1024/1024/1024),
					map[bool]string{true: "YES", false: "NO"}[e.CloudInit],
					map[bool]string{true: "YES", false: "NO"}[e.GuestAgent],
				})
			}

			table.Render()

			// If we have templates, show them in a separate section
			if len(templateVMs) > 0 {
				fmt.Println("\nTemplates:")
				templateTable := tablewriter.NewWriter(os.Stdout)
				templateTable.SetHeader([]string{"NODE", "VMID", "TYPE", "NAME", "STATUS", "CPU", "MEM(MB)", "DISK(GB)", "CLOUDINIT", "AGENT"})

				for _, e := range templateVMs {
					templateTable.Append([]string{
						e.Node,
						strconv.Itoa(e.VMID),
						e.Type,
						e.Name,
						e.Status,
						fmt.Sprintf("%d", e.CPUs),
						fmt.Sprintf("%d", e.MaxMem/1024/1024),
						fmt.Sprintf("%d", e.MaxDisk/1024/1024/1024),
						map[bool]string{true: "YES", false: "NO"}[e.CloudInit],
						map[bool]string{true: "YES", false: "NO"}[e.GuestAgent],
					})
				}
				templateTable.Render()
			}
		}

		return nil
	},
}

func init() {
	listCmd.Flags().StringVarP(&nodeFlag, "node", "n", "", "Filter instances by node name")
	listCmd.Flags().StringVarP(&instanceType, "type", "t", "", "Filter by type: vm or ct")
	listCmd.Flags().BoolVarP(&runningOnly, "running", "r", false, "Show only running instances")
	listCmd.Flags().BoolVar(&tableFormat, "table", false, "Force table output format (overrides config)")
	rootCmd.AddCommand(listCmd)
}
