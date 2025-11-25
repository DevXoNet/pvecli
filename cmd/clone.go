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
	"strconv"

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/proxmox"

	"github.com/spf13/cobra"
)

var (
	cloneNewID       int
	cloneName        string
	cloneDescription string
	clonePool        string
	cloneStorage     string
	cloneFull        bool
	cloneStart       bool
)

var cloneCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "clone <template-vmid>",
	Short:         "Clone a VM template",
	Long: `Clone a VM template to create a new VM.

Default mode is Full Clone (independent copy). Linked Clone is optional for same-node fast clones.

Examples:
  # Full clone (default, independent copy)
  pvecli clone 9005 --new-id 107 --name web-server-01

  # Linked clone (same node only, fast and space-efficient)
  pvecli clone 9005 --new-id 107 --name web-server-01 --full=false

  # Clone with custom storage (typically for full clone)
  pvecli clone 9005 --new-id 107 --name web-server-01 --storage local-lvm

  # Clone and start immediately
  pvecli clone 9005 --new-id 107 --name web-server-01 --start`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid template ID: %s", args[0])
		}

		if cloneNewID == 0 {
			return fmt.Errorf("--new-id is required")
		}

		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := proxmox.NewClient(cfg)

		// Pre-check: ensure new ID is not already in use anywhere in the cluster //TODO now work , I need to fix the error output
		if _, _, err := client.FindVM(cloneNewID); err == nil {
			return fmt.Errorf("new VM ID %d is already in use", cloneNewID)
		}

		// Find template VM
		fmt.Printf("Looking for template VM %d...\n", templateID)
		node, vmType, err := client.FindVM(templateID)
		if err != nil {
			return fmt.Errorf("template VM %d not found: %w", templateID, err)
		}

		// Verify it's a template
		vmConfig, err := client.GetVMConfig(node, strconv.Itoa(templateID))
		if err != nil {
			return fmt.Errorf("failed to get template config: %w", err)
		}

		isTemplate := false
		if template, ok := vmConfig["template"]; ok {
			if tplVal, ok := template.(float64); ok && tplVal == 1 {
				isTemplate = true
			}
		}

		if !isTemplate {
			return fmt.Errorf("VM %d is not a template", templateID)
		}

		// Target node is always the source node (cross-node clone via API is not supported here)
		targetNode := node

		// Prepare clone parameters
		cloneMode := "full"
		if !cloneFull {
			cloneMode = "linked"
		}

		// Generate name if not provided
		if cloneName == "" {
			cloneName = fmt.Sprintf("vm-%d", cloneNewID)
		}

		fmt.Printf("\n╔════════════════════════════════════════╗\n")
		fmt.Printf("║         Cloning VM Template            ║\n")
		fmt.Printf("╚════════════════════════════════════════╝\n\n")
		fmt.Printf("Template ID:   %d\n", templateID)
		fmt.Printf("New VM ID:     %d\n", cloneNewID)
		fmt.Printf("Name:          %s\n", cloneName)
		fmt.Printf("Source Node:   %s\n", node)
		fmt.Printf("Target Node:   %s\n", targetNode)
		fmt.Printf("Clone Mode:    %s\n", cloneMode)
		if cloneStorage != "" {
			fmt.Printf("Storage:       %s\n", cloneStorage)
		}
		if clonePool != "" {
			fmt.Printf("Pool:          %s\n", clonePool)
		}
		fmt.Println()

		// Execute clone
		fmt.Printf("Cloning VM %d -> %d...\n", templateID, cloneNewID)

		taskID, err := client.CloneVM(node, templateID, vmType, proxmox.CloneParams{
			NewID:       cloneNewID,
			Name:        cloneName,
			Description: cloneDescription,
			Target:      targetNode,
			Pool:        clonePool,
			Storage:     cloneStorage,
			Full:        cloneFull,
		})
		if err != nil {
			return fmt.Errorf("clone failed: %w", err)
		}

		fmt.Printf("Clone task started: %s\n", taskID)
		fmt.Println("Waiting for clone to complete...")

		// Wait for task to complete
		err = client.WaitForTask(targetNode, taskID, 300) // 5 minute timeout
		if err != nil {
			return fmt.Errorf("clone task failed: %w", err)
		}

		fmt.Printf("\n✓ VM %d cloned successfully to VM %d\n", templateID, cloneNewID)

		// Start VM if requested
		if cloneStart {
			fmt.Printf("\nStarting VM %d...\n", cloneNewID)
			err = client.StartVM(targetNode, cloneNewID, vmType)
			if err != nil {
				return fmt.Errorf("failed to start VM: %w", err)
			}
			fmt.Printf("✓ VM %d started successfully\n", cloneNewID)
		}

		// Prepare output data
		outputFormat := config.GetOutputFormat()
		if outputFormat == config.OutputFormatJSON || outputFormat == config.OutputFormatYAML {
			// Structured output for JSON/YAML
			data := map[string]interface{}{
				"template_id": templateID,
				"new_vm_id":   cloneNewID,
				"name":        cloneName,
				"node":        targetNode,
				"mode":        cloneMode,
				"started":     cloneStart,
			}
			if cloneStorage != "" {
				data["storage"] = cloneStorage
			}
			if clonePool != "" {
				data["pool"] = clonePool
			}
			return output.PrintSuccess(fmt.Sprintf("VM %d cloned successfully to VM %d", templateID, cloneNewID), data)
		}

		// Text output with formatted display
		fmt.Println("\n╔════════════════════════════════════════╗")
		fmt.Println("║           Clone Complete               ║")
		fmt.Println("╚════════════════════════════════════════╝")
		fmt.Printf("\nNew VM ID: %d\n", cloneNewID)
		fmt.Printf("Name:      %s\n", cloneName)
		fmt.Printf("Node:      %s\n", targetNode)
		fmt.Printf("Mode:      %s clone\n", cloneMode)
		if cloneStart {
			fmt.Printf("Status:    running\n")
		} else {
			fmt.Printf("Status:    stopped\n")
			fmt.Printf("\nTo start the VM, run:\n")
			fmt.Printf("  pvecli start %d\n", cloneNewID)
		}

		return nil
	},
}

func init() {
	cloneCmd.Flags().IntVar(&cloneNewID, "new-id", 0, "New VM ID (required)")
	cloneCmd.Flags().StringVar(&cloneName, "name", "", "Name for the new VM")
	cloneCmd.Flags().StringVar(&cloneDescription, "description", "", "Description for the new VM")
	cloneCmd.Flags().StringVar(&clonePool, "pool", "", "Add to resource pool")
	cloneCmd.Flags().StringVar(&cloneStorage, "storage", "", "Target storage (for full clone)")
	cloneFull = true
	cloneCmd.Flags().BoolVar(&cloneFull, "full", cloneFull, "Full clone (set --full=false for linked clone)")
	cloneCmd.Flags().BoolVar(&cloneStart, "start", false, "Start VM after cloning")

	cloneCmd.MarkFlagRequired("new-id")

	rootCmd.AddCommand(cloneCmd)
}
