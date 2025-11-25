// Copyright 2025 DevXo part of vByte Ltd //
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"
	"pvecli/internal/ssh"
)

var (
	consoleSSHUser    string
	consoleSSHKeyPath string
	consoleSSHPort    int
)

var consoleCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "console <vmid>",
	Short:         "Open an SSH console to a VM or container's host node",
	Long: `Open an interactive SSH console to the Proxmox node hosting the specified VM or container.

This connects you to the host node via SSH, not directly to the VM/container.
To access the VM/container itself, use 'qm terminal <vmid>' or 'pct console <vmid>' after connecting.

Features:
- Automatic node detection
- Uses SSH keys for authentication
- Configurable SSH user, key path, and port

Example:
pvecli console 105

Press Ctrl+D or type 'exit' to close the SSH session.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			output.PrintError(fmt.Errorf("config error: %w", err))
			return nil
		}
		client := pve.NewClient(cfg)

		// Get cluster config for SSH settings
		cluster, err := config.GetCurrentCluster()
		if err != nil {
			output.PrintError(err)
			return nil
		}

		// Get SSH settings from config or flags
		sshUser := consoleSSHUser
		if sshUser == "" && cluster.SSHUser != "" {
			sshUser = cluster.SSHUser
		}
		if sshUser == "" {
			sshUser = "root"
		}

		sshKeyPath := consoleSSHKeyPath
		if sshKeyPath == "" && cluster.SSHKeyPath != "" {
			sshKeyPath = cluster.SSHKeyPath
		}
		if sshKeyPath == "" {
			sshKeyPath = ssh.GetDefaultSSHKeyPath()
		}

		sshPort := consoleSSHPort
		if sshPort == 0 && cluster.SSHPort != 0 {
			sshPort = cluster.SSHPort
		}
		if sshPort == 0 {
			sshPort = 22
		}

		// Find node and instance type
		fmt.Printf("Locating instance %s...\n", vmid)
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			output.PrintError(err)
			return nil
		}

		// Try to get VM/container IP
		var vmIP string
		var ipErr error
		
		if instanceType == "qemu" {
			fmt.Printf("Getting IP from QEMU Guest Agent...\n")
			vmIP, ipErr = client.GetVMIPFromAgent(node, vmid)
		} else {
			fmt.Printf("Getting IP from container config...\n")
			vmIP, ipErr = client.GetContainerIPFromConfig(node, vmid)
		}

		if ipErr != nil || vmIP == "" {
			// Fallback to node SSH
			fmt.Printf("Could not get %s IP: %v\n", instanceType, ipErr)
			fmt.Printf("\nFalling back to node SSH connection...\n")
			fmt.Printf("After connecting, use:\n")
			if instanceType == "qemu" {
				fmt.Printf("  qm terminal %s    (for VM console)\n", vmid)
			} else {
				fmt.Printf("  pct console %s    (for container console)\n", vmid)
			}
			fmt.Println()

			// Get node IP
			nodes, err := client.GetNodes()
			if err != nil {
				output.PrintError(fmt.Errorf("failed to get nodes: %w", err))
				return nil
			}

			var nodeIP string
			for _, n := range nodes {
				if n.Node == node {
					nodeIP = n.IP
					break
				}
			}

			if nodeIP == "" {
				output.PrintError(fmt.Errorf("could not find IP for node %s", node))
				return nil
			}

			vmIP = nodeIP
			fmt.Printf("Connecting to node %s (%s) as %s@%s:%d...\n", node, nodeIP, sshUser, nodeIP, sshPort)
		} else {
			fmt.Printf("Found %s %s with IP %s\n", instanceType, vmid, vmIP)
			fmt.Printf("Connecting to %s@%s:%d...\n", sshUser, vmIP, sshPort)
		}

		// Use native SSH client for better terminal handling
		sshCmd := exec.Command("ssh",
			"-i", sshKeyPath,
			"-p", fmt.Sprintf("%d", sshPort),
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			fmt.Sprintf("%s@%s", sshUser, vmIP),
		)
		sshCmd.Stdin = os.Stdin
		sshCmd.Stdout = os.Stdout
		sshCmd.Stderr = os.Stderr

		if err := sshCmd.Run(); err != nil {
			output.PrintError(err)
			return nil
		}
		return nil
	},
}

func init() {
	consoleCmd.Flags().StringVar(&consoleSSHUser, "ssh-user", "", "SSH user (default: root or from config)")
	consoleCmd.Flags().StringVar(&consoleSSHKeyPath, "ssh-key", "", "SSH private key path (default: ~/.ssh/id_rsa or from config)")
	consoleCmd.Flags().IntVar(&consoleSSHPort, "ssh-port", 0, "SSH port (default: 22 or from config)")
	rootCmd.AddCommand(consoleCmd)
}
