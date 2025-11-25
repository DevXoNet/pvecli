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
	"strings"
)

// GetVMsLight returns list of VMs on a node without fetching config (faster)
func (c *Client) GetVMsLight(node string) ([]Instance, error) {
	type instanceList struct {
		Data []Instance `json:"data"`
	}
	var out instanceList
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/qemu", node), &out)
	if err != nil {
		return nil, err
	}
	for i := range out.Data {
		out.Data[i].Type = "vm"
	}
	return out.Data, nil
}

// GetVMs returns list of VMs on a node with full config details
func (c *Client) GetVMs(node string) ([]Instance, error) {
	type instanceList struct {
		Data []Instance `json:"data"`
	}
	var out instanceList
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/qemu", node), &out)
	if err != nil {
		return nil, err
	}
	for i := range out.Data {
		out.Data[i].Type = "vm"
		config, err := c.GetVMConfig(node, fmt.Sprintf("%d", out.Data[i].VMID))
		if err == nil {
			if template, ok := config["template"]; ok {
				if tplVal, ok := template.(float64); ok && tplVal == 1 {
					out.Data[i].Template = true
				}
			}
			for key, value := range config {
				if strVal, ok := value.(string); ok {
					if (strings.HasPrefix(key, "ide") || strings.HasPrefix(key, "scsi")) && strings.Contains(strVal, "cloudinit") {
						out.Data[i].CloudInit = true
						break
					}
				}
			}
			if agent, ok := config["agent"]; ok {
				if agentStr, ok := agent.(string); ok {
					if agentStr == "1" || strings.Contains(agentStr, "enabled=1") {
						out.Data[i].GuestAgent = true
					}
				} else if agentNum, ok := agent.(float64); ok {
					if agentNum == 1 {
						out.Data[i].GuestAgent = true
					}
				}
			}
		}
	}
	return out.Data, nil
}

// GetVMConfig returns configuration for a VM
func (c *Client) GetVMConfig(node, vmid string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap
	vmidInt, err := strconv.Atoi(vmid)
	if err != nil {
		return nil, fmt.Errorf("invalid vmid: %s", vmid)
	}
	err = c.doRequest("GET", fmt.Sprintf("/nodes/%s/qemu/%d/config", node, vmidInt), &out)
	return out.Data, err
}

// GetVMStatus gets current status and statistics for a VM
func (c *Client) GetVMStatus(node string, vmid int, vmType string) (*VMStatus, error) {
	var out struct {
		Data map[string]interface{} `json:"data"`
	}
	endpoint := fmt.Sprintf("/nodes/%s/%s/%d/status/current", node, vmType, vmid)
	err := c.doRequest("GET", endpoint, &out)
	if err != nil {
		return nil, err
	}
	status := &VMStatus{
		RawData: out.Data,
	}
	if name, ok := out.Data["name"].(string); ok {
		status.Name = name
	}
	if st, ok := out.Data["status"].(string); ok {
		status.Status = st
	}
	return status, nil
}

// StartVM starts a VM or container
func (c *Client) StartVM(node string, vmid int, vmType string) error {
	var out struct {
		Data string `json:"data"`
	}
	endpoint := fmt.Sprintf("/nodes/%s/%s/%d/status/start", node, vmType, vmid)
	return c.doRequestWithData("POST", endpoint, nil, &out)
}

// VMAction performs an action on a VM (start, stop, etc.)
func (c *Client) VMAction(node, vmid, action string) error {
	_, instanceType, err := c.FindNodeByVMID(vmid)
	if err != nil {
		return err
	}
	var out struct {
		Data string `json:"data"`
	}
	endpoint := fmt.Sprintf("/nodes/%s/%s/%s/status/%s", node, instanceType, vmid, action)
	return c.doRequestWithData("POST", endpoint, nil, &out)
}

// SuspendVM suspends a VM
func (c *Client) SuspendVM(node, vmid string) error {
	var out struct {
		Data string `json:"data"`
	}
	return c.doRequestWithData("POST", fmt.Sprintf("/nodes/%s/qemu/%s/status/suspend", node, vmid), nil, &out)
}

// ResumeVM resumes a suspended VM
func (c *Client) ResumeVM(node, vmid string) error {
	var out struct {
		Data string `json:"data"`
	}
	return c.doRequestWithData("POST", fmt.Sprintf("/nodes/%s/qemu/%s/status/resume", node, vmid), nil, &out)
}

// CreateTemplate converts a VM to a template
func (c *Client) CreateTemplate(node, vmid string) error {
	var out struct {
		Data string `json:"data"`
	}
	return c.doRequestWithData("POST", fmt.Sprintf("/nodes/%s/qemu/%s/template", node, vmid), nil, &out)
}

// ResizeDisk resizes a VM disk
func (c *Client) ResizeDisk(node, vmid, disk, size string) error {
	data := map[string]string{
		"disk": disk,
		"size": size,
	}
	return c.doRequestWithData("PUT", fmt.Sprintf("/nodes/%s/qemu/%s/resize", node, vmid), data, nil)
}

// CloneVM clones a VM or container
func (c *Client) CloneVM(node string, vmid int, vmType string, params CloneParams) (string, error) {
	data := map[string]string{
		"newid": fmt.Sprintf("%d", params.NewID),
	}
	if params.Name != "" {
		data["name"] = params.Name
	}
	if params.Description != "" {
		data["description"] = params.Description
	}
	if params.Pool != "" {
		data["pool"] = params.Pool
	}
	if params.Storage != "" {
		data["storage"] = params.Storage
	}
	if params.Full {
		data["full"] = "1"
	}
	var out struct {
		Data string `json:"data"`
	}
	endpoint := fmt.Sprintf("/nodes/%s/%s/%d/clone", node, vmType, vmid)
	err := c.doRequestWithData("POST", endpoint, data, &out)
	return out.Data, err
}

// Migrate migrates a VM or container to another node
func (c *Client) Migrate(node string, vmid int, vmType string, target string, online bool) (string, error) {
	data := map[string]string{
		"target": target,
	}
	if online {
		data["online"] = "1"
	}
	var out struct {
		Data string `json:"data"`
	}
	endpoint := fmt.Sprintf("/nodes/%s/%s/%d/migrate", node, vmType, vmid)
	err := c.doRequestWithData("POST", endpoint, data, &out)
	return out.Data, err
}

// GetVMAgentNetworkInterfaces gets network interfaces from VM guest agent
func (c *Client) GetVMAgentNetworkInterfaces(node, vmid string) ([]map[string]interface{}, error) {
	type wrap struct {
		Data struct {
			Result []map[string]interface{} `json:"result"`
		} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/qemu/%s/agent/network-get-interfaces", node, vmid), &out)
	return out.Data.Result, err
}

// GetVMIPFromAgent gets IP address from VM guest agent
func (c *Client) GetVMIPFromAgent(node, vmid string) (string, error) {
	interfaces, err := c.GetVMAgentNetworkInterfaces(node, vmid)
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		if name, ok := iface["name"].(string); ok && name != "lo" {
			if ipAddresses, ok := iface["ip-addresses"].([]interface{}); ok {
				for _, ipAddr := range ipAddresses {
					if ipMap, ok := ipAddr.(map[string]interface{}); ok {
						if ipType, ok := ipMap["ip-address-type"].(string); ok && ipType == "ipv4" {
							if ip, ok := ipMap["ip-address"].(string); ok {
								return ip, nil
							}
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("no IP address found")
}
