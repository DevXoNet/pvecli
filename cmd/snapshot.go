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

	"pvecli/config"
	"pvecli/internal/proxmox"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	snapshotPBS bool
	snapshotNFS bool
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot <vmid>",
	Short: "Create a backup snapshot to PBS or NFS storage",
	Long: `Create a backup snapshot of a VM or container to Proxmox Backup Server or NFS storage.

The command automatically detects the first available PBS or NFS storage on the node.

Example:
  pvecli snapshot 230 --pbs
  pvecli snapshot 202 --nfs`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]

		// Load configuration
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

		// Validate that exactly one storage type flag is specified
		if !snapshotPBS && !snapshotNFS {
			return fmt.Errorf("either --pbs or --nfs flag is required")
		}
		if snapshotPBS && snapshotNFS {
			return fmt.Errorf("cannot use both --pbs and --nfs flags together")
		}

		// Find which node hosts this VM/container
		node, instanceType, err := client.FindNodeByVMID(vmid)
		if err != nil {
			return err
		}

		// Get list of all storages on the node
		storages, err := client.GetStorages(node)
		if err != nil {
			return fmt.Errorf("failed to get storages: %w", err)
		}

		// Determine which storage type to search for
		var targetStorage string
		var targetType string

		if snapshotPBS {
			targetType = "pbs"
		} else if snapshotNFS {
			targetType = "nfs"
		}

		// Find first storage matching the requested type
		for _, storage := range storages {
			if storageMap, ok := storage.(map[string]interface{}); ok {
				storageName := ""
				storageType := ""
				if storageMap["storage"] != nil {
					storageName = storageMap["storage"].(string)
				}
				if storageMap["type"] != nil {
					storageType = storageMap["type"].(string)
				}

				// Use first matching storage
				if storageType == targetType {
					targetStorage = storageName
					break
				}
			}
		}

		// Error if no matching storage found
		if targetStorage == "" {
			return fmt.Errorf("no %s storage found on node %s", targetType, node)
		}

		// Create backup task
		taskID, err := client.CreateBackup(node, vmid, instanceType, targetStorage)
		if err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		// Output based on configured format
		outputFormat := config.GetOutputFormat()

		switch outputFormat {
		case config.OutputFormatJSON:
			output := map[string]interface{}{
				"vmid":         vmid,
				"node":         node,
				"type":         instanceType,
				"storage":      targetStorage,
				"storage_type": targetType,
				"task_id":      taskID,
				"status":       "started",
			}
			b, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			output := map[string]interface{}{
				"vmid":         vmid,
				"node":         node,
				"type":         instanceType,
				"storage":      targetStorage,
				"storage_type": targetType,
				"task_id":      taskID,
				"status":       "started",
			}
			b, err := yaml.Marshal(output)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
			fmt.Print(string(b))

		case config.OutputFormatText:
			fmt.Printf("Backup task started: %s\n", taskID)
			fmt.Println("Use 'pvecli task <task_id>' to monitor progress")
		}

		return nil
	},
}

func init() {
	snapshotCmd.Flags().BoolVar(&snapshotPBS, "pbs", false, "Use PBS (Proxmox Backup Server) storage")
	snapshotCmd.Flags().BoolVar(&snapshotNFS, "nfs", false, "Use NFS storage")
	rootCmd.AddCommand(snapshotCmd)
}
