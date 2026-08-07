// Copyright 2025 DevXo part of vByte Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package ceph

import (
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
)

func stringValue(value interface{}) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprint(value)
}

func numberValue(value interface{}) float64 {
	switch value := value.(type) {
	case float64:
		return value
	case int:
		return float64(value)
	case int64:
		return float64(value)
	default:
		return 0
	}
}

func byteSize(value interface{}) string {
	bytes := numberValue(value)
	const (
		gib = 1024 * 1024 * 1024
		tib = 1024 * gib
	)
	if bytes >= tib {
		return fmt.Sprintf("%.2f TiB", bytes/tib)
	}
	if bytes >= gib {
		return fmt.Sprintf("%.2f GiB", bytes/gib)
	}
	return fmt.Sprintf("%.0f B", bytes)
}

func healthData(status map[string]interface{}) (string, map[string]interface{}) {
	health, _ := status["health"].(map[string]interface{})
	value, _ := health["status"].(string)
	return value, health
}

func printHealthText(node string, status map[string]interface{}) error {
	healthStatus, health := healthData(status)
	indicator := "✓"
	if healthStatus != "HEALTH_OK" {
		indicator = "⚠"
	}

	fmt.Printf("\nCEPH HEALTH - %s\n\n", node)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"STATUS", "CHECKS", "MUTES"})
	checks, _ := health["checks"].(map[string]interface{})
	mutes, _ := health["mutes"].([]interface{})
	table.Append([]string{
		fmt.Sprintf("%s %s", indicator, healthStatus),
		fmt.Sprintf("%d", len(checks)),
		fmt.Sprintf("%d", len(mutes)),
	})
	table.Render()
	return nil
}

func printStatusText(node string, status map[string]interface{}) error {
	healthStatus, _ := healthData(status)
	osdmap, _ := status["osdmap"].(map[string]interface{})
	pgmap, _ := status["pgmap"].(map[string]interface{})
	quorumNames, _ := status["quorum_names"].([]interface{})

	fmt.Printf("\nCEPH CLUSTER STATUS - %s\n\n", node)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"HEALTH", "MON QUORUM", "OSDs", "PGs", "CAPACITY", "USED"})
	table.Append([]string{
		healthStatus,
		fmt.Sprintf("%d (%s)", len(quorumNames), joinValues(quorumNames)),
		fmt.Sprintf("%.0f up / %.0f in / %.0f total", numberValue(osdmap["num_up_osds"]), numberValue(osdmap["num_in_osds"]), numberValue(osdmap["num_osds"])),
		fmt.Sprintf("%.0f", numberValue(pgmap["num_pgs"])),
		byteSize(pgmap["bytes_total"]),
		byteSize(pgmap["bytes_used"]),
	})
	table.Render()

	if states, ok := pgmap["pgs_by_state"].([]interface{}); ok && len(states) > 0 {
		parts := make([]string, 0, len(states))
		for _, raw := range states {
			state, _ := raw.(map[string]interface{})
			parts = append(parts, fmt.Sprintf("%.0f %s", numberValue(state["count"]), stringValue(state["state_name"])))
		}
		fmt.Printf("\nPG state: %s\n", strings.Join(parts, ", "))
	}
	return nil
}

func joinValues(values []interface{}) string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = append(result, stringValue(value))
	}
	return strings.Join(result, ", ")
}

func collectOSDs(value interface{}, result *[]map[string]interface{}) {
	switch value := value.(type) {
	case map[string]interface{}:
		if value["type"] == "osd" {
			*result = append(*result, value)
		}
		if children, ok := value["children"].([]interface{}); ok {
			for _, child := range children {
				collectOSDs(child, result)
			}
		}
	case []interface{}:
		for _, child := range value {
			collectOSDs(child, result)
		}
	}
}

func printOSDsText(node string, data map[string]interface{}) error {
	var osds []map[string]interface{}
	collectOSDs(data["root"], &osds)

	fmt.Printf("\nCEPH OSD STATUS - %s\n\n", node)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"OSD", "HOST", "STATUS", "IN", "CLASS", "USED", "TOTAL", "PGs", "VERSION"})
	for _, osd := range osds {
		inStatus := "out"
		if numberValue(osd["in"]) == 1 {
			inStatus = "in"
		}
		status := stringValue(osd["status"])
		if status == "up" {
			status = "✓ up"
		} else {
			status = "✗ " + status
		}
		table.Append([]string{
			stringValue(osd["name"]), stringValue(osd["host"]), status, inStatus,
			stringValue(osd["device_class"]), byteSize(osd["bytes_used"]), byteSize(osd["total_space"]),
			fmt.Sprintf("%.0f", numberValue(osd["pgs"])), stringValue(osd["ceph_version_short"]),
		})
	}
	table.Render()
	return nil
}

func printPoolsText(node string, pools []map[string]interface{}) error {
	fmt.Printf("\nCEPH POOLS - %s\n\n", node)
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"POOL", "TYPE", "SIZE", "MIN SIZE", "PGs", "AUTOSCALE", "USED", "USAGE"})
	for _, pool := range pools {
		table.Append([]string{
			stringValue(pool["pool_name"]), stringValue(pool["type"]),
			fmt.Sprintf("%.0f", numberValue(pool["size"])), fmt.Sprintf("%.0f", numberValue(pool["min_size"])),
			fmt.Sprintf("%.0f", numberValue(pool["pg_num"])), stringValue(pool["pg_autoscale_mode"]),
			byteSize(pool["bytes_used"]), fmt.Sprintf("%.1f%%", numberValue(pool["percent_used"])*100),
		})
	}
	table.Render()
	return nil
}
