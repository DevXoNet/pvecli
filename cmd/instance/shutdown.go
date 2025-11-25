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

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"

	"github.com/spf13/cobra"
)

var (
	shutdownForce bool
)

var shutdownCmd = &cobra.Command{
	Use:           "shutdown <vmid>",
	Short:         "Shutdown VM or container gracefully",
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		vmid := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := pve.NewClient(cfg)

		// Find node
		node, instanceType, err := FindInstance(client, vmid)
		if err != nil {
			return err
		}

		// Shutdown
		err = client.ShutdownInstance(node, vmid, instanceType, shutdownForce)
		if err != nil {
			return fmt.Errorf("failed to shutdown: %w", err)
		}

		msg := fmt.Sprintf("Instance %s shutdown gracefully", vmid)
		if shutdownForce {
			msg = fmt.Sprintf("Instance %s forcefully shutdown", vmid)
		}

		return output.PrintSuccess(msg, map[string]interface{}{
			"vmid":   vmid,
			"node":   node,
			"type":   instanceType,
			"action": "shutdown",
			"force":  shutdownForce,
		})
	},
}
