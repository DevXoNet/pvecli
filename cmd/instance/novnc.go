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
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"
	"pvecli/internal/ssh"
)

var (
	novncLocalPort   int
	novncRemotePort  int
	novncOpenBrowser bool
	novncTerminal    bool
	novncVNCViewer   bool
)

var novncCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "novnc <vmid>",
	Short:         "Open console for VM or LXC container",
	Long: `Access VM/Container console - multiple modes available.

Modes:
1. Browser (default): Opens noVNC in browser via SSH tunnel
2. Terminal (--terminal): Opens console directly (qm terminal for VM, pct enter for LXC)
3. VNC Viewer (--vncviewer): Opens VNC client (if installed)

This command:
1. Automatically detects which node hosts the instance
2. Creates SSH tunnel (browser/vncviewer) or direct SSH (terminal)
3. Opens console in selected mode

Exit Instructions:
- Terminal mode (VM): Press Ctrl+O to exit, then 'exit' for SSH
- Terminal mode (LXC): Type 'exit' or press Ctrl+D, then 'exit' for SSH
- Browser/VNC mode: Press Ctrl+C to close tunnel

Examples:
  pvecli novnc 240                    # Browser mode (default)
  pvecli novnc 240 --terminal         # Terminal mode (qm terminal)
  pvecli novnc 240 -t                 # Short form
  pvecli novnc 240 --vncviewer        # VNC client mode
  pvecli novnc 240 --browser=false    # Tunnel only, no browser
  pvecli novnc 240 --local-port 12345 # Custom local port`,
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

		// Get SSH settings from config (no flags - use 'console' command for custom SSH)
		sshUser := cluster.SSHUser
		if sshUser == "" {
			sshUser = "root"
		}

		sshKeyPath := cluster.SSHKeyPath
		if sshKeyPath == "" {
			sshKeyPath = ssh.GetDefaultSSHKeyPath()
		}

		sshPort := cluster.SSHPort
		if sshPort == 0 {
			sshPort = 22
		}

		// Find node and instance type
		fmt.Printf("Locating instance %s...\n", vmid)
		node, instanceType, err := FindInstance(client, vmid)
		if err != nil {
			output.PrintError(err)
			return nil
		}

		// Terminal mode - direct terminal via SSH (qm terminal for VMs, pct enter for LXC)
		terminalCommand := "qm terminal"
		if instanceType == "lxc" {
			terminalCommand = "pct enter"
		}
		if novncTerminal {
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

			instanceName := "VM"
			if instanceType == "lxc" {
				instanceName = "Container"
			}
			
			fmt.Printf("Found %s %s on node %s (%s)\n", instanceName, vmid, node, nodeIP)
			fmt.Printf("Opening terminal console via SSH...\n")
			fmt.Printf("\n")
			
			// Fixed width box - 60 characters
			fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
			fmt.Printf("║  HOW TO EXIT:                                            ║\n")
			if instanceType == "lxc" {
				fmt.Printf("║  Type 'exit' or press Ctrl+D                             ║\n")
			} else {
				fmt.Printf("║  Press Ctrl+O to exit serial terminal                    ║\n")
			}
			fmt.Printf("║  Then type 'exit' to close SSH                           ║\n")
			fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
			fmt.Printf("\n")

			// Use qm terminal (VM) or pct console (LXC) command via SSH
			sshTerminalCmd := exec.Command("ssh",
				"-i", sshKeyPath,
				"-p", strconv.Itoa(sshPort),
				"-o", "StrictHostKeyChecking=no",
				"-o", "UserKnownHostsFile=/dev/null",
				"-t", // Force pseudo-terminal allocation
				fmt.Sprintf("%s@%s", sshUser, nodeIP),
				fmt.Sprintf("%s %s", terminalCommand, vmid),
			)
			sshTerminalCmd.Stdin = os.Stdin
			sshTerminalCmd.Stdout = os.Stdout
			sshTerminalCmd.Stderr = os.Stderr

			if err := sshTerminalCmd.Run(); err != nil {
				if instanceType == "lxc" {
					fmt.Printf("\n⚠️  LXC console access failed (common issue with unprivileged containers)\n")
					fmt.Printf("💡 Try using browser mode instead:\n")
					fmt.Printf("   ./pvecli novnc %s\n\n", vmid)
					fmt.Printf("Or use SSH to the node and run:\n")
					fmt.Printf("   pct enter %s\n\n", vmid)
				} else {
					output.PrintError(fmt.Errorf("failed to open terminal: %w", err))
				}
				return nil
			}

			fmt.Printf("\nTerminal session closed.\n")
			return nil
		}

		// Get VM config to find VNC display
		vmConfig, err := client.GetInstanceConfig(node, vmid, instanceType)
		if err != nil {
			output.PrintError(fmt.Errorf("failed to get VM config: %w", err))
			return nil
		}

		// Calculate VNC port (default is 5900 + display number)
		// Most VMs use display 0, so VNC port is 5900
		vncDisplay := 0
		if displayStr, ok := vmConfig["display"].(string); ok {
			// Parse display string (e.g., "vnc=:0")
			fmt.Sscanf(displayStr, "vnc=:%d", &vncDisplay)
		}

		remotePort := novncRemotePort
		if remotePort == 0 {
			remotePort = 5900 + vncDisplay
		}

		localPort := novncLocalPort
		if localPort == 0 {
			localPort = 5900
		}

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

		fmt.Printf("Found VM %s on node %s (%s)\n", vmid, node, nodeIP)
		fmt.Printf("VNC display: %d, VNC port: %d\n", vncDisplay, remotePort)
		fmt.Printf("Creating SSH tunnel: localhost:%d -> %s:%d\n", localPort, nodeIP, remotePort)

		// Create SSH tunnel command
		// ssh -L local_port:localhost:remote_port user@node -N -f
		tunnelSpec := fmt.Sprintf("127.0.0.1:%d:127.0.0.1:%d", localPort, remotePort)
		sshCmd := exec.Command("ssh",
			"-L", tunnelSpec,
			"-i", sshKeyPath,
			"-p", strconv.Itoa(sshPort),
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "ExitOnForwardFailure=yes",
			"-N", // Don't execute remote command
			fmt.Sprintf("%s@%s", sshUser, nodeIP),
		)

		// Start SSH tunnel
		if err := sshCmd.Start(); err != nil {
			output.PrintError(fmt.Errorf("failed to start SSH tunnel: %w", err))
			return nil
		}

		// Wait a bit for tunnel to establish
		time.Sleep(1 * time.Second)

		fmt.Printf("SSH tunnel established!\n\n")

		// VNC Viewer mode - open VNC client
		if novncVNCViewer {
			fmt.Printf("Opening VNC viewer...\n")

			// Try common VNC viewers
			vncViewers := []string{"vncviewer", "vinagre", "krdc", "remmina"}
			var viewerCmd *exec.Cmd
			viewerFound := false

			for _, viewer := range vncViewers {
				if _, err := exec.LookPath(viewer); err == nil {
					fmt.Printf("   Using %s\n", viewer)
					viewerCmd = exec.Command(viewer, fmt.Sprintf("localhost:%d", localPort))
					viewerFound = true
					break
				}
			}

			if !viewerFound {
				fmt.Printf("No VNC viewer found. Please install one of: %v\n", vncViewers)
				fmt.Printf("   Or connect manually to: localhost:%d\n\n", localPort)
			} else {
				if err := viewerCmd.Start(); err != nil {
					fmt.Printf("Failed to start VNC viewer: %v\n", err)
					fmt.Printf("   Connect manually to: localhost:%d\n\n", localPort)
				} else {
					fmt.Printf("VNC viewer started\n\n")
				}
			}

			fmt.Printf("Press Ctrl+C to close the tunnel...\n")

			// Wait for interrupt
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
			<-sigChan

			fmt.Printf("\nClosing SSH tunnel...\n")
			if err := sshCmd.Process.Kill(); err != nil {
				fmt.Printf("Warning: failed to kill SSH process: %v\n", err)
			}
			fmt.Printf("Tunnel closed. Goodbye!\n")
			return nil
		}

		// Get Proxmox VE web interface URL for noVNC
		proxmoxURL := cluster.APIURL
		if proxmoxURL == "" {
			proxmoxURL = fmt.Sprintf("https://%s:8006", nodeIP)
		}

		// Construct noVNC URL
		// Format: https://proxmox:8006/?console=kvm&novnc=1&vmid=100&node=node01
		novncURL := fmt.Sprintf("%s/?console=kvm&novnc=1&vmid=%s&node=%s", proxmoxURL, vmid, node)

		fmt.Printf("noVNC Console URL:\n")
		fmt.Printf("   %s\n\n", novncURL)

		// Open browser if requested
		if novncOpenBrowser {
			fmt.Printf("Opening browser...\n")
			var browserCmd *exec.Cmd
			switch {
			case exec.Command("xdg-open", novncURL).Run() == nil:
				// Linux
			case exec.Command("open", novncURL).Run() == nil:
				// macOS
			case exec.Command("cmd", "/c", "start", novncURL).Run() == nil:
				// Windows
			default:
				fmt.Printf("Could not open browser automatically\n")
				fmt.Printf("   Please open the URL above manually\n")
			}
			_ = browserCmd
		}

		fmt.Printf("Alternative: Use VNC client to connect to localhost:%d\n", localPort)
		fmt.Printf("   Example: vncviewer localhost:%d\n\n", localPort)
		fmt.Printf("Press Ctrl+C to close the tunnel...\n")

		// Wait for interrupt signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		fmt.Printf("\nClosing SSH tunnel...\n")

		// Kill SSH process
		if err := sshCmd.Process.Kill(); err != nil {
			fmt.Printf("Warning: failed to kill SSH process: %v\n", err)
		}

		fmt.Printf("Tunnel closed. Goodbye!\n")
		return nil
	},
}

func init() {
	novncCmd.Flags().BoolVarP(&novncTerminal, "terminal", "t", false, "Open qm terminal directly (no VNC)")
	novncCmd.Flags().BoolVarP(&novncVNCViewer, "vncviewer", "v", false, "Open VNC viewer client (requires vncviewer installed)")
	novncCmd.Flags().IntVarP(&novncLocalPort, "local-port", "l", 5900, "Local port for SSH tunnel")
	novncCmd.Flags().IntVarP(&novncRemotePort, "remote-port", "r", 0, "Remote VNC port (default: auto-detect)")
	novncCmd.Flags().BoolVarP(&novncOpenBrowser, "browser", "b", true, "Open browser automatically (browser mode only)")
}
