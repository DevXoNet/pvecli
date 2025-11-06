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
	"strings"
	"time"

	"pvecli/config"
	"pvecli/internal/proxmox"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	Use:   "list",
	Short: "List recent tasks",
	Long:  "List recent tasks from all nodes or a specific node",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

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

		// Output based on configured format
		outputFormat := config.GetOutputFormat()

		switch outputFormat {
		case config.OutputFormatJSON:
			b, err := json.MarshalIndent(allTasks, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			b, err := yaml.Marshal(allTasks)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
			fmt.Print(string(b))

		case config.OutputFormatText:
			for _, task := range allTasks {
				upid := task["upid"]
				taskType := task["type"]
				status := task["status"]
				node := task["node"]
				user := task["user"]
				
				fmt.Printf("%-50s | %-10s | %-10s | %-10s | %s\n", 
					upid, taskType, status, node, user)
			}
		}

		return nil
	},
}

var taskStatusCmd = &cobra.Command{
	Use:   "status <upid>",
	Short: "Get task status",
	Long:  "Get the current status of a task by its UPID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		upid := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

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

		// Output based on configured format
		outputFormat := config.GetOutputFormat()

		switch outputFormat {
		case config.OutputFormatJSON:
			b, err := json.MarshalIndent(status, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			b, err := yaml.Marshal(status)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
			fmt.Print(string(b))

		case config.OutputFormatText:
			fmt.Printf("Task: %s\n", upid)
			fmt.Printf("Status: %s\n", status["status"])
			if exitStatus, ok := status["exitstatus"]; ok && exitStatus != nil {
				fmt.Printf("Exit Status: %s\n", exitStatus)
			}
			if startTime, ok := status["starttime"]; ok && startTime != nil {
				fmt.Printf("Start Time: %v\n", startTime)
			}
			if endTime, ok := status["endtime"]; ok && endTime != nil {
				fmt.Printf("End Time: %v\n", endTime)
			}
		}

		return nil
	},
}

var taskLogCmd = &cobra.Command{
	Use:   "log <upid>",
	Short: "Get task log",
	Long:  "Get the log output of a task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		upid := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

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

		// Output based on configured format
		outputFormat := config.GetOutputFormat()

		switch outputFormat {
		case config.OutputFormatJSON:
			output := map[string]interface{}{
				"upid": upid,
				"log":  log,
			}
			b, err := json.MarshalIndent(output, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}
			fmt.Println(string(b))

		case config.OutputFormatYAML:
			output := map[string]interface{}{
				"upid": upid,
				"log":  log,
			}
			b, err := yaml.Marshal(output)
			if err != nil {
				return fmt.Errorf("marshal yaml: %w", err)
			}
			fmt.Print(string(b))

		case config.OutputFormatText:
			for _, line := range log {
				fmt.Println(line)
			}
		}

		return nil
	},
}

var taskWaitCmd = &cobra.Command{
	Use:   "wait <upid>",
	Short: "Wait for task completion",
	Long:  "Wait for a task to complete and return its final status",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		upid := args[0]

		cfg, err := config.LoadConfig()
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		client := proxmox.NewClient(cfg)

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
				// Task completed
				outputFormat := config.GetOutputFormat()

				switch outputFormat {
				case config.OutputFormatJSON:
					b, err := json.MarshalIndent(status, "", "  ")
					if err != nil {
						return fmt.Errorf("marshal json: %w", err)
					}
					fmt.Println(string(b))

				case config.OutputFormatYAML:
					b, err := yaml.Marshal(status)
					if err != nil {
						return fmt.Errorf("marshal yaml: %w", err)
					}
					fmt.Print(string(b))

				case config.OutputFormatText:
					exitStatus := "unknown"
					if es, ok := status["exitstatus"].(string); ok {
						exitStatus = es
					}
					if exitStatus == "OK" {
						fmt.Printf("Task completed successfully\n")
					} else {
						fmt.Printf("Task failed with exit status: %s\n", exitStatus)
					}
				}

				// Return error if task failed
				if exitStatus, ok := status["exitstatus"].(string); ok && exitStatus != "OK" {
					return fmt.Errorf("task failed with exit status: %s", exitStatus)
				}

				return nil
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

func init() {
	taskListCmd.Flags().IntVarP(&taskLimit, "limit", "l", 5, "Number of tasks to show")
	taskListCmd.Flags().StringVarP(&taskNode, "node", "n", "", "Filter by node")
	
	taskLogCmd.Flags().IntVarP(&taskLines, "lines", "n", 0, "Number of lines to show (0 = all)")
	
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskStatusCmd)
	taskCmd.AddCommand(taskLogCmd)
	taskCmd.AddCommand(taskWaitCmd)
	rootCmd.AddCommand(taskCmd)
}
