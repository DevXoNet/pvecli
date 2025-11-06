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
	backupStorage string
	backupVMID    string
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup management commands",
	Long:  "List and manage backups on PBS or NFS storage",
}

var backupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List backups",
	Long:  "List all backups on specified storage (use storage name or type: pbs/nfs)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if backupStorage == "" {
			return fmt.Errorf("--storage flag is required")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

		// Get any node (storage is accessible cluster-wide)
		nodes, err := client.GetNodes()
		if err != nil {
			return err
		}
		if len(nodes) == 0 {
			return fmt.Errorf("no nodes found")
		}

		// Resolve storage name (if type is given, find first matching storage)
		resolvedStorage := backupStorage
		if backupStorage == "pbs" || backupStorage == "nfs" {
			// Auto-detect storage by type
			storages, err := client.GetStorages(nodes[0].Node)
			if err != nil {
				return fmt.Errorf("failed to get storages: %w", err)
			}

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

					if storageType == backupStorage {
						resolvedStorage = storageName
						break
					}
				}
			}

			if resolvedStorage == backupStorage {
				return fmt.Errorf("no %s storage found", backupStorage)
			}
		}

		// List all backups on the resolved storage
		backups, err := client.ListBackups(nodes[0].Node, resolvedStorage, backupVMID)
		if err != nil {
			return fmt.Errorf("failed to list backups: %w", err)
		}

		// Output based on configured format
		outputFormat := config.GetOutputFormat()

		switch outputFormat {
		case config.OutputFormatJSON:
			b, _ := json.MarshalIndent(backups, "", "  ")
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			b, _ := yaml.Marshal(backups)
			fmt.Print(string(b))

		case config.OutputFormatText:
			for _, backup := range backups {
				fmt.Printf("%s\n", backup["volid"])
			}
		}

		return nil
	},
}

var backupDeleteCmd = &cobra.Command{
	Use:   "delete <volid>",
	Short: "Delete a backup",
	Long:  "Delete a backup by its volume ID (use storage name or type: pbs/nfs)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		volid := args[0]

		if backupStorage == "" {
			return fmt.Errorf("--storage flag is required")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

		// Get any node (storage is accessible cluster-wide)
		nodes, err := client.GetNodes()
		if err != nil {
			return err
		}
		if len(nodes) == 0 {
			return fmt.Errorf("no nodes found")
		}

		// Resolve storage name (if type is given, find first matching storage)
		resolvedStorage := backupStorage
		if backupStorage == "pbs" || backupStorage == "nfs" {
			// Auto-detect storage by type
			storages, err := client.GetStorages(nodes[0].Node)
			if err != nil {
				return fmt.Errorf("failed to get storages: %w", err)
			}

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

					if storageType == backupStorage {
						resolvedStorage = storageName
						break
					}
				}
			}

			if resolvedStorage == backupStorage {
				return fmt.Errorf("no %s storage found", backupStorage)
			}
		}

		// Delete the specified backup
		err = client.DeleteBackup(nodes[0].Node, resolvedStorage, volid)
		if err != nil {
			return fmt.Errorf("failed to delete backup: %w", err)
		}

		// Output based on configured format
		outputFormat := config.GetOutputFormat()

		switch outputFormat {
		case config.OutputFormatJSON:
			output := map[string]interface{}{
				"volid":   volid,
				"storage": backupStorage,
				"action":  "delete",
				"status":  "success",
			}
			b, _ := json.MarshalIndent(output, "", "  ")
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			output := map[string]interface{}{
				"volid":   volid,
				"storage": backupStorage,
				"action":  "delete",
				"status":  "success",
			}
			b, _ := yaml.Marshal(output)
			fmt.Print(string(b))

		case config.OutputFormatText:
			fmt.Printf("Backup %s deleted successfully\n", volid)
		}

		return nil
	},
}

func init() {
	backupListCmd.Flags().StringVarP(&backupStorage, "storage", "s", "", "Storage name or type (pbs/nfs) - required")
	backupListCmd.Flags().StringVar(&backupVMID, "vmid", "", "Filter by VMID")
	backupListCmd.MarkFlagRequired("storage")

	backupDeleteCmd.Flags().StringVarP(&backupStorage, "storage", "s", "", "Storage name or type (pbs/nfs) - required")
	backupDeleteCmd.MarkFlagRequired("storage")

	backupCmd.AddCommand(backupListCmd)
	backupCmd.AddCommand(backupDeleteCmd)
	rootCmd.AddCommand(backupCmd)
}
