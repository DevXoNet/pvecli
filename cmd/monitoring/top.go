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

package monitoring

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"pvecli/config"
	"pvecli/internal/pve"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var (
	topRefreshInterval int
	topSortBy          string
	previousStats      map[string]vmStats // Track previous stats for rate calculation
)

var topCmd = &cobra.Command{
	SilenceUsage:  true,
	SilenceErrors: true,
	Use:           "top",
	Short:         "Real-time monitoring of VMs and containers",
	Long:          "Display real-time CPU, RAM, disk I/O, and network statistics for all VMs and containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		client := pve.NewClient(cfg)

		// Initialize previous stats map
		previousStats = make(map[string]vmStats)

		// Clear screen and hide cursor
		clearScreen()
		hideCursor()
		defer showCursor()

		// Set terminal to raw mode for key detection
		exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
		defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()

		// Channel for quit signal
		quit := make(chan bool)

		// Goroutine to listen for 'q' key
		go func() {
			b := make([]byte, 1)
			for {
				os.Stdin.Read(b)
				if b[0] == 'q' || b[0] == 'Q' {
					quit <- true
					return
				}
			}
		}()

		// Main monitoring loop
		for {
			// Get all resources at once (much faster than individual queries)
			resources, err := client.GetClusterResources()
			if err != nil {
				return err
			}

			// Collect stats from cluster resources
			var allStats []vmStats
			for _, resource := range resources {
				resMap, ok := resource.(map[string]interface{})
				if !ok {
					continue
				}

				// Only process VMs and containers
				resType, _ := resMap["type"].(string)
				if resType != "qemu" && resType != "lxc" {
					continue
				}

				// Extract stats directly from resource
				stats := vmStats{}

				if vmid, ok := resMap["vmid"].(float64); ok {
					stats.VMID = int(vmid)
				}
				if name, ok := resMap["name"].(string); ok {
					stats.Name = name
				}
				if node, ok := resMap["node"].(string); ok {
					stats.Node = node
				}
				if status, ok := resMap["status"].(string); ok {
					stats.Status = status
				}
				if cpu, ok := resMap["cpu"].(float64); ok {
					stats.CPUUsage = cpu * 100
				}
				if cpus, ok := resMap["maxcpu"].(float64); ok {
					stats.CPUCores = int(cpus)
				}
				if mem, ok := resMap["mem"].(float64); ok {
					stats.MemUsage = uint64(mem)
				}
				if maxmem, ok := resMap["maxmem"].(float64); ok {
					stats.MemTotal = uint64(maxmem)
				}
				if diskread, ok := resMap["diskread"].(float64); ok {
					stats.DiskRead = uint64(diskread)
				}
				if diskwrite, ok := resMap["diskwrite"].(float64); ok {
					stats.DiskWrite = uint64(diskwrite)
				}
				if netin, ok := resMap["netin"].(float64); ok {
					stats.NetIn = uint64(netin)
				}
				if netout, ok := resMap["netout"].(float64); ok {
					stats.NetOut = uint64(netout)
				}

				if resType == "qemu" {
					stats.Type = "VM"
				} else {
					stats.Type = "CT"
				}

				allStats = append(allStats, stats)
			}

			// Calculate rates based on previous stats
			calculateRates(allStats)

			// Sort stats
			sortStats(allStats, topSortBy)

			// Display
			clearScreen()
			displayTop(allStats)

			// Wait for refresh interval or quit signal
			select {
			case <-quit:
				return nil
			case <-time.After(time.Duration(topRefreshInterval) * time.Second):
				// Continue to next iteration
			}
		}
	},
}

type vmStats struct {
	VMID          int
	Name          string
	Node          string
	Type          string
	Status        string
	CPUUsage      float64
	CPUCores      int
	MemUsage      uint64
	MemTotal      uint64
	DiskRead      uint64 // Total bytes read
	DiskWrite     uint64 // Total bytes written
	NetIn         uint64 // Total bytes in
	NetOut        uint64 // Total bytes out
	DiskReadRate  uint64 // Bytes per second
	DiskWriteRate uint64 // Bytes per second
	NetInRate     uint64 // Bytes per second
	NetOutRate    uint64 // Bytes per second
}

func calculateRates(stats []vmStats) {
	for i := range stats {
		// Create unique key for this VM/CT
		key := fmt.Sprintf("%s-%d", stats[i].Node, stats[i].VMID)

		// Check if we have previous stats
		if prev, exists := previousStats[key]; exists && stats[i].Status == "running" {
			// Calculate rates (bytes per refresh interval, then convert to per second)
			timeDelta := float64(topRefreshInterval)
			if timeDelta == 0 {
				timeDelta = 1 // Prevent division by zero
			}

			// Disk read rate
			if stats[i].DiskRead >= prev.DiskRead {
				delta := stats[i].DiskRead - prev.DiskRead
				stats[i].DiskReadRate = uint64(float64(delta) / timeDelta)
			} else {
				// Counter reset
				stats[i].DiskReadRate = 0
			}

			// Disk write rate
			if stats[i].DiskWrite >= prev.DiskWrite {
				stats[i].DiskWriteRate = uint64(float64(stats[i].DiskWrite-prev.DiskWrite) / timeDelta)
			} else {
				stats[i].DiskWriteRate = 0
			}

			// Network in rate
			if stats[i].NetIn >= prev.NetIn {
				stats[i].NetInRate = uint64(float64(stats[i].NetIn-prev.NetIn) / timeDelta)
			} else {
				stats[i].NetInRate = 0
			}

			// Network out rate
			if stats[i].NetOut >= prev.NetOut {
				stats[i].NetOutRate = uint64(float64(stats[i].NetOut-prev.NetOut) / timeDelta)
			} else {
				stats[i].NetOutRate = 0
			}
		}

		// Store current stats for next iteration
		previousStats[key] = stats[i]
	}
}

func sortStats(stats []vmStats, sortBy string) {
	sort.Slice(stats, func(i, j int) bool {
		switch sortBy {
		case "cpu":
			return stats[i].CPUUsage > stats[j].CPUUsage
		case "mem":
			return stats[i].MemUsage > stats[j].MemUsage
		case "disk":
			return (stats[i].DiskRead + stats[i].DiskWrite) > (stats[j].DiskRead + stats[j].DiskWrite)
		case "net":
			return (stats[i].NetIn + stats[i].NetOut) > (stats[j].NetIn + stats[j].NetOut)
		default:
			return stats[i].VMID < stats[j].VMID
		}
	})
}

func displayTop(stats []vmStats) {
	// Header - fixed width 100 chars
	borderTop := "╔══════════════════════════════════════════════════════════════════════════════════════════════════╗"
	borderBottom := "╚══════════════════════════════════════════════════════════════════════════════════════════════════╝"

	fmt.Printf("\n%s\n", borderTop)

	// Title line
	title := "PVECLI TOP - Real-time VM/Container Monitor"
	totalWidth := 98 // 100 - 2 for borders
	paddingLeft := (totalWidth - len(title)) / 2
	paddingRight := totalWidth - len(title) - paddingLeft
	fmt.Printf("║%s%s%s║\n", strings.Repeat(" ", paddingLeft), title, strings.Repeat(" ", paddingRight))

	// Info line
	infoLine := fmt.Sprintf("Refresh: %ds | Sort: %s | Time: %s | Press 'q' or Ctrl+C to exit",
		topRefreshInterval, topSortBy, time.Now().Format("2006-01-02 15:04:05"))
	paddingLeft = (totalWidth - len(infoLine)) / 2
	paddingRight = totalWidth - len(infoLine) - paddingLeft
	fmt.Printf("║%s%s%s║\n", strings.Repeat(" ", paddingLeft), infoLine, strings.Repeat(" ", paddingRight))

	fmt.Printf("%s\n\n", borderBottom)

	// Summary
	totalVMs := 0
	runningVMs := 0
	totalCPU := 0.0
	totalMem := uint64(0)
	totalMemUsed := uint64(0)

	for _, s := range stats {
		if s.Type == "VM" {
			totalVMs++
		}
		if s.Status == "running" {
			runningVMs++
			totalCPU += s.CPUUsage
			totalMem += s.MemTotal
			totalMemUsed += s.MemUsage
		}
	}

	avgCPU := 0.0
	if runningVMs > 0 {
		avgCPU = totalCPU / float64(runningVMs)
	}

	memPercent := 0.0
	if totalMem > 0 {
		memPercent = float64(totalMemUsed) / float64(totalMem) * 100
	}

	fmt.Printf("VMs/CTs: %d total, %d running | Avg CPU: %.1f%% | Total Memory: %s / %s (%.1f%%)\n\n",
		len(stats), runningVMs, avgCPU,
		formatBytes(totalMemUsed), formatBytes(totalMem), memPercent)

	// Table
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "NAME", "NODE", "TYPE", "STATUS", "CPU%", "CPU", "MEMORY", "MEM%", "DISK I/O", "NET I/O"})
	table.SetBorder(false)
	table.SetAutoWrapText(false)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_RIGHT,  // ID
		tablewriter.ALIGN_LEFT,   // NAME
		tablewriter.ALIGN_LEFT,   // NODE
		tablewriter.ALIGN_CENTER, // TYPE
		tablewriter.ALIGN_CENTER, // STATUS
		tablewriter.ALIGN_RIGHT,  // CPU%
		tablewriter.ALIGN_CENTER, // CPU
		tablewriter.ALIGN_RIGHT,  // MEMORY
		tablewriter.ALIGN_RIGHT,  // MEM%
		tablewriter.ALIGN_RIGHT,  // DISK I/O
		tablewriter.ALIGN_RIGHT,  // NET I/O
	})

	for _, s := range stats {
		// Skip stopped VMs - show only running
		if s.Status != "running" {
			continue
		}

		// CPU bar
		cpuBar := makeBar(s.CPUUsage, 100, 10)

		// Memory percentage
		memPercent := 0.0
		if s.MemTotal > 0 {
			memPercent = float64(s.MemUsage) / float64(s.MemTotal) * 100
		}
		memBar := makeBar(memPercent, 100, 10)

		// Format disk I/O rates
		diskIO := "-"
		if s.DiskReadRate > 0 || s.DiskWriteRate > 0 {
			diskIO = fmt.Sprintf("R:%s/s W:%s/s", formatBytes(s.DiskReadRate), formatBytes(s.DiskWriteRate))
		}

		// Format network I/O rates
		netIO := "-"
		if s.NetInRate > 0 || s.NetOutRate > 0 {
			netIO = fmt.Sprintf("↓%s/s ↑%s/s", formatBytes(s.NetInRate), formatBytes(s.NetOutRate))
		}

		table.Append([]string{
			fmt.Sprintf("%d", s.VMID),
			truncate(s.Name, 15),
			s.Node,
			s.Type,
			colorStatus(s.Status),
			fmt.Sprintf("%.1f", s.CPUUsage),
			cpuBar,
			fmt.Sprintf("%s/%s", formatBytes(s.MemUsage), formatBytes(s.MemTotal)),
			memBar,
			diskIO,
			netIO,
		})
	}

	table.Render()

	// Footer
	fmt.Println()
	fmt.Printf("\033[2m") // Dim text
	fmt.Println("Options: --sort-by cpu|mem|disk|net|id  --interval <seconds>")
	fmt.Printf("\033[0m") // Reset
}

func makeBar(value, max float64, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}

	filled := int(value / max * float64(width))
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("%s %.1f%%", bar, value)
}

func colorStatus(status string) string {
	switch status {
	case "running":
		return "✓ RUN"
	case "stopped":
		return "✗ STOP"
	case "paused":
		return "⏸ PAUSE"
	default:
		return status
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func clearScreen() {
	cmd := exec.Command("clear")
	if os.Getenv("OS") == "Windows_NT" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\033[?25h")
}

func init() {
	topCmd.Flags().IntVarP(&topRefreshInterval, "interval", "i", 5, "Refresh interval in seconds")
	topCmd.Flags().StringVarP(&topSortBy, "sort-by", "s", "id", "Sort by: cpu, mem, disk, net, id")
}
