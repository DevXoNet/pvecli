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
	"os"
	"sort"
	"strings"

	"pvecli/config"
	"pvecli/internal/proxmox"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// progressBar creates a simple ASCII progress bar with color indicators
func progressBar(percent float64, width int) string {
	filled := int(percent / 100.0 * float64(width))
	if filled > width {
		filled = width
	}

	// Use different characters for better visibility
	bar := strings.Repeat("█", filled) + strings.Repeat("·", width-filled)

	// Add color indicator based on usage
	indicator := "✓"
	if percent > 80 {
		indicator = "!"
	} else if percent > 60 {
		indicator = "~"
	}

	return fmt.Sprintf("%s [%s] %.1f%%", indicator, bar, percent)
}

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Show cluster overview and health status",
	Long:  "Display cluster health, nodes status, VM/CT statistics, and resource usage",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := proxmox.NewClient(cfg)

		// Get cluster status for cluster name and node IPs
		clusterStatusData, err := client.GetClusterStatus()
		clusterName := ""
		nodeIPMap := make(map[string]string)

		if err == nil && clusterStatusData != nil {
			// clusterStatusData is the data array
			for _, item := range clusterStatusData {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if itemMap["type"] == "cluster" && itemMap["name"] != nil {
						clusterName = itemMap["name"].(string)
					} else if itemMap["type"] == "node" {
						if itemMap["name"] != nil {
							nodeName := itemMap["name"].(string)
							nodeIP := ""
							if itemMap["ip"] != nil {
								nodeIP = itemMap["ip"].(string)
							}
							nodeIPMap[nodeName] = nodeIP
						}
					}
				}
			}
		}

		// Get all nodes
		nodes, err := client.GetNodes()
		if err != nil {
			return err
		}

		// Get cluster resources for storage info
		clusterResources, err := client.GetClusterResources()

		// Counters
		var (
			nodesOnline, nodesOffline                             int
			totalCPU, usedCPU                                     float64
			totalRAM, usedRAM, totalStorage, usedStorage          float64
			vmRunning, vmStopped, ctRunning, ctStopped, templates int
		)

		// Calculate storage and node resources from cluster resources
		// Use a map to track unique storage pools (avoid counting shared storage multiple times)
		storageMap := make(map[string]struct {
			total float64
			used  float64
		})

		// Map to store node-specific RAM info from cluster resources
		nodeRAMMap := make(map[string]struct {
			used float64
			max  float64
		})

		if err == nil && clusterResources != nil {
			for _, resource := range clusterResources {
				if resMap, ok := resource.(map[string]interface{}); ok {
					// Get storage info
					if resMap["type"] == "storage" {
						// Get storage ID to avoid duplicates
						storageID := ""
						if resMap["storage"] != nil {
							storageID = resMap["storage"].(string)
						}

						if storageID != "" {
							// Only add if not already counted
							if _, exists := storageMap[storageID]; !exists {
								var maxdisk, disk float64
								if resMap["maxdisk"] != nil {
									if val, ok := resMap["maxdisk"].(float64); ok {
										maxdisk = val
									}
								}
								if resMap["disk"] != nil {
									if val, ok := resMap["disk"].(float64); ok {
										disk = val
									}
								}
								storageMap[storageID] = struct {
									total float64
									used  float64
								}{maxdisk, disk}
							}
						}
					}

					// Get node CPU and RAM info
					if resMap["type"] == "node" {
						nodeName := ""
						if resMap["node"] != nil {
							nodeName = resMap["node"].(string)
						}

						if resMap["maxcpu"] != nil {
							if val, ok := resMap["maxcpu"].(float64); ok {
								totalCPU += val
							}
						}
						if resMap["cpu"] != nil {
							if val, ok := resMap["cpu"].(float64); ok {
								usedCPU += val
							}
						}
						if resMap["maxmem"] != nil {
							if val, ok := resMap["maxmem"].(float64); ok {
								totalRAM += val
								if nodeName != "" {
									info := nodeRAMMap[nodeName]
									info.max = val
									nodeRAMMap[nodeName] = info
								}
							}
						}
						if resMap["mem"] != nil {
							if val, ok := resMap["mem"].(float64); ok {
								usedRAM += val
								if nodeName != "" {
									info := nodeRAMMap[nodeName]
									info.used = val
									nodeRAMMap[nodeName] = info
								}
							}
						}
					}
				}
			}

			// Sum up unique storage pools
			for _, storage := range storageMap {
				totalStorage += storage.total
				usedStorage += storage.used
			}
		}

		// Node details for table
		type nodeDetail struct {
			name       string
			ip         string
			status     string
			cpuPercent float64
			ramPercent float64
			uptime     string
		}
		var nodeDetails []nodeDetail

		// Collect node stats
		for _, node := range nodes {
			status, err := client.GetNodeStatus(node.Node)
			if err != nil {
				nodesOffline++
				nodeDetails = append(nodeDetails, nodeDetail{
					name:       node.Node,
					ip:         nodeIPMap[node.Node],
					status:     "offline",
					cpuPercent: 0,
					ramPercent: 0,
					uptime:     "N/A",
				})
				continue
			}

			nodeStatus := "online"
			nodesOnline++

			// Calculate node-specific percentages
			var nodeCPUPercent, nodeRAMPercent float64
			var nodeUptime string

			// CPU - get percentage from node status
			if status["cpu"] != nil {
				if val, ok := status["cpu"].(float64); ok {
					nodeCPUPercent = val * 100
				}
			}

			// RAM - get percentage from cluster resources
			if ramInfo, exists := nodeRAMMap[node.Node]; exists && ramInfo.max > 0 {
				nodeRAMPercent = (ramInfo.used / ramInfo.max) * 100
			}

			// Uptime
			if status["uptime"] != nil {
				if uptimeSec, ok := status["uptime"].(float64); ok {
					days := int(uptimeSec) / 86400
					hours := (int(uptimeSec) % 86400) / 3600
					if days > 0 {
						nodeUptime = fmt.Sprintf("%dd %dh", days, hours)
					} else {
						nodeUptime = fmt.Sprintf("%dh", hours)
					}
				}
			}

			// Get VMs and CTs for this node
			instances, err := client.GetAllInstances(node.Node)
			if err == nil {
				for _, inst := range instances {
					if inst.Template {
						templates++
					} else if inst.Type == "vm" {
						if inst.Status == "running" {
							vmRunning++
						} else {
							vmStopped++
						}
					} else if inst.Type == "ct" {
						if inst.Status == "running" {
							ctRunning++
						} else {
							ctStopped++
						}
					}
				}
			}

			// Add node to details
			nodeDetails = append(nodeDetails, nodeDetail{
				name:       node.Node,
				ip:         nodeIPMap[node.Node],
				status:     nodeStatus,
				cpuPercent: nodeCPUPercent,
				ramPercent: nodeRAMPercent,
				uptime:     nodeUptime,
			})
		}

		// Sort nodes by name
		sort.Slice(nodeDetails, func(i, j int) bool {
			return nodeDetails[i].name < nodeDetails[j].name
		})

		// Print header with wider border (88 chars)
		borderTop := "╔════════════════════════════════════════════════════════════════════════════════════════╗"
		borderBottom := "╚════════════════════════════════════════════════════════════════════════════════════════╝"
		fmt.Printf("\n%s\n", borderTop)

		// Create title - limit cluster name to 10 chars max
		displayName := clusterName
		if len(displayName) > 10 {
			displayName = displayName[:10]
		}

		var title string
		if displayName != "" {
			title = fmt.Sprintf("PROXMOX CLUSTER OVERVIEW - %s", displayName)
		} else {
			title = "PROXMOX CLUSTER OVERVIEW"
		}

		// Center the title in 88 character width
		totalWidth := 88
		if len(title) > totalWidth {
			title = title[:totalWidth]
		}
		paddingLeft := (totalWidth - len(title)) / 2
		paddingRight := totalWidth - len(title) - paddingLeft

		fmt.Printf("║%s%s%s║\n", strings.Repeat(" ", paddingLeft), title, strings.Repeat(" ", paddingRight))
		fmt.Printf("%s\n\n", borderBottom)

		// Cluster status table
		statusTable := tablewriter.NewWriter(os.Stdout)
		statusTable.SetHeader([]string{"CLUSTER STATUS", "NODES", "VMs", "CONTAINERS", "TEMPLATES"})
		statusTable.SetBorder(false)

		healthStatus := "✓ Healthy"
		if nodesOffline > 0 {
			healthStatus = "⚠ Warning"
		}

		statusTable.Append([]string{
			healthStatus,
			fmt.Sprintf("%d online / %d total", nodesOnline, nodesOnline+nodesOffline),
			fmt.Sprintf("%d run / %d stop", vmRunning, vmStopped),
			fmt.Sprintf("%d run / %d stop", ctRunning, ctStopped),
			fmt.Sprintf("%d", templates),
		})
		statusTable.Render()
		fmt.Println()

		// Resources table
		// CPU percent is already the average usage across all CPUs
		cpuPercent := 0.0
		if totalCPU > 0 {
			// usedCPU is the sum of CPU usage percentages, convert to overall percentage
			cpuPercent = (usedCPU / float64(nodesOnline)) * 100
		}

		ramUsedGB := usedRAM / 1024 / 1024 / 1024
		ramTotalGB := totalRAM / 1024 / 1024 / 1024
		ramPercent := 0.0
		if totalRAM > 0 {
			ramPercent = (usedRAM / totalRAM) * 100
		}

		storageUsedTB := usedStorage / 1024 / 1024 / 1024 / 1024
		storageTotalTB := totalStorage / 1024 / 1024 / 1024 / 1024
		storagePercent := 0.0
		if totalStorage > 0 {
			storagePercent = (usedStorage / totalStorage) * 100
		}

		resourceTable := tablewriter.NewWriter(os.Stdout)
		resourceTable.SetHeader([]string{"RESOURCE", "USAGE", "TOTAL", "GRAPH"})
		resourceTable.SetBorder(false)

		resourceTable.Append([]string{
			"CPU",
			fmt.Sprintf("%.1f%%", cpuPercent),
			fmt.Sprintf("%.0f cores", totalCPU),
			progressBar(cpuPercent, 20),
		})
		resourceTable.Append([]string{
			"Memory",
			fmt.Sprintf("%.1f GB", ramUsedGB),
			fmt.Sprintf("%.1f GB", ramTotalGB),
			progressBar(ramPercent, 20),
		})
		resourceTable.Append([]string{
			"Storage",
			fmt.Sprintf("%.2f TiB", storageUsedTB),
			fmt.Sprintf("%.2f TiB", storageTotalTB),
			progressBar(storagePercent, 20),
		})
		resourceTable.Render()
		fmt.Println()

		// Subscription status - commented out
		/*
			subStatus, err := client.GetSubscriptionStatus(nodes[0].Node)
			subStatusStr := "✗ No active subscription"
			if err == nil && subStatus["status"] != nil && subStatus["status"].(string) == "active" {
				subStatusStr = "✓ Active"
				if subStatus["productname"] != nil {
					subStatusStr = fmt.Sprintf("✓ Active (%s)", subStatus["productname"])
				}
			}

			subTable := tablewriter.NewWriter(os.Stdout)
			subTable.SetHeader([]string{"SUBSCRIPTION STATUS"})
			subTable.SetBorder(false)
			subTable.Append([]string{subStatusStr})
			subTable.Render()
			fmt.Println()
		*/

		// Mounted Storage section (PBS, NFS, ZFS, Ceph)
		type mountedStorage struct {
			storageType string
			name        string
			used        float64
			total       float64
		}

		// Use map to avoid duplicates (shared storage appears once per node)
	mountedStorageMap := make(map[string]mountedStorage)

	for _, resource := range clusterResources {
		resMap, ok := resource.(map[string]interface{})
		if !ok {
			continue
		}
		if resMap["type"] != "storage" || resMap["plugintype"] == nil {
			continue
		}

		plugintype := resMap["plugintype"].(string)
		storageName := ""
		if resMap["storage"] != nil {
			storageName = resMap["storage"].(string)
		}

		// Check for PBS, NFS, ZFS, Ceph
		if plugintype == "pbs" || plugintype == "nfs" || plugintype == "zfspool" || plugintype == "rbd" {
			// Skip if already added
			if _, exists := mountedStorageMap[storageName]; exists {
				continue
			}

			var used, total float64
			if resMap["disk"] != nil {
				used = resMap["disk"].(float64)
			}
			if resMap["maxdisk"] != nil {
				total = resMap["maxdisk"].(float64)
			}

			if total > 0 {
				typeLabel := ""
				switch plugintype {
				case "pbs":
					typeLabel = "Proxmox Backup"
				case "nfs":
					typeLabel = "NFS"
				case "zfspool":
					typeLabel = "ZFS"
				case "rbd":
					typeLabel = "Ceph RBD"
				}

				mountedStorageMap[storageName] = mountedStorage{
					storageType: typeLabel,
					name:        storageName,
					used:        used,
					total:       total,
				}
			}
		}
	}

		// Convert map to slice for display
		var mountedStorages []mountedStorage
		for _, ms := range mountedStorageMap {
			mountedStorages = append(mountedStorages, ms)
		}

		// Display mounted storage table
		if len(mountedStorages) > 0 {
			fmt.Println("=== MOUNTED STORAGE ===")
			mountedTable := tablewriter.NewWriter(os.Stdout)
			mountedTable.SetHeader([]string{"TYPE", "NAME", "USED", "TOTAL", "USAGE"})
			mountedTable.SetBorder(false)

			for _, ms := range mountedStorages {
				usedTB := ms.used / 1024 / 1024 / 1024 / 1024
				totalTB := ms.total / 1024 / 1024 / 1024 / 1024
				percent := (ms.used / ms.total) * 100

				mountedTable.Append([]string{
					ms.storageType,
					ms.name,
					fmt.Sprintf("%.2f TiB", usedTB),
					fmt.Sprintf("%.2f TiB", totalTB),
					fmt.Sprintf("%.1f%%", percent),
				})
			}
			mountedTable.Render()
			fmt.Println()
		}

		// Print nodes table
		fmt.Println("=== NODES DETAILS ===")
		nodesTable := tablewriter.NewWriter(os.Stdout)
		nodesTable.SetHeader([]string{"NODE", "IP ADDRESS", "STATUS", "CPU", "MEMORY", "UPTIME"})
		nodesTable.SetBorder(false)

		for _, nd := range nodeDetails {
			statusIcon := "✓"
			if nd.status == "offline" {
				statusIcon = "✗"
			}
			nodesTable.Append([]string{
				nd.name,
				nd.ip,
				fmt.Sprintf("%s %s", statusIcon, nd.status),
				progressBar(nd.cpuPercent, 15),
				progressBar(nd.ramPercent, 15),
				nd.uptime,
			})
		}
		nodesTable.Render()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)
}
