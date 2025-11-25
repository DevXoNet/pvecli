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
	"strings"
)

// GetLXC returns list of LXC containers on a node
func (c *Client) GetLXC(node string) ([]Instance, error) {
	type instanceList struct {
		Data []Instance `json:"data"`
	}
	var out instanceList
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/lxc", node), &out)
	if err != nil {
		return nil, err
	}
	for i := range out.Data {
		out.Data[i].Type = "ct"
	}
	return out.Data, nil
}

// GetContainers is an alias for GetLXC
func (c *Client) GetContainers(node string) ([]Instance, error) {
	return c.GetLXC(node)
}

// GetAllInstances returns all VMs and containers on a node
func (c *Client) GetAllInstances(node string) ([]Instance, error) {
	vms, err := c.GetVMs(node)
	if err != nil {
		return nil, fmt.Errorf("fetch VMs: %w", err)
	}
	cts, err := c.GetLXC(node)
	if err != nil {
		return nil, fmt.Errorf("fetch containers: %w", err)
	}
	return append(vms, cts...), nil
}

// GetContainerIPFromConfig gets IP address from container config
func (c *Client) GetContainerIPFromConfig(node, vmid string) (string, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/lxc/%s/config", node, vmid), &out)
	if err != nil {
		return "", err
	}
	for key, value := range out.Data {
		if strings.HasPrefix(key, "net") {
			if netConfig, ok := value.(string); ok {
				parts := strings.Split(netConfig, ",")
				for _, part := range parts {
					if strings.HasPrefix(part, "ip=") {
						ipWithMask := strings.TrimPrefix(part, "ip=")
						ip := strings.Split(ipWithMask, "/")[0]
						if ip != "" && ip != "dhcp" {
							return ip, nil
						}
					}
				}
			}
		}
	}
	return "", fmt.Errorf("no IP address found in config")
}

// FindVM finds which node a VM is on and returns node name and VM type
func (c *Client) FindVM(vmid int) (string, string, error) {
	nodes, err := c.GetNodes()
	if err != nil {
		return "", "", err
	}
	for _, node := range nodes {
		vms, err := c.GetVMs(node.Node)
		if err == nil {
			for _, vm := range vms {
				if vm.VMID == vmid {
					return node.Node, "qemu", nil
				}
			}
		}
		containers, err := c.GetContainers(node.Node)
		if err == nil {
			for _, ct := range containers {
				if ct.VMID == vmid {
					return node.Node, "lxc", nil
				}
			}
		}
	}
	return "", "", fmt.Errorf("VM %d not found", vmid)
}
