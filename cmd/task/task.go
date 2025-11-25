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

package task

import (
	"fmt"
	"strings"
	"time"

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"

	"github.com/spf13/cobra"
)

var (
	taskFollow bool
	taskLines  int
	taskLimit  int
	taskNode   string
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Task management commands",
	Long:  "Manage and monitor Proxmox tasks (backup, migration, etc.)",
}

var taskListCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "list",
	Short:         "List recent tasks",
	Long:          "List recent tasks from all nodes or a specific node",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := pve.NewClient(cfg)

		var allTasks []map[string]interface{}

		if taskNode != "" {
			// Get tasks from specified node only
			tasks, err := client.GetTasks(taskNode, taskLimit)
			if err != nil {
				return fmt.Errorf("failed to get tasks: %w", err)
			}
			allTasks = tasks
		} else {
			// Get tasks from all cluster nodes
			nodes, err := client.GetNodes()
			if err != nil {
				return err
			}

			// Collect tasks from each node
			for _, node := range nodes {
				tasks, err := client.GetTasks(node.Node, taskLimit)
				if err != nil {
					continue
				}
				allTasks = append(allTasks, tasks...)
			}

			// Sort by start time (newest first) and limit to requested count
			if len(allTasks) > taskLimit {
				// Bubble sort by starttime
				for i := 0; i < len(allTasks)-1; i++ {
					for j := i + 1; j < len(allTasks); j++ {
						ti, _ := allTasks[i]["starttime"].(float64)
						tj, _ := allTasks[j]["starttime"].(float64)
						if tj > ti {
							allTasks[i], allTasks[j] = allTasks[j], allTasks[i]
						}
					}
				}
				allTasks = allTasks[:taskLimit]
			}
		}

		// Output using output package
		return output.Print(allTasks)
	},
}

var taskStatusCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "status <upid>",
	Short:         "Get task status",
	Long:          "Get the current status of a task by its UPID",
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		upid := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := pve.NewClient(cfg)

		// Parse UPID to extract node
		node, err := parseUPIDNode(upid)
		if err != nil {
			return err
		}

		// Get task status
		status, err := client.GetTaskStatus(node, upid)
		if err != nil {
			return fmt.Errorf("failed to get task status: %w", err)
		}

		// Output using output package
		return output.Print(status)
	},
}

var taskLogCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "log <upid>",
	Short:         "Get task log",
	Long:          "Get the log output of a task",
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		upid := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := pve.NewClient(cfg)

		// Parse UPID to extract node
		node, err := parseUPIDNode(upid)
		if err != nil {
			return err
		}

		// Get task log
		log, err := client.GetTaskLog(node, upid, taskLines)
		if err != nil {
			return fmt.Errorf("failed to get task log: %w", err)
		}

		// Output using output package
		data := map[string]interface{}{
			"upid": upid,
			"log":  log,
		}
		return output.Print(data)
	},
}

var taskWaitCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "wait <upid>",
	Short:         "Wait for task completion",
	Long:          "Wait for a task to complete and return its final status",
	Args:          cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		upid := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := pve.NewClient(cfg)

		// Parse UPID to extract node
		node, err := parseUPIDNode(upid)
		if err != nil {
			return err
		}

		// Wait for task completion
		for {
			status, err := client.GetTaskStatus(node, upid)
			if err != nil {
				return fmt.Errorf("failed to get task status: %w", err)
			}

			taskStatus, ok := status["status"].(string)
			if !ok {
				return fmt.Errorf("invalid task status response")
			}

			if taskStatus == "stopped" {
				// Task completed - check exit status
				exitStatus := "unknown"
				if es, ok := status["exitstatus"].(string); ok {
					exitStatus = es
				}

				if exitStatus == "OK" {
					// Success
					data := map[string]interface{}{
						"upid":   upid,
						"status": status,
					}
					return output.PrintSuccess("Task completed successfully", data)
				} else {
					// Failed - output status and return error
					output.Print(status)
					return fmt.Errorf("task failed with exit status: %s", exitStatus)
				}
			}

			// Wait before checking again
			time.Sleep(2 * time.Second)
		}
	},
}

// parseUPIDNode extracts the node name from a UPID
// UPID format: UPID:node:PID:PSTART:STARTTIME:TYPE:ID:USER:
func parseUPIDNode(upid string) (string, error) {
	parts := strings.Split(upid, ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid UPID format")
	}
	return parts[1], nil
}


// Init registers all task commands
func Init(root *cobra.Command) {
root.AddCommand(taskCmd)
}
