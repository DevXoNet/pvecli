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
	"time"

	"pvecli/config"
	"pvecli/internal/proxmox"

	"github.com/spf13/cobra"
)

var (
	dryRun bool
)

var clusterUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update all cluster nodes (apt update && apt upgrade -y)",
	Long:  "Run system updates on all nodes in the Proxmox cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := proxmox.NewClient(cfg)

		// Get all nodes
		nodes, err := client.GetNodes()
		if err != nil {
			return err
		}

		if dryRun {
			fmt.Println("╔════════════════════════════════════════════════════════════════╗")
			fmt.Println("║                    DRY RUN MODE - NO CHANGES                   ║")
			fmt.Println("╚════════════════════════════════════════════════════════════════╝")
			fmt.Println()
		}

		fmt.Println("Starting cluster-wide system update...")
		fmt.Printf("Found %d nodes to update\n\n", len(nodes))

		successCount := 0
		failCount := 0

		for _, node := range nodes {
			fmt.Printf("Node: %s\n", node.Node)

			if dryRun {
				fmt.Println("  → Would run: apt update")
				fmt.Println("  → Would run: apt upgrade -y")
				fmt.Println("  ✓ Dry run completed")
				fmt.Println()
				successCount++
				continue
			}

			fmt.Println("  → Running apt update...")

			err := client.RunNodeCommand(node.Node, "apt update")
			if err != nil {
				fmt.Printf("  ✗ Failed to update package list: %v\n\n", err)
				failCount++
				continue
			}
			fmt.Println("  ✓ Package list updated")

			// Small delay between commands
			time.Sleep(1 * time.Second)

			fmt.Println("  → Running apt upgrade -y...")
			err = client.RunNodeCommand(node.Node, "apt upgrade -y")
			if err != nil {
				fmt.Printf("  ✗ Failed to upgrade packages: %v\n\n", err)
				failCount++
				continue
			}

			fmt.Println("  ✓ Packages upgraded successfully")
			fmt.Println()
			successCount++
		}

		fmt.Println("═══════════════════════════════════════")
		if dryRun {
			fmt.Println("Dry Run Summary:")
			fmt.Printf("  Would update: %d nodes\n", successCount)
		} else {
			fmt.Println("Update Summary:")
			fmt.Printf("  ✓ Success: %d nodes\n", successCount)
			if failCount > 0 {
				fmt.Printf("  ✗ Failed:  %d nodes\n", failCount)
			}
		}
		fmt.Println("═══════════════════════════════════════")

		return nil
	},
}

func init() {
	clusterUpdateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without executing")
	clusterCmd.AddCommand(clusterUpdateCmd)
}
