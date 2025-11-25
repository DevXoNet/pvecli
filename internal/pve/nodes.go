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

package pve

import (
	"fmt"
	"strconv"
	"time"
)

// GetNodes returns list of all nodes in the cluster
func (c *Client) GetNodes() ([]Node, error) {
	type nodeList struct {
		Data []Node `json:"data"`
	}
	var out nodeList
	err := c.doRequest("GET", "/nodes", &out)
	if err != nil {
		return nil, err
	}

	// Get IP addresses from cluster status
	clusterStatus, err := c.GetClusterStatus()
	if err == nil && clusterStatus != nil {
		nodeIPMap := make(map[string]string)
		for _, item := range clusterStatus {
			if itemMap, ok := item.(map[string]interface{}); ok {
				if itemMap["type"] == "node" {
					nodeName := ""
					nodeIP := ""
					if itemMap["name"] != nil {
						nodeName = itemMap["name"].(string)
					}
					if itemMap["ip"] != nil {
						nodeIP = itemMap["ip"].(string)
					}
					if nodeName != "" && nodeIP != "" {
						nodeIPMap[nodeName] = nodeIP
					}
				}
			}
		}

		for i := range out.Data {
			if ip, ok := nodeIPMap[out.Data[i].Node]; ok {
				out.Data[i].IP = ip
			}
		}
	}

	return out.Data, nil
}

// GetNodeStatus returns status information for a specific node
func (c *Client) GetNodeStatus(node string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/status", node), &out)
	return out.Data, err
}

// GetStorages returns storage information for a node
func (c *Client) GetStorages(node string) ([]interface{}, error) {
	type wrap struct {
		Data []interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/storage", node), &out)
	return out.Data, err
}

// GetTasks returns list of tasks for a node
func (c *Client) GetTasks(node string, limit int) ([]map[string]interface{}, error) {
	type wrap struct {
		Data []map[string]interface{} `json:"data"`
	}
	var out wrap
	path := fmt.Sprintf("/nodes/%s/tasks", node)
	if limit > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, limit)
	}
	err := c.doRequest("GET", path, &out)
	return out.Data, err
}

// GetTaskStatus returns status of a specific task
func (c *Client) GetTaskStatus(node, upid string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/tasks/%s/status", node, upid), &out)
	return out.Data, err
}

// GetTaskLog returns log lines for a specific task
func (c *Client) GetTaskLog(node, upid string, lines int) ([]string, error) {
	type logEntry struct {
		N int    `json:"n"`
		T string `json:"t"`
	}
	type wrap struct {
		Data []logEntry `json:"data"`
	}
	var out wrap
	path := fmt.Sprintf("/nodes/%s/tasks/%s/log", node, upid)
	if lines > 0 {
		path = fmt.Sprintf("%s?limit=%d", path, lines)
	}
	err := c.doRequest("GET", path, &out)
	if err != nil {
		return nil, err
	}
	var logLines []string
	for _, entry := range out.Data {
		logLines = append(logLines, entry.T)
	}
	return logLines, nil
}

// WaitForTask waits for a Proxmox task to complete
func (c *Client) WaitForTask(node, taskID string, timeoutSeconds int) error {
	endpoint := fmt.Sprintf("/nodes/%s/tasks/%s/status", node, taskID)
	for i := 0; i < timeoutSeconds; i++ {
		var out struct {
			Data struct {
				Status     string `json:"status"`
				ExitStatus string `json:"exitstatus"`
			} `json:"data"`
		}
		err := c.doRequest("GET", endpoint, &out)
		if err != nil {
			return err
		}
		if out.Data.Status == "stopped" {
			if out.Data.ExitStatus == "OK" {
				return nil
			}
			return fmt.Errorf("task failed with status: %s", out.Data.ExitStatus)
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("task timeout after %d seconds", timeoutSeconds)
}

// ListBackups returns list of backups for a VM/CT on specified storage
func (c *Client) ListBackups(node, storage, vmid string) ([]map[string]interface{}, error) {
	type wrap struct {
		Data []map[string]interface{} `json:"data"`
	}
	var out wrap
	path := fmt.Sprintf("/nodes/%s/storage/%s/content", node, storage)
	if vmid != "" {
		path = fmt.Sprintf("%s?vmid=%s", path, vmid)
	}
	err := c.doRequest("GET", path, &out)
	return out.Data, err
}

// DeleteBackup deletes a backup from storage
func (c *Client) DeleteBackup(node, storage, volid string) error {
	return c.doRequest("DELETE", fmt.Sprintf("/nodes/%s/storage/%s/content/%s", node, storage, volid), nil)
}

// FindVMIDOnNode searches for a VMID on a specific node
func (c *Client) FindVMIDOnNode(node, vmid string) (instanceType string, err error) {
	vms, err := c.GetVMs(node)
	if err == nil {
		for _, vm := range vms {
			if strconv.Itoa(vm.VMID) == vmid {
				return "qemu", nil
			}
		}
	}
	cts, err := c.GetLXC(node)
	if err == nil {
		for _, ct := range cts {
			if strconv.Itoa(ct.VMID) == vmid {
				return "lxc", nil
			}
		}
	}
	return "", fmt.Errorf("Instance %s not found on node %s", vmid, node)
}

// FindNodeByVMID finds which node a VMID is on
func (c *Client) FindNodeByVMID(vmid string) (node string, instanceType string, err error) {
	nodes, err := c.GetNodes()
	if err != nil {
		return "", "", err
	}
	for _, n := range nodes {
		instanceType, err := c.FindVMIDOnNode(n.Node, vmid)
		if err == nil {
			return n.Node, instanceType, nil
		}
	}
	return "", "", fmt.Errorf("Instance %s not found", vmid)
}

// GetInstanceStatus returns detailed status information for a VM or CT
func (c *Client) GetInstanceStatus(node, vmid string, instanceType string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap
	var apiType string
	switch instanceType {
	case "vm", "qemu":
		apiType = "qemu"
	case "ct", "lxc":
		apiType = "lxc"
	default:
		apiType = instanceType
	}
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/%s/%s/status/current", node, apiType, vmid), &out)
	return out.Data, err
}

// GetInstanceConfig returns configuration for a VM or CT
func (c *Client) GetInstanceConfig(node, vmid string, instanceType string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap
	var apiType string
	switch instanceType {
	case "vm", "qemu":
		apiType = "qemu"
	case "ct", "lxc":
		apiType = "lxc"
	default:
		apiType = instanceType
	}
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/%s/%s/config", node, apiType, vmid), &out)
	return out.Data, err
}

// GetSnapshots returns list of snapshots for a VM or container
func (c *Client) GetSnapshots(node, vmid, instanceType string) ([]map[string]interface{}, error) {
	var apiType string
	switch instanceType {
	case "vm", "qemu":
		apiType = "qemu"
	case "ct", "lxc":
		apiType = "lxc"
	default:
		apiType = instanceType
	}
	type wrap struct {
		Data []map[string]interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/%s/%s/snapshot", node, apiType, vmid), &out)
	return out.Data, err
}

// DeleteSnapshot deletes a snapshot of a VM or container
func (c *Client) DeleteSnapshot(node, vmid, instanceType, snapname string) (string, error) {
	snapshots, err := c.GetSnapshots(node, vmid, instanceType)
	if err != nil {
		return "", err
	}
	found := false
	for _, snap := range snapshots {
		if name, ok := snap["name"].(string); ok && name == snapname {
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("snapshot '%s' not found", snapname)
	}
	var apiType string
	switch instanceType {
	case "vm", "qemu":
		apiType = "qemu"
	case "ct", "lxc":
		apiType = "lxc"
	default:
		apiType = instanceType
	}
	var out struct {
		Data string `json:"data"`
	}
	endpoint := fmt.Sprintf("/nodes/%s/%s/%s/snapshot/%s", node, apiType, vmid, snapname)
	err = c.doRequest("DELETE", endpoint, &out)
	return out.Data, err
}

// CreateSnapshot creates a snapshot of a VM or container
func (c *Client) CreateSnapshot(node, vmid, instanceType, snapname, description string, includeRAM bool) (string, error) {
	snapshots, err := c.GetSnapshots(node, vmid, instanceType)
	if err != nil {
		return "", err
	}
	for _, snap := range snapshots {
		if name, ok := snap["name"].(string); ok && name == snapname {
			return "", fmt.Errorf("snapshot '%s' already exists", snapname)
		}
	}
	var apiType string
	switch instanceType {
	case "vm", "qemu":
		apiType = "qemu"
	case "ct", "lxc":
		apiType = "lxc"
	default:
		apiType = instanceType
	}
	data := map[string]string{
		"snapname": snapname,
	}
	if description != "" {
		data["description"] = description
	}
	if apiType == "qemu" && includeRAM {
		data["vmstate"] = "1"
	}
	var out struct {
		Data string `json:"data"`
	}
	endpoint := fmt.Sprintf("/nodes/%s/%s/%s/snapshot", node, apiType, vmid)
	err = c.doRequestWithData("POST", endpoint, data, &out)
	return out.Data, err
}

// RebootInstance reboots a VM or container
func (c *Client) RebootInstance(node, vmid, instanceType string) error {
	var out struct {
		Data string `json:"data"`
	}
	return c.doRequestWithData("POST", fmt.Sprintf("/nodes/%s/%s/%s/status/reboot", node, instanceType, vmid), nil, &out)
}

// ShutdownInstance shuts down a VM or container gracefully
func (c *Client) ShutdownInstance(node, vmid, instanceType string, force bool) error {
	data := map[string]string{}
	if force {
		data["forceStop"] = "1"
	}
	var out struct {
		Data string `json:"data"`
	}
	return c.doRequestWithData("POST", fmt.Sprintf("/nodes/%s/%s/%s/status/shutdown", node, instanceType, vmid), data, &out)
}
