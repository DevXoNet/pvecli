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

package oci

import (
	"fmt"

	"github.com/spf13/cobra"

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"
)

var ociCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "oci",
	Short:         "OCI image operations (Proxmox VE 9.1+)",
	Long:          "Query OCI image tags from registries. Note: OCI containers are managed via standard LXC commands.",
}

// oci tags
var ociTagsCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "tags <node> <repository>",
	Short:         "Query available tags for an OCI image repository",
	Long: `Query available tags for an OCI image repository.

Examples:
  pvecli oci tags node01 nginx
  pvecli oci tags node01 docker.io/library/nginx
  pvecli oci tags node01 ghcr.io/linuxserver/nginx`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		node := args[0]
		repository := args[1]

		cfg, err := config.LoadConfig()
		if err != nil {
			output.PrintError(fmt.Errorf("config error: %w", err))
			return nil
		}
		client := pve.NewClient(cfg)

		tags, err := client.QueryOCIRepoTags(node, repository)
		if err != nil {
			output.PrintError(err)
			return nil
		}

		return output.Print(map[string]interface{}{
			"repository": repository,
			"node":       node,
			"tags":       tags,
		})
	},
}

// oci images - only OCI images (tar format)
var ociImagesCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "images [storage]",
	Short:         "List OCI images (tar format only)",
	Long: `List OCI images (Docker/Podman images) stored as tar files.

OCI images are in tar format and stored in template storages.

Examples:
  pvecli oci images              # All storages
  pvecli oci images NFStorage    # Specific storage`,
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

		ociImages := make([]map[string]interface{}, 0)
		seenImages := make(map[string]bool) // Avoid duplicates

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
			if !ok || !containsSubstr(contentTypes, "vztmpl") {
				continue
			}

			// Skip storages that don't have Container template enabled
			if contentTypes == "" {
				continue
			}

			// Get content from storage (use first node for shared storages)
			content, err := client.ListBackups(firstNode, storageName, "")
			if err != nil {
				continue
			}

			// Filter for OCI images (tar format only)
			for _, item := range content {
				if contentType, ok := item["content"].(string); ok && contentType == "vztmpl" {
					if format, ok := item["format"].(string); ok && format == "tar" {
						volid, _ := item["volid"].(string)

						// Skip if already seen (avoid duplicates from shared storage)
						if seenImages[volid] {
							continue
						}
						seenImages[volid] = true

						image := map[string]interface{}{
							"storage": storageName,
							"volid":   volid,
							"size":    item["size"],
							"format":  format,
						}
						ociImages = append(ociImages, image)
					}
				}
			}
		}

		if len(ociImages) == 0 {
			return output.Print(map[string]interface{}{
				"message": "No OCI images found",
				"images":  []string{},
			})
		} else {
			return output.Print(map[string]interface{}{
				"count":  len(ociImages),
				"images": ociImages,
			})
		}
	},
}

// Helper function to check if string contains substring
func containsSubstr(s, substr string) bool {
	if len(s) == 0 || len(substr) == 0 {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	ociCmd.AddCommand(ociTagsCmd)
	ociCmd.AddCommand(ociImagesCmd)

}

// Init registers all OCI commands
func Init(root *cobra.Command) {
	root.AddCommand(ociCmd)
}
