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

	"pvecli/config"
	"pvecli/internal/output"
	"pvecli/internal/pve"

	"github.com/spf13/cobra"
)

var cephNode string

var cephCmd = &cobra.Command{
	Use:   "ceph",
	Short: "Inspect Ceph cluster health and resources",
	Long:  "Read Ceph health, OSD, and pool information through the Proxmox API",
}

func newClient() (*pve.Client, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}
	return pve.NewClient(cfg), nil
}

func cephNodes(client *pve.Client) ([]string, error) {
	if cephNode != "" {
		return []string{cephNode}, nil
	}

	nodes, err := client.GetNodes()
	if err != nil {
		return nil, fmt.Errorf("get cluster nodes: %w", err)
	}

	result := make([]string, 0, len(nodes))
	for _, node := range nodes {
		if node.Status == "online" {
			result = append(result, node.Node)
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no online Proxmox nodes found")
	}
	return result, nil
}

func withCephNode[T any](client *pve.Client, fetch func(string) (T, error)) (string, T, error) {
	var zero T
	nodes, err := cephNodes(client)
	if err != nil {
		return "", zero, err
	}

	var lastErr error
	for _, node := range nodes {
		data, err := fetch(node)
		if err == nil {
			return node, data, nil
		}
		lastErr = err
	}
	return "", zero, fmt.Errorf("ceph API unavailable on checked nodes: %w", lastErr)
}

var statusCmd = &cobra.Command{
	Use:           "status",
	Short:         "Show full Ceph cluster status",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		node, status, err := withCephNode(client, client.GetCephStatus)
		if err != nil {
			return err
		}
		if config.GetOutputFormat() == config.OutputFormatText {
			return printStatusText(node, status)
		}
		return output.Print(map[string]interface{}{"node": node, "status": status})
	},
}

var healthCmd = &cobra.Command{
	Use:           "health",
	Short:         "Show Ceph health summary",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		node, status, err := withCephNode(client, client.GetCephStatus)
		if err != nil {
			return err
		}
		health, ok := status["health"]
		if !ok {
			return fmt.Errorf("ceph status response does not contain health information")
		}
		if config.GetOutputFormat() == config.OutputFormatText {
			return printHealthText(node, status)
		}
		return output.Print(map[string]interface{}{"node": node, "health": health})
	},
}

var osdCmd = &cobra.Command{
	Use:           "osd",
	Short:         "List Ceph OSD topology and state",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		node, osds, err := withCephNode(client, client.GetCephOSDs)
		if err != nil {
			return err
		}
		if config.GetOutputFormat() == config.OutputFormatText {
			return printOSDsText(node, osds)
		}
		return output.Print(map[string]interface{}{"node": node, "osds": osds})
	},
}

var poolsCmd = &cobra.Command{
	Use:           "pools",
	Short:         "List Ceph pools",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		node, pools, err := withCephNode(client, client.GetCephPools)
		if err != nil {
			return err
		}
		if config.GetOutputFormat() == config.OutputFormatText {
			return printPoolsText(node, pools)
		}
		return output.Print(map[string]interface{}{"node": node, "pools": pools})
	},
}

// Init registers Ceph commands with the root command.
func Init(root *cobra.Command) {
	cephCmd.PersistentFlags().StringVar(&cephNode, "node", "", "Proxmox node to query (auto-detected by default)")
	cephCmd.AddCommand(statusCmd, healthCmd, osdCmd, poolsCmd)
	root.AddCommand(cephCmd)
}
