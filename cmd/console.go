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
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"

	"pvecli/config"
	"pvecli/internal/console"
	"pvecli/internal/proxmox"
)

var (
	consoleNodeOverride string
	consoleSourceNode   string
	consoleTargetNode   string
)

var consoleCmd = &cobra.Command{
	Use:   "console <vmid>",
	Short: "Open an interactive console to a VM or container",
	Long: `Open an interactive TTY console to a VM or container using Proxmox VNC WebSocket.

This provides SSM-like access without requiring SSH or installing any agent inside the VM/container.

The console connection is established through the Proxmox API using termproxy and vncwebsocket.

Features:
- No SSH required
- No agent installation needed
- Automatic terminal resizing
- Works with both VMs and LXC containers

Example:
pvecli console 105

Press Ctrl+C or close the terminal to exit the console session.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]
		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

		// Find the node where the VM/container is located
		var node, instanceType string

		if consoleSourceNode != "" {
			// If source node is specified, search only on that node
			fmt.Printf("Locating instance %s on node %s...\n", vmid, consoleSourceNode)
			var err error
			instanceType, err = client.FindVMIDOnNode(consoleSourceNode, vmid)
			if err != nil {
				return err
			}
			node = consoleSourceNode
			fmt.Printf("Found %s %s on node %s\n", instanceType, vmid, node)
		} else {
			// Otherwise, search all nodes
			fmt.Printf("Locating instance %s...\n", vmid)
			var err error
			node, instanceType, err = client.FindNodeByVMID(vmid)
			if err != nil {
				return err
			}
			fmt.Printf("Found %s %s on node %s\n", instanceType, vmid, node)
		}

		// Determine source node (for API call/ticket creation)
		sourceNode := node
		if consoleSourceNode != "" && consoleSourceNode != node {
			sourceNode = consoleSourceNode
			fmt.Printf("Using different source node for ticket creation: %s\n", sourceNode)
		}

		// Determine target node (for WebSocket connection)
		targetNode := node
		if consoleTargetNode != "" && consoleTargetNode != node {
			targetNode = consoleTargetNode
			fmt.Printf("Using different target node for WebSocket: %s\n", targetNode)
		}

		// Create terminal proxy session
		fmt.Println("Creating console session...")
		startTime := time.Now()
		// Use the source node for the termproxy request (ticket creation)
		proxyResp, err := client.CreateTermProxy(sourceNode, vmid, instanceType)
		if err != nil {
			return fmt.Errorf("create terminal proxy: %w", err)
		}

		if config.Debug() {
			fmt.Printf("DEBUG: TermProxy response: UPID=%s, Port=%d, User=%s, Ticket=%s\n", proxyResp.UPID, proxyResp.Port, proxyResp.User, proxyResp.Ticket)
			fmt.Printf("DEBUG: Time after termproxy: %v\n", time.Since(startTime))
		}

		// Determine WebSocket connection URL
		nodeURL := cfg.APIURL
		if consoleNodeOverride != "" {
			// Use override host if specified
			u, _ := url.Parse(cfg.APIURL)
			u.Host = fmt.Sprintf("%s:8006", consoleNodeOverride)
			nodeURL = u.String()
			if config.Debug() {
				fmt.Printf("DEBUG: Using --node override: %s\n", consoleNodeOverride)
			}
		} else if config.Debug() {
			u, _ := url.Parse(cfg.APIURL)
			if u.Host != "" {
				fmt.Printf("DEBUG: Using API URL host: %s\n", u.Host)
			}
		}

		// Connect to VNC WebSocket
		fmt.Println("Connecting to console...")
		if config.Debug() {
			fmt.Printf("DEBUG: Time before WebSocket dial: %v\n", time.Since(startTime))
		}
		vnc, err := console.Connect(
			nodeURL,
			targetNode,
			vmid,
			instanceType,
			proxyResp.Port,
			proxyResp.Ticket,
			cfg.InsecureSkipVerify,
			config.Debug(),
		)
		if err != nil {
			return fmt.Errorf("connect to console: %w", err)
		}
		defer vnc.Close()

		fmt.Println("Connected! Press Ctrl+C to exit.")
		fmt.Println("---")
		// Start interactive console session
		if err := vnc.Start(); err != nil {
			// Don't print error if it's just a normal close
			if err.Error() != "EOF" && err.Error() != "websocket: close 1000 (normal)" {
				fmt.Fprintf(os.Stderr, "\nConsole error: %v\n", err)
			}
		}
		fmt.Println("\n---")
		fmt.Println("Console session closed.")
		return nil
	},
}

func init() {
	consoleCmd.Flags().StringVar(&consoleNodeOverride, "node", "", "Override node to connect to (useful for testing different nodes)")
	consoleCmd.Flags().StringVar(&consoleSourceNode, "source", "", "Source node to create the ticket from (where the API call is made)")
	consoleCmd.Flags().StringVar(&consoleTargetNode, "target", "", "Target node to connect the WebSocket to (where the console connection is established)")
	rootCmd.AddCommand(consoleCmd)
}
