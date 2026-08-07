// Copyright 2025 DevXo part of vByte Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package pve

import "fmt"

// GetCephStatus returns the Ceph cluster status reported by a Proxmox node.
func (c *Client) GetCephStatus(node string) (map[string]interface{}, error) {
	var out struct {
		Data map[string]interface{} `json:"data"`
	}
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/ceph/status", node), &out)
	return out.Data, err
}

// GetCephOSDs returns the Ceph OSD tree reported by a Proxmox node.
func (c *Client) GetCephOSDs(node string) (map[string]interface{}, error) {
	var out struct {
		Data map[string]interface{} `json:"data"`
	}
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/ceph/osd", node), &out)
	return out.Data, err
}

// GetCephPools returns the Ceph pools reported by a Proxmox node.
func (c *Client) GetCephPools(node string) ([]map[string]interface{}, error) {
	var out struct {
		Data []map[string]interface{} `json:"data"`
	}
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/ceph/pool", node), &out)
	return out.Data, err
}
