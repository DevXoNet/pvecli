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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pvecli/config"
)

// Client represents a Proxmox VE API client
type Client struct {
	BaseURL  string
	TokenID  string
	Secret   string
	Insecure bool
	Debug    bool
	client   *http.Client
}

// NewClient creates a new Proxmox API client
func NewClient(cfg *config.Config) *Client {
	cluster, err := config.GetCurrentCluster()
	if err != nil {
		if cfg.APIURL != "" && cfg.TokenID != "" && cfg.TokenSecret != "" {
			cluster = &config.ClusterConfig{
				APIURL:             cfg.APIURL,
				TokenID:            cfg.TokenID,
				TokenSecret:        cfg.TokenSecret,
				InsecureSkipVerify: cfg.InsecureSkipVerify,
			}
		} else {
			panic(fmt.Sprintf("failed to get cluster config: %v", err))
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: cluster.InsecureSkipVerify},
	}
	return &Client{
		BaseURL:  cluster.APIURL,
		TokenID:  cluster.TokenID,
		Secret:   cluster.TokenSecret,
		Insecure: cluster.InsecureSkipVerify,
		Debug:    config.Debug(),
		client:   &http.Client{Timeout: 20 * time.Second, Transport: tr},
	}
}

// doRequest performs HTTP request
func (c *Client) doRequest(method, path string, target interface{}) error {
	if c.Debug {
		fmt.Printf("DEBUG: %s %s%s\n", method, c.BaseURL, path)
	}
	url := fmt.Sprintf("%s%s", c.BaseURL, path)
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret))

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)

		if c.Debug {
			fmt.Printf("DEBUG: Response Status: %d\nDEBUG: Response Body: %s\n", resp.StatusCode, string(body))
		}

		var errResp struct {
			Data    interface{} `json:"data"`
			Message string      `json:"message"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			switch {
			case strings.Contains(errResp.Message, "already running"):
				parts := strings.Split(path, "/")
				id := parts[len(parts)-2]
				return &FriendlyError{Msg: fmt.Sprintf("instance %s is already running", id)}
			case strings.Contains(errResp.Message, "not running"):
				parts := strings.Split(path, "/")
				id := parts[len(parts)-2]
				return &FriendlyError{Msg: fmt.Sprintf("instance %s is not running", id)}
			default:
				return fmt.Errorf("%s", errResp.Message)
			}
		}

		resp.Body = io.NopCloser(bytes.NewBuffer(body))
		return fmt.Errorf("http %d for %s %s", resp.StatusCode, method, path)
	}

	if target != nil {
		return json.NewDecoder(resp.Body).Decode(target)
	}
	return nil
}

// doRequestWithData performs HTTP request with form data
func (c *Client) doRequestWithData(method, path string, data map[string]string, target interface{}) error {
	if c.Debug {
		fmt.Printf("DEBUG: %s %s%s with data: %v\n", method, c.BaseURL, path, data)
	}

	url := fmt.Sprintf("%s%s", c.BaseURL, path)

	formData := ""
	for k, v := range data {
		if formData != "" {
			formData += "&"
		}
		formData += fmt.Sprintf("%s=%s", k, v)
	}

	req, err := http.NewRequest(method, url, strings.NewReader(formData))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		if c.Debug {
			fmt.Printf("DEBUG: Response Status: %d\nDEBUG: Response Body: %s\n", resp.StatusCode, string(body))
		}
		return fmt.Errorf("http %d for %s %s: %s", resp.StatusCode, method, path, string(body))
	}

	if target != nil {
		return json.NewDecoder(resp.Body).Decode(target)
	}
	return nil
}

// doRequestWithJSON performs HTTP request with JSON data
func (c *Client) doRequestWithJSON(method, path string, data interface{}, target interface{}) error {
	if c.Debug {
		fmt.Printf("DEBUG: %s %s%s with JSON data: %+v\n", method, c.BaseURL, path, data)
	}

	url := fmt.Sprintf("%s%s", c.BaseURL, path)

	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("PVEAPIToken=%s=%s", c.TokenID, c.Secret))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		if c.Debug {
			fmt.Printf("DEBUG: Response Status: %d\nDEBUG: Response Body: %s\n", resp.StatusCode, string(body))
		}
		return fmt.Errorf("http %d for %s %s: %s", resp.StatusCode, method, path, string(body))
	}

	if target != nil {
		return json.NewDecoder(resp.Body).Decode(target)
	}
	return nil
}

// FriendlyError is an error that should be displayed without the 'Error:' prefix
type FriendlyError struct {
	Msg string
}

func (e *FriendlyError) Error() string {
	return e.Msg
}

// Node represents a Proxmox node
type Node struct {
	Node   string  `json:"node"`
	IP     string  `json:"ip"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"`
	Mem    uint64  `json:"mem"`
	MaxMem uint64  `json:"maxmem"`
	Uptime int64   `json:"uptime"`
}

// Instance represents a VM or container
type Instance struct {
	VMID       int    `json:"vmid"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	MaxMem     int64  `json:"maxmem"`
	MaxDisk    int64  `json:"maxdisk"`
	CPUs       int    `json:"cpus"`
	Type       string `json:"-"`
	Template   bool   `json:"-"`
	CloudInit  bool   `json:"-"`
	GuestAgent bool   `json:"-"`
}

// UnmarshalJSON implements custom unmarshaling to handle vmid as both string and int
func (i *Instance) UnmarshalJSON(data []byte) error {
	type Alias Instance
	aux := &struct {
		VMID interface{} `json:"vmid"`
		*Alias
	}{
		Alias: (*Alias)(i),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	switch v := aux.VMID.(type) {
	case string:
		vmid, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("invalid vmid string: %v", v)
		}
		i.VMID = vmid
	case float64:
		i.VMID = int(v)
	case int:
		i.VMID = v
	default:
		return fmt.Errorf("vmid has unexpected type: %T", v)
	}

	return nil
}

// VMStatus represents VM status information
type VMStatus struct {
	Name    string
	Status  string
	RawData map[string]interface{}
}

// TermProxyResponse represents terminal proxy response
type TermProxyResponse struct {
	UPID   string
	Port   int
	Ticket string
	User   string
}

// CloneParams represents parameters for cloning a VM
type CloneParams struct {
	NewID       int
	Name        string
	Description string
	Target      string
	Pool        string
	Storage     string
	Full        bool
}

// GuestAgentIPAddress represents an IP address from guest agent
type GuestAgentIPAddress struct {
	IPAddress     string `json:"ip-address"`
	IPAddressType string `json:"ip-address-type"`
	Prefix        int    `json:"prefix"`
}

// GuestAgentNetworkInterface represents a network interface from guest agent
type GuestAgentNetworkInterface struct {
	Name            string                `json:"name"`
	HardwareAddress string                `json:"hardware-address"`
	IPAddresses     []GuestAgentIPAddress `json:"ip-addresses"`
}

// GuestAgentOSInfo represents OS information from guest agent
type GuestAgentOSInfo struct {
	Name          string `json:"name"`
	PrettyName    string `json:"pretty-name"`
	Version       string `json:"version"`
	VersionID     string `json:"version-id"`
	KernelRelease string `json:"kernel-release"`
	KernelVersion string `json:"kernel-version"`
	Machine       string `json:"machine"`
}

// GuestAgentExecResult represents result of command execution
type GuestAgentExecResult struct {
	PID      int    `json:"pid"`
	OutData  string `json:"out-data"`
	ErrData  string `json:"err-data"`
	Exited   bool   `json:"exited"`
	ExitCode int    `json:"exitcode"`
}

// GuestAgentFSInfo represents filesystem information from guest agent
type GuestAgentFSInfo struct {
	Name       string `json:"name"`
	Mountpoint string `json:"mountpoint"`
	Type       string `json:"type"`
	TotalBytes int64  `json:"total-bytes"`
	UsedBytes  int64  `json:"used-bytes"`
}
