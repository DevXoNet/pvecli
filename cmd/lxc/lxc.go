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

package lxc

import (
	"fmt"

	"github.com/spf13/cobra"

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"
)

var lxcCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "lxc",
	Short:         "LXC container operations",
	Long:          "Manage LXC containers and templates",
}

// lxc templates
var lxcTemplatesCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "templates [storage]",
	Short:         "List LXC templates",
	Long: `List LXC container templates in compressed formats (txz, tzst, tgz).

LXC templates are stored in template storages and used to create new containers.

Examples:
  pvecli lxc templates           # All storages
  pvecli lxc templates local     # Specific storage
  pvecli lxc templates NFStorage # NFS storage`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			output.PrintError(fmt.Errorf("config error: %w", err))
			return nil
		}
		client := pve.NewClient(cfg)

		// Get all storages from datacenter (faster than per-node)
		storages, err := client.GetAllStorages()
		if err != nil {
			output.PrintError(err)
			return nil
		}

		// Get first node for querying content
		nodes, err := client.GetNodes()
		if err != nil || len(nodes) == 0 {
			output.PrintError(fmt.Errorf("no nodes available"))
			return nil
		}
		firstNode := nodes[0].Node

		lxcTemplates := make([]map[string]interface{}, 0)
		seenTemplates := make(map[string]bool) // Avoid duplicates

		for _, storage := range storages {
			storageMap, ok := storage.(map[string]interface{})
			if !ok {
				continue
			}

			storageName, _ := storageMap["storage"].(string)

			// Filter by storage if specified
			if len(args) > 0 && storageName != args[0] {
				continue
			}

			// Check if storage supports Container template
			contentTypes, ok := storageMap["content"].(string)
			if !ok || !contains(contentTypes, "vztmpl") {
				continue
			}

			// Skip storages that don't have Container template enabled
			// Content field contains comma-separated values like "backup,vztmpl,iso"
			if contentTypes == "" {
				continue
			}

			// Get content from storage (use first node for shared storages)
			content, err := client.ListBackups(firstNode, storageName, "")
			if err != nil {
				continue
			}

			// Filter for LXC templates (txz, tzst, tgz formats)
			for _, item := range content {
				if contentType, ok := item["content"].(string); ok && contentType == "vztmpl" {
					if format, ok := item["format"].(string); ok {
						if format == "txz" || format == "tzst" || format == "tgz" {
							volid, _ := item["volid"].(string)

							// Skip if already seen (avoid duplicates from shared storage)
							if seenTemplates[volid] {
								continue
							}
							seenTemplates[volid] = true

							template := map[string]interface{}{
								"storage": storageName,
								"volid":   volid,
								"size":    item["size"],
								"format":  format,
							}
							lxcTemplates = append(lxcTemplates, template)
						}
					}
				}
			}
		}

		if len(lxcTemplates) == 0 {
			return output.Print(map[string]interface{}{
				"message":   "No LXC templates found",
				"templates": []string{},
			})
		} else {
			return output.Print(map[string]interface{}{
				"count":     len(lxcTemplates),
				"templates": lxcTemplates,
			})
		}
	},
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || len(s) > len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	lxcCmd.AddCommand(lxcTemplatesCmd)
}

// Init registers all LXC commands
func Init(root *cobra.Command) {
	root.AddCommand(lxcCmd)
}
