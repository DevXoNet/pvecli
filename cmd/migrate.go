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

//TODO fix the output

package cmd

import (
	"fmt"
	"strconv"

	"pvecli/config"
	"pvecli/internal/proxmox"

	"github.com/spf13/cobra"
)

var (
	migrateTarget  string
	migrateOnline  bool
	migrateWait    bool
	migrateTimeout int
)

var migrateCmd = &cobra.Command{
	Use:          "migrate <vmid/ctid> --target <node>",
	Short:        "Migrate a VM or container to another node",
	SilenceUsage: true,
	Args:         cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrateTarget == "" {
			return fmt.Errorf("--target is required")
		}

		id, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid ID: %s", args[0])
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}
		client := proxmox.NewClient(cfg)

		// Find current node and type
		node, vmType, err := client.FindVM(id)
		if err != nil {
			return err
		}

		fmt.Printf("Migrating %s/%d from %s to %s (online=%v) ...\n", vmType, id, node, migrateTarget, migrateOnline)
		taskID, err := client.Migrate(node, id, vmType, migrateTarget, migrateOnline)
		if err != nil {
			return err
		}

		fmt.Printf("Task started: %s\n", taskID)
		if migrateWait {
			fmt.Println("Waiting for migration to complete...")
			if migrateTimeout <= 0 {
				migrateTimeout = 600
			}
			if err := client.WaitForTask(node, taskID, migrateTimeout); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}
			fmt.Println("✓ Migration completed successfully")
		} else {
			fmt.Println("Hint: use --wait to wait for completion")
		}

		return nil
	},
}

func init() {
	migrateCmd.Flags().StringVar(&migrateTarget, "target", "", "Target node name (required)")
	migrateCmd.Flags().BoolVar(&migrateOnline, "online", false, "Perform live (online) migration")
	migrateCmd.Flags().BoolVar(&migrateWait, "wait", true, "Wait for the migration task to complete")
	migrateCmd.Flags().IntVar(&migrateTimeout, "timeout", 600, "Wait timeout in seconds")

	migrateCmd.MarkFlagRequired("target")

	rootCmd.AddCommand(migrateCmd)
}
