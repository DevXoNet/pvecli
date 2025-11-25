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
	"time"
)

// AgentPing pings the QEMU guest agent
func (c *Client) AgentPing(node, vmid string) error {
	return c.doRequest("POST", fmt.Sprintf("/nodes/%s/qemu/%s/agent/ping", node, vmid), nil)
}

// AgentGetNetworkInterfaces gets network interfaces from guest agent
func (c *Client) AgentGetNetworkInterfaces(node, vmid string) ([]GuestAgentNetworkInterface, error) {
	type wrap struct {
		Data struct {
			Result []GuestAgentNetworkInterface `json:"result"`
		} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/qemu/%s/agent/network-get-interfaces", node, vmid), &out)
	return out.Data.Result, err
}

// AgentGetOSInfo gets OS information from guest agent
func (c *Client) AgentGetOSInfo(node, vmid string) (*GuestAgentOSInfo, error) {
	type wrap struct {
		Data struct {
			Result GuestAgentOSInfo `json:"result"`
		} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/qemu/%s/agent/get-osinfo", node, vmid), &out)
	return &out.Data.Result, err
}

// AgentExec executes a command in the guest via agent
func (c *Client) AgentExec(node, vmid string, command []string) (*GuestAgentExecResult, error) {
	commandStr := strings.Join(command, " ")
	data := map[string]string{
		"command": commandStr,
	}
	type wrap struct {
		Data struct {
			PID int `json:"pid"`
		} `json:"data"`
	}
	var out wrap
	err := c.doRequestWithData("POST", fmt.Sprintf("/nodes/%s/qemu/%s/agent/exec", node, vmid), data, &out)
	if err != nil {
		return nil, err
	}

	// Wait for command to complete
	for i := 0; i < 30; i++ {
		statusData := map[string]string{
			"pid": fmt.Sprintf("%d", out.Data.PID),
		}
		type statusWrap struct {
			Data struct {
				Result GuestAgentExecResult `json:"result"`
			} `json:"data"`
		}
		var statusResponse statusWrap
		err = c.doRequestWithData("GET", fmt.Sprintf("/nodes/%s/qemu/%s/agent/exec-status", node, vmid), statusData, &statusResponse)
		if err != nil {
			return nil, err
		}
		if statusResponse.Data.Result.Exited {
			return &statusResponse.Data.Result, nil
		}
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("command execution timeout")
}

// AgentGetFSInfo gets filesystem information from guest agent
func (c *Client) AgentGetFSInfo(node, vmid string) ([]GuestAgentFSInfo, error) {
	type wrap struct {
		Data struct {
			Result []GuestAgentFSInfo `json:"result"`
		} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/qemu/%s/agent/get-fsinfo", node, vmid), &out)
	return out.Data.Result, err
}
