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

package cluster

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"pvecli/config"
	"pvecli/internal/pve"
	"pvecli/internal/ssh"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	updateAutoApprove bool
	updateSSHUser     string
	updateSSHKeyPath  string
	updateSSHPort     int
)

var clusterUpdateCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "update",
	Short:         "Update all cluster nodes via SSH",
	Long: `Update all cluster nodes by running apt update && apt upgrade via SSH.

The command will:
1. Get node IPs from Proxmox API
2. Connect via SSH using configured credentials
3. Check for available updates
4. Show what will be updated
5. Ask for confirmation (unless --yes is used)
6. Apply updates on all nodes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := pve.NewClient(cfg)

		// Get cluster config
		cluster, err := config.GetCurrentCluster()
		if err != nil {
			return err
		}

		// Get SSH settings from config or flags
		sshUser := updateSSHUser
		if sshUser == "" && cluster.SSHUser != "" {
			sshUser = cluster.SSHUser
		}
		if sshUser == "" {
			sshUser = "root" // Default to root
		}

		sshKeyPath := updateSSHKeyPath
		if sshKeyPath == "" && cluster.SSHKeyPath != "" {
			sshKeyPath = cluster.SSHKeyPath
		}
		if sshKeyPath == "" {
			sshKeyPath = ssh.GetDefaultSSHKeyPath()
		}

		sshPort := updateSSHPort
		if sshPort == 0 && cluster.SSHPort != 0 {
			sshPort = cluster.SSHPort
		}
		if sshPort == 0 {
			sshPort = 22 // Default SSH port
		}

		// Check if SSH key exists
		if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
			return fmt.Errorf("SSH key not found: %s\n\nPlease ensure:\n1. SSH key exists at the specified path\n2. SSH public key is added to nodes' authorized_keys\n3. Use: ssh-copy-id -i %s root@<node-ip>", sshKeyPath, sshKeyPath)
		}

		// Get all nodes
		nodes, err := client.GetNodes()
		if err != nil {
			return err
		}

		// Get output format
		outputFormat := config.GetOutputFormat()

		if outputFormat == config.OutputFormatText {
			fmt.Printf("Found %d nodes in cluster\n", len(nodes))
			fmt.Printf("SSH User: %s\n", sshUser)
			fmt.Printf("SSH Key: %s\n", sshKeyPath)
			fmt.Printf("SSH Port: %d\n\n", sshPort)
		}

		// Check for updates on all nodes (in parallel for speed)
		type nodeUpdate struct {
			name    string
			ip      string
			updates string
			err     error
		}

		if outputFormat == config.OutputFormatText {
			fmt.Println("Checking for updates on all nodes...")
		}

		// Use channels for parallel execution
		resultsChan := make(chan nodeUpdate, len(nodes))

		for _, node := range nodes {
			go func(n pve.Node) {
				result := nodeUpdate{
					name: n.Node,
					ip:   n.IP,
				}

				sshClient, err := ssh.NewClient(n.IP, sshUser, sshKeyPath, sshPort)
				if err != nil {
					result.err = fmt.Errorf("SSH connection failed: %w", err)
					resultsChan <- result
					return
				}

				// Run apt update
				_, err = sshClient.RunCommand("apt update -qq")
				if err != nil {
					sshClient.Close()
					result.err = fmt.Errorf("apt update failed: %w", err)
					resultsChan <- result
					return
				}

				// Check what would be upgraded
				output, err := sshClient.RunCommand("apt list --upgradable 2>/dev/null | grep -v 'Listing' || true")
				sshClient.Close()

				if err != nil {
					result.err = fmt.Errorf("failed to check updates: %w", err)
					resultsChan <- result
					return
				}

				result.updates = strings.TrimSpace(output)
				resultsChan <- result
			}(node)
		}

		// Collect results from channel
		nodeUpdates := make([]nodeUpdate, 0, len(nodes))
		for i := 0; i < len(nodes); i++ {
			result := <-resultsChan
			nodeUpdates = append(nodeUpdates, result)

			if outputFormat == config.OutputFormatText {
				if result.err != nil {
					fmt.Printf("  → %s (%s)... ✗ %v\n", result.name, result.ip, result.err)
				} else if result.updates == "" {
					fmt.Printf("  → %s (%s)... ✓ No updates available\n", result.name, result.ip)
				} else {
					lines := strings.Split(strings.TrimSpace(result.updates), "\n")
					fmt.Printf("  → %s (%s)... ✓ %d updates available\n", result.name, result.ip, len(lines))
				}
			}
		}

		// Show updates summary
		hasUpdates := false
		for _, nu := range nodeUpdates {
			if nu.err == nil && nu.updates != "" {
				hasUpdates = true
				break
			}
		}

		// Format output based on config
		switch outputFormat {
		case config.OutputFormatJSON:
			type nodeResult struct {
				Node           string   `json:"node"`
				IP             string   `json:"ip"`
				Status         string   `json:"status"`
				Error          string   `json:"error,omitempty"`
				UpdatesCount   int      `json:"updates_count,omitempty"`
				UpdatesPreview []string `json:"updates_preview,omitempty"`
			}

			results := make([]nodeResult, 0, len(nodeUpdates))
			for _, nu := range nodeUpdates {
				result := nodeResult{
					Node: nu.name,
					IP:   nu.ip,
				}

				if nu.err != nil {
					result.Status = "error"
					result.Error = nu.err.Error()
				} else if nu.updates == "" {
					result.Status = "up_to_date"
				} else {
					result.Status = "updates_available"
					lines := strings.Split(strings.TrimSpace(nu.updates), "\n")
					result.UpdatesCount = len(lines)
					if len(lines) > 10 {
						result.UpdatesPreview = lines[:10]
					} else {
						result.UpdatesPreview = lines
					}
				}
				results = append(results, result)
			}

			output := map[string]interface{}{
				"nodes":       results,
				"has_updates": hasUpdates,
			}
			b, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			type nodeResult struct {
				Node           string   `yaml:"node"`
				IP             string   `yaml:"ip"`
				Status         string   `yaml:"status"`
				Error          string   `yaml:"error,omitempty"`
				UpdatesCount   int      `yaml:"updates_count,omitempty"`
				UpdatesPreview []string `yaml:"updates_preview,omitempty"`
			}

			results := make([]nodeResult, 0, len(nodeUpdates))
			for _, nu := range nodeUpdates {
				result := nodeResult{
					Node: nu.name,
					IP:   nu.ip,
				}

				if nu.err != nil {
					result.Status = "error"
					result.Error = nu.err.Error()
				} else if nu.updates == "" {
					result.Status = "up_to_date"
				} else {
					result.Status = "updates_available"
					lines := strings.Split(strings.TrimSpace(nu.updates), "\n")
					result.UpdatesCount = len(lines)
					if len(lines) > 10 {
						result.UpdatesPreview = lines[:10]
					} else {
						result.UpdatesPreview = lines
					}
				}
				results = append(results, result)
			}

			output := map[string]interface{}{
				"nodes":       results,
				"has_updates": hasUpdates,
			}
			b, _ := yaml.Marshal(output)
			fmt.Print(string(b))

		case config.OutputFormatText:
			fmt.Println("\n═══════════════════════════════════════")
			fmt.Println("Updates Summary:")
			fmt.Println("═══════════════════════════════════════")

			for _, nu := range nodeUpdates {
				if nu.err != nil {
					fmt.Printf("\n%s (%s): ✗ Error\n", nu.name, nu.ip)
					fmt.Printf("  %v\n", nu.err)
				} else if nu.updates == "" {
					fmt.Printf("\n%s (%s): ✓ Up to date\n", nu.name, nu.ip)
				} else {
					fmt.Printf("\n%s (%s): Updates available\n", nu.name, nu.ip)
					lines := strings.Split(strings.TrimSpace(nu.updates), "\n")
					for i, line := range lines {
						if i < 10 { // Show first 10 packages
							fmt.Printf("  • %s\n", line)
						}
					}
					if len(lines) > 10 {
						fmt.Printf("  ... and %d more\n", len(lines)-10)
					}
				}
			}
		}

		if !hasUpdates {
			if outputFormat == config.OutputFormatText {
				fmt.Println("\n✓ All nodes are up to date!")
			}
			return nil
		}

		// Ask for confirmation (only in text mode)
		if !updateAutoApprove && outputFormat == config.OutputFormatText {
			fmt.Print("\nDo you want to apply these updates? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				fmt.Println("Update cancelled.")
				return nil
			}
		} else if !updateAutoApprove {
			// For JSON/YAML, require --yes flag
			return fmt.Errorf("--yes flag is required for non-interactive mode (JSON/YAML output)")
		}

		// Apply updates
		if outputFormat == config.OutputFormatText {
			fmt.Println("\n═══════════════════════════════════════")
			fmt.Println("Applying updates...")
			fmt.Println("═══════════════════════════════════════")
			fmt.Println()
		}

		successCount := 0
		failCount := 0
		type applyResult struct {
			node   string
			ip     string
			status string
			error  string
		}
		applyResults := make([]applyResult, 0)

		for _, nu := range nodeUpdates {
			if nu.err != nil || nu.updates == "" {
				continue // Skip nodes with errors or no updates
			}

			if outputFormat == config.OutputFormatText {
				fmt.Printf("%s (%s):\n", nu.name, nu.ip)
			}

			sshClient, err := ssh.NewClient(nu.ip, sshUser, sshKeyPath, sshPort)
			if err != nil {
				if outputFormat == config.OutputFormatText {
					fmt.Printf("  ✗ SSH connection failed: %v\n\n", err)
				}
				failCount++
				applyResults = append(applyResults, applyResult{
					node:   nu.name,
					ip:     nu.ip,
					status: "failed",
					error:  err.Error(),
				})
				continue
			}

			if outputFormat == config.OutputFormatText {
				fmt.Println("  → Running apt upgrade -y...")
			}
			output, err := sshClient.RunCommand("DEBIAN_FRONTEND=noninteractive apt upgrade -y")
			sshClient.Close()

			if err != nil {
				if outputFormat == config.OutputFormatText {
					fmt.Printf("  ✗ Update failed: %v\n", err)
					if output != "" {
						fmt.Printf("  Output: %s\n", output)
					}
					fmt.Println()
				}
				failCount++
				applyResults = append(applyResults, applyResult{
					node:   nu.name,
					ip:     nu.ip,
					status: "failed",
					error:  err.Error(),
				})
				continue
			}

			if outputFormat == config.OutputFormatText {
				fmt.Println("  ✓ Updates applied successfully")
				fmt.Println()
			}
			successCount++
			applyResults = append(applyResults, applyResult{
				node:   nu.name,
				ip:     nu.ip,
				status: "success",
			})
		}

		// Final summary
		switch outputFormat {
		case config.OutputFormatJSON:
			type resultNode struct {
				Node   string `json:"node"`
				IP     string `json:"ip"`
				Status string `json:"status"`
				Error  string `json:"error,omitempty"`
			}

			results := make([]resultNode, 0, len(applyResults))
			for _, r := range applyResults {
				results = append(results, resultNode{
					Node:   r.node,
					IP:     r.ip,
					Status: r.status,
					Error:  r.error,
				})
			}

			summary := map[string]interface{}{
				"results":       results,
				"success_count": successCount,
				"failed_count":  failCount,
			}
			b, _ := json.MarshalIndent(summary, "", "  ")
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			type resultNode struct {
				Node   string `yaml:"node"`
				IP     string `yaml:"ip"`
				Status string `yaml:"status"`
				Error  string `yaml:"error,omitempty"`
			}

			results := make([]resultNode, 0, len(applyResults))
			for _, r := range applyResults {
				results = append(results, resultNode{
					Node:   r.node,
					IP:     r.ip,
					Status: r.status,
					Error:  r.error,
				})
			}

			summary := map[string]interface{}{
				"results":       results,
				"success_count": successCount,
				"failed_count":  failCount,
			}
			b, _ := yaml.Marshal(summary)
			fmt.Print(string(b))

		case config.OutputFormatText:
			fmt.Println("═══════════════════════════════════════")
			fmt.Println("Final Summary:")
			fmt.Printf("  ✓ Success: %d nodes\n", successCount)
			if failCount > 0 {
				fmt.Printf("  ✗ Failed:  %d nodes\n", failCount)
			}
			fmt.Println("═══════════════════════════════════════")
		}

		return nil
	},
}

func init() {
	clusterUpdateCmd.Flags().BoolVarP(&updateAutoApprove, "yes", "y", false, "Auto-approve updates without confirmation")
	clusterUpdateCmd.Flags().StringVar(&updateSSHUser, "ssh-user", "", "SSH user (default: root or from config)")
	clusterUpdateCmd.Flags().StringVar(&updateSSHKeyPath, "ssh-key", "", "SSH private key path (default: ~/.ssh/id_rsa or from config)")
	clusterUpdateCmd.Flags().IntVar(&updateSSHPort, "ssh-port", 0, "SSH port (default: 22 or from config)")
	clusterCmd.AddCommand(clusterUpdateCmd)
}
