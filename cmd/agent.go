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

var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Interact with QEMU Guest Agent",
	Long: `Interact with QEMU Guest Agent running inside VMs.

The QEMU Guest Agent must be installed and running inside the VM.
Some commands (like exec) may be disabled by default for security reasons.`,
}

// agent ping
var agentPingCmd = &cobra.Command{
	Use:   "ping <vmid>",
	Short: "Ping the guest agent",
	Long:  "Check if the QEMU Guest Agent is running and responding",
	Args:  cobra.ExactArgs(1),
	SilenceUsage: true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := proxmox.NewClient(cfg)
		vmid := args[0]

		// Find node
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		// Only works with VMs, not containers
		if instanceType != "vm" {
			return fmt.Errorf("guest agent only works with VMs, not containers")
		}

		// Ping agent
		err = client.AgentPing(node, vmid)
		if err != nil {
			return fmt.Errorf("guest agent not responding: %w", err)
		}

		// Output success
		data := map[string]interface{}{
			"vmid":   vmid,
			"status": "agent is running",
		}
		
		return output.Print(data)
	},
}

// agent network
var agentNetworkCmd = &cobra.Command{
	Use:   "network <vmid>",
	Short: "Get network interfaces information",
	Long:  "Retrieve network interfaces, IP addresses, and MAC addresses from the guest",
	Args:  cobra.ExactArgs(1),
	SilenceUsage: true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := proxmox.NewClient(cfg)
		vmid := args[0]

		// Find node
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		if instanceType != "vm" {
			return fmt.Errorf("guest agent only works with VMs, not containers")
		}

		// Get network info
		networkInfo, err := client.AgentGetNetworkInterfaces(node, vmid)
		if err != nil {
			return err
		}

		return output.Print(networkInfo)
	},
}

// agent osinfo
var agentOSInfoCmd = &cobra.Command{
	Use:   "osinfo <vmid>",
	Short: "Get operating system information",
	Long:  "Retrieve OS name, version, kernel version, and architecture from the guest",
	Args:  cobra.ExactArgs(1),
	SilenceUsage: true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := proxmox.NewClient(cfg)
		vmid := args[0]

		// Find node
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		if instanceType != "vm" {
			return fmt.Errorf("guest agent only works with VMs, not containers")
		}

		// Get OS info
		osInfo, err := client.AgentGetOSInfo(node, vmid)
		if err != nil {
			return err
		}

		return output.Print(osInfo)
	},
}

// agent exec
var agentExecCmd = &cobra.Command{
	Use:   "exec <vmid> <command>",
	Short: "Execute a command in the guest",
	Long: `Execute a command inside the guest VM.

WARNING: This command is disabled by default for security reasons.
To enable it, you must modify the guest agent configuration inside the VM:

Linux (Debian/Ubuntu):
  Edit /etc/default/qemu-guest-agent
  Remove 'guest-exec' and 'guest-exec-status' from BLACKLIST_RPC
  Restart: systemctl restart qemu-guest-agent

Linux (RHEL/CentOS/AlmaLinux):
  Edit /etc/sysconfig/qemu-ga
  Add 'guest-exec' and 'guest-exec-status' to FILTER_RPC_ARGS (--allow-rpcs)
  Restart: systemctl restart qemu-guest-agent`,
	Args: cobra.MinimumNArgs(2),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := proxmox.NewClient(cfg)
		vmid := args[0]
		command := args[1:]

		// Find node
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		if instanceType != "vm" {
			return fmt.Errorf("guest agent only works with VMs, not containers")
		}

		// Execute command
		result, err := client.AgentExec(node, vmid, command)
		if err != nil {
			return err
		}

		return output.Print(result)
	},
}

// agent fsinfo
var agentFSInfoCmd = &cobra.Command{
	Use:   "fsinfo <vmid>",
	Short: "Get filesystem information",
	Long:  "Retrieve filesystem information including mount points, disk usage, and filesystem types",
	Args:  cobra.ExactArgs(1),
	SilenceUsage: true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := proxmox.NewClient(cfg)
		vmid := args[0]

		// Find node
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		if instanceType != "vm" {
			return fmt.Errorf("guest agent only works with VMs, not containers")
		}

		// Get filesystem info
		fsInfo, err := client.AgentGetFSInfo(node, vmid)
		if err != nil {
			return err
		}

		return output.Print(fsInfo)
	},
}

func init() {
	// Add subcommands
	agentCmd.AddCommand(agentPingCmd)
	agentCmd.AddCommand(agentNetworkCmd)
	agentCmd.AddCommand(agentOSInfoCmd)
	// agentCmd.AddCommand(agentExecCmd) // Disabled - timing issues with fast commands
	agentCmd.AddCommand(agentFSInfoCmd)

	// Add to root
	rootCmd.AddCommand(agentCmd)
}
