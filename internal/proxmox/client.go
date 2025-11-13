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

package proxmox

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

type Client struct {
	BaseURL  string
	TokenID  string
	Secret   string
	Insecure bool
	Debug    bool
	client   *http.Client
}

// Migrate migrates a VM or container to another node and returns the task ID
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
	if err != nil {
		return "", err
	}

	return out.Data, nil
}

func NewClient(cfg *config.Config) *Client {
	// Get current cluster configuration
	cluster, err := config.GetCurrentCluster()
	if err != nil {
		// Fallback to legacy config format for backward compatibility
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

		// Try to parse error response
		var errResp struct {
			Data    interface{} `json:"data"`
			Message string      `json:"message"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			// Handle specific error messages
			switch {
			case strings.Contains(errResp.Message, "already running"):
				parts := strings.Split(path, "/")
				id := parts[len(parts)-2]
				// Return just the message without error wrapping
				return &FriendlyError{Msg: fmt.Sprintf("instance %s is already running", id)}
			case strings.Contains(errResp.Message, "not running"):
				parts := strings.Split(path, "/")
				id := parts[len(parts)-2]
				return &FriendlyError{Msg: fmt.Sprintf("instance %s is not running", id)}
			default:
				return fmt.Errorf(errResp.Message)
			}
		}

		// Create a new reader with the same content for json.Decoder
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

	// Encode form data
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

// FriendlyError is an error that should be displayed without the 'Error:' prefix
type FriendlyError struct {
	Msg string
}

func (e *FriendlyError) Error() string {
	return e.Msg
}

type Node struct {
	Node string `json:"node"`
	IP   string `json:"ip"`
}
type nodeList struct {
	Data []Node `json:"data"`
}

type Instance struct {
	VMID       int    `json:"vmid"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	MaxMem     int64  `json:"maxmem"`
	MaxDisk    int64  `json:"maxdisk"`
	CPUs       int    `json:"cpus"`
	Type       string `json:"-"` // vm or ct
	Template   bool   `json:"-"` // is this a template?
	CloudInit  bool   `json:"-"` // has cloud-init configured?
	GuestAgent bool   `json:"-"` // has guest agent enabled?
}

// UnmarshalJSON implements custom unmarshaling to handle vmid as both string and int
func (i *Instance) UnmarshalJSON(data []byte) error {
	// Create a temporary struct with vmid as interface{} to handle both types
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
	
	// Handle vmid as either string or int
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

type VM = Instance // for backward compatibility
type instanceList struct {
	Data []Instance `json:"data"`
}

func (c *Client) GetNodes() ([]Node, error) {
	var out nodeList
	err := c.doRequest("GET", "/nodes", &out)
	return out.Data, err
}

// GetNodeIP returns the IP address of a specific node
func (c *Client) GetNodeIP(nodeName string) (string, error) {
	nodes, err := c.GetNodes()
	if err != nil {
		return "", err
	}
	for _, n := range nodes {
		if n.Node == nodeName {
			return n.IP, nil
		}
	}
	return "", fmt.Errorf("node %s not found", nodeName)
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

// GetSubscriptionStatus returns subscription status for a node
func (c *Client) GetSubscriptionStatus(node string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/subscription", node), &out)
	return out.Data, err
}

// GetClusterStatus returns cluster status information
func (c *Client) GetClusterStatus() ([]interface{}, error) {
	type wrap struct {
		Data []interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", "/cluster/status", &out)
	return out.Data, err
}

// GetClusterResources returns cluster resources (storage, nodes, VMs, etc.)
func (c *Client) GetClusterResources() ([]interface{}, error) {
	type wrap struct {
		Data []interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", "/cluster/resources", &out)
	return out.Data, err
}

// RunNodeCommand executes a shell command on a node
func (c *Client) RunNodeCommand(node, command string) error {
	data := map[string]string{
		"command": command,
	}

	var result map[string]interface{}
	err := c.doRequestWithData("POST", fmt.Sprintf("/nodes/%s/execute", node), data, &result)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetVMs(node string) ([]Instance, error) {
	var out instanceList
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/qemu", node), &out)
	if err != nil {
		return nil, err
	}

	// Get config for each VM to check if it's a template and has cloud-init
	for i := range out.Data {
		out.Data[i].Type = "vm"
		config, err := c.GetVMConfig(node, fmt.Sprintf("%d", out.Data[i].VMID))
		if err == nil {
			// Check template flag
			if template, ok := config["template"]; ok {
				if tplVal, ok := template.(float64); ok && tplVal == 1 {
					out.Data[i].Template = true
				}
			}

			// Check for cloud-init drive
			for key, value := range config {
				if strVal, ok := value.(string); ok {
					if (strings.HasPrefix(key, "ide") || strings.HasPrefix(key, "scsi")) && strings.Contains(strVal, "cloudinit") {
						out.Data[i].CloudInit = true
						break
					}
				}
			}

			// Check for guest agent
			if agent, ok := config["agent"]; ok {
				if agentStr, ok := agent.(string); ok {
					// Agent can be "1" or "enabled=1" or similar
					if agentStr == "1" || strings.Contains(agentStr, "enabled=1") {
						out.Data[i].GuestAgent = true
					}
				} else if agentNum, ok := agent.(float64); ok {
					if agentNum == 1 {
						out.Data[i].GuestAgent = true
					}
				}
			}

			// Get CPU cores
			if cores, ok := config["cores"]; ok {
				if coresVal, ok := cores.(float64); ok {
					out.Data[i].CPUs = int(coresVal)
				}
			}
		}
	}
	return out.Data, nil
}

func (c *Client) GetLXC(node string) ([]Instance, error) {
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

func (c *Client) GetAllInstances(node string) ([]Instance, error) {
	vms, err := c.GetVMs(node)
	if err != nil {
		return nil, fmt.Errorf("fetch VMs: %w", err)
	}

	cts, err := c.GetLXC(node)
	if err != nil {
		return nil, fmt.Errorf("fetch LXC: %w", err)
	}

	return append(vms, cts...), nil
}

// FindVMIDOnNode searches for a VMID on a specific node
func (c *Client) FindVMIDOnNode(node, vmid string) (instanceType string, err error) {
	// Try VMs first
	vms, err := c.GetVMs(node)
	if err == nil {
		for _, vm := range vms {
			if fmt.Sprintf("%d", vm.VMID) == vmid {
				return "vm", nil
			}
		}
	}

	// Then try LXC containers
	cts, err := c.GetLXC(node)
	if err == nil {
		for _, ct := range cts {
			if fmt.Sprintf("%d", ct.VMID) == vmid {
				return "ct", nil
			}
		}
	}

	return "", fmt.Errorf("Instance %s not found on node %s", vmid, node)
}

func (c *Client) FindNodeByVMID(vmid string) (node string, instanceType string, err error) {
	nodes, err := c.GetNodes()
	if err != nil {
		return "", "", err
	}
	for _, n := range nodes {
		// Try VMs first
		vms, err := c.GetVMs(n.Node)
		if err == nil {
			for _, vm := range vms {
				if fmt.Sprintf("%d", vm.VMID) == vmid {
					return n.Node, "vm", nil
				}
			}
		}

		// Then try LXC containers
		cts, err := c.GetLXC(n.Node)
		if err == nil {
			for _, ct := range cts {
				if fmt.Sprintf("%d", ct.VMID) == vmid {
					return n.Node, "ct", nil
				}
			}
		}
	}
	return "", "", fmt.Errorf("Instance %s not found", vmid)
}

func (c *Client) VMAction(node, vmid, action string) error {
	// First find out if this is a VM or CT
	_, instanceType, err := c.FindNodeByVMID(vmid)
	if err != nil {
		return err
	}

	// Use the correct endpoint based on instance type
	var endpoint string
	if instanceType == "vm" {
		endpoint = fmt.Sprintf("/nodes/%s/qemu/%s/status/%s", node, vmid, action)
	} else {
		endpoint = fmt.Sprintf("/nodes/%s/lxc/%s/status/%s", node, vmid, action)
	}

	err = c.doRequest("POST", endpoint, nil)
	if err != nil {
		// Check if it's already in the desired state
		if strings.Contains(err.Error(), "already running") && action == "start" {
			return &FriendlyError{Msg: fmt.Sprintf("instance %s is already running", vmid)}
		}
		if strings.Contains(err.Error(), "not running") && action == "stop" {
			return &FriendlyError{Msg: fmt.Sprintf("instance %s is not running", vmid)}
		}
		return err
	}

	// Return success message
	switch action {
	case "start":
		return &FriendlyError{Msg: fmt.Sprintf("instance %s started successfully", vmid)}
	case "stop":
		return &FriendlyError{Msg: fmt.Sprintf("instance %s stopped successfully", vmid)}
	default:
		return nil
	}
}

// GetInstanceStatus returns detailed status information for a VM or CT
func (c *Client) GetInstanceStatus(node, vmid string, instanceType string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap

	var endpoint string
	if instanceType == "vm" {
		endpoint = fmt.Sprintf("/nodes/%s/qemu/%s/status/current", node, vmid)
	} else {
		endpoint = fmt.Sprintf("/nodes/%s/lxc/%s/status/current", node, vmid)
	}

	err := c.doRequest("GET", endpoint, &out)
	return out.Data, err
}

func (c *Client) GetInstanceConfig(node, vmid string, instanceType string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap

	var endpoint string
	if instanceType == "vm" {
		endpoint = fmt.Sprintf("/nodes/%s/qemu/%s/config", node, vmid)
	} else {
		endpoint = fmt.Sprintf("/nodes/%s/lxc/%s/config", node, vmid)
	}

	err := c.doRequest("GET", endpoint, &out)
	return out.Data, err
}

func (c *Client) GetVMConfig(node, vmid string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/qemu/%s/config", node, vmid), &out)
	return out.Data, err
}

// TermProxyResponse represents the response from termproxy endpoint
type TermProxyResponse struct {
	UPID   string `json:"upid"`
	Port   int    `json:"-"` // We'll parse this manually
	Ticket string `json:"ticket"`
	User   string `json:"user"`
}

// CreateTermProxy creates a terminal proxy session for a VM or container
func (c *Client) CreateTermProxy(node, vmid, instanceType string) (*TermProxyResponse, error) {
	type rawResponse struct {
		UPID   string      `json:"upid"`
		Port   interface{} `json:"port"` // Can be string or int
		Ticket string      `json:"ticket"`
		User   string      `json:"user"`
	}
	type wrap struct {
		Data rawResponse `json:"data"`
	}
	var out wrap

	var endpoint string
	if instanceType == "vm" {
		endpoint = fmt.Sprintf("/nodes/%s/qemu/%s/termproxy", node, vmid)
	} else {
		endpoint = fmt.Sprintf("/nodes/%s/lxc/%s/termproxy", node, vmid)
	}

	err := c.doRequest("POST", endpoint, &out)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		fmt.Printf("DEBUG: Raw termproxy response - Port type: %T, Port value: %v\n", out.Data.Port, out.Data.Port)
	}

	// Parse port (can be string or int)
	var port int
	switch v := out.Data.Port.(type) {
	case float64:
		port = int(v)
	case string:
		port, err = strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("parse port: %w", err)
		}
	default:
		return nil, fmt.Errorf("unexpected port type: %T", v)
	}

	return &TermProxyResponse{
		UPID:   out.Data.UPID,
		Port:   port,
		Ticket: out.Data.Ticket,
		User:   out.Data.User,
	}, nil
}

// VMStatus represents VM/container current status with raw data
type VMStatus struct {
	Name    string
	Status  string
	RawData map[string]interface{}
}

// GetVMStatus gets current status and statistics for a VM or container
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

	// Extract name
	if name, ok := out.Data["name"].(string); ok {
		status.Name = name
	}

	// Extract status
	if st, ok := out.Data["status"].(string); ok {
		status.Status = st
	}

	return status, nil
}

// GetContainers gets list of LXC containers on a node
func (c *Client) GetContainers(node string) ([]Instance, error) {
	var out instanceList
	err := c.doRequest("GET", fmt.Sprintf("/nodes/%s/lxc", node), &out)
	if err != nil {
		return nil, err
	}

	// Mark all as containers
	for i := range out.Data {
		out.Data[i].Type = "ct"
	}

	return out.Data, nil
}

// FindVM finds which node a VM is on and returns node name and VM type
func (c *Client) FindVM(vmid int) (string, string, error) {
	nodes, err := c.GetNodes()
	if err != nil {
		return "", "", err
	}

	for _, node := range nodes {
		// Check VMs
		vms, err := c.GetVMs(node.Node)
		if err != nil {
			continue
		}
		for _, vm := range vms {
			if vm.VMID == vmid {
				return node.Node, "qemu", nil
			}
		}

		// Check containers
		containers, err := c.GetContainers(node.Node)
		if err != nil {
			continue
		}
		for _, ct := range containers {
			if ct.VMID == vmid {
				return node.Node, "lxc", nil
			}
		}
	}

	return "", "", fmt.Errorf("VM %d not found", vmid)
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
	if params.Target != "" {
		data["target"] = params.Target
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
	if err != nil {
		return "", err
	}

	return out.Data, nil
}

// WaitForTask waits for a Proxmox task to complete
func (c *Client) WaitForTask(node, taskID string, timeoutSeconds int) error {
	endpoint := fmt.Sprintf("/nodes/%s/tasks/%s/status", node, taskID)

	for i := 0; i < timeoutSeconds; i++ {
		var out struct {
			Data struct {
				Status   string `json:"status"`
				ExitCode string `json:"exitstatus"`
			} `json:"data"`
		}

		err := c.doRequest("GET", endpoint, &out)
		if err != nil {
			return err
		}

		if out.Data.Status == "stopped" {
			if out.Data.ExitCode == "OK" {
				return nil
			}
			return fmt.Errorf("task failed with exit code: %s", out.Data.ExitCode)
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("task timeout after %d seconds", timeoutSeconds)
}

// GetStorages returns storage information for a node
func (c *Client) GetStorages(node string) ([]interface{}, error) {
	type wrap struct {
		Data []interface{} `json:"data"`
	}
	var out wrap

	endpoint := fmt.Sprintf("/nodes/%s/storage", node)
	err := c.doRequest("GET", endpoint, &out)
	if err != nil {
		return nil, err
	}

	return out.Data, nil
}

// GetSnapshots returns list of snapshots for a VM or container
func (c *Client) GetSnapshots(node, vmid, instanceType string) ([]map[string]interface{}, error) {
	// Map instance type to API endpoint type
	var apiType string
	switch instanceType {
	case "vm":
		apiType = "qemu"
	case "ct":
		apiType = "lxc"
	case "qemu", "lxc":
		apiType = instanceType
	default:
		return nil, fmt.Errorf("unsupported instance type: %s", instanceType)
	}

	type wrap struct {
		Data []map[string]interface{} `json:"data"`
	}
	var out wrap

	endpoint := fmt.Sprintf("/nodes/%s/%s/%s/snapshot", node, apiType, vmid)
	err := c.doRequest("GET", endpoint, &out)
	if err != nil {
		return nil, err
	}

	return out.Data, nil
}

// DeleteSnapshot deletes a snapshot of a VM or container
func (c *Client) DeleteSnapshot(node, vmid, instanceType, snapname string) (string, error) {
	// Check if snapshot exists
	snapshots, err := c.GetSnapshots(node, vmid, instanceType)
	if err != nil {
		return "", fmt.Errorf("failed to check existing snapshots: %w", err)
	}

	// Verify snapshot exists
	snapshotExists := false
	for _, snap := range snapshots {
		if name, ok := snap["name"].(string); ok && name == snapname {
			snapshotExists = true
			break
		}
	}

	if !snapshotExists {
		return "", fmt.Errorf("snapshot '%s' not found", snapname)
	}

	// Map instance type to API endpoint type
	var apiType string
	switch instanceType {
	case "vm":
		apiType = "qemu"
	case "ct":
		apiType = "lxc"
	case "qemu", "lxc":
		apiType = instanceType
	default:
		return "", fmt.Errorf("unsupported instance type: %s", instanceType)
	}

	type wrap struct {
		Data string `json:"data"`
	}
	var out wrap

	endpoint := fmt.Sprintf("/nodes/%s/%s/%s/snapshot/%s", node, apiType, vmid, snapname)
	err = c.doRequest("DELETE", endpoint, &out)
	if err != nil {
		return "", err
	}

	// Return the task ID (UPID)
	return out.Data, nil
}

// CreateSnapshot creates a snapshot of a VM or container
func (c *Client) CreateSnapshot(node, vmid, instanceType, snapname, description string, includeRAM bool) (string, error) {
	// Check if snapshot name already exists
	snapshots, err := c.GetSnapshots(node, vmid, instanceType)
	if err != nil {
		return "", fmt.Errorf("failed to check existing snapshots: %w", err)
	}

	for _, snap := range snapshots {
		if name, ok := snap["name"].(string); ok && name == snapname {
			return "", fmt.Errorf("snapshot name '%s' already exists", snapname)
		}
	}

	// Build snapshot parameters
	data := map[string]string{
		"snapname": snapname,
	}
	
	if description != "" {
		data["description"] = description
	}
	
	// Include RAM state (vmstate) - ONLY for QEMU VMs, NOT for LXC containers
	// LXC containers don't support vmstate parameter
	if includeRAM && (instanceType == "vm" || instanceType == "qemu") {
		data["vmstate"] = "1"
	}

	type wrap struct {
		Data string `json:"data"`
	}
	var out wrap

	// Map instance type to API endpoint type
	// FindNodeByVMID returns "vm" or "ct", but API expects "qemu" or "lxc"
	var apiType string
	switch instanceType {
	case "vm":
		apiType = "qemu"
	case "ct":
		apiType = "lxc"
	case "qemu", "lxc":
		apiType = instanceType
	default:
		return "", fmt.Errorf("unsupported instance type: %s", instanceType)
	}

	// Use appropriate endpoint based on instance type
	endpoint := fmt.Sprintf("/nodes/%s/%s/%s/snapshot", node, apiType, vmid)
	
	// Use doRequestWithData for POST with form data
	err = c.doRequestWithData("POST", endpoint, data, &out)
	if err != nil {
		return "", err
	}

	// Return the task ID (UPID)
	return out.Data, nil
}

// CreateBackup creates a backup/snapshot to PBS storage
func (c *Client) CreateBackup(node, vmid, instanceType, storage string) (string, error) {
	// Build backup parameters with notes template
	data := map[string]string{
		"vmid":           vmid,
		"storage":        storage,
		"mode":           "snapshot", // Use snapshot mode for minimal downtime
		"remove":         "0",        // Don't remove old backups automatically
		"notes-template": "Backup of {{guestname}} (VMID: {{vmid}}) from node {{node}} in cluster {{cluster}}",
	}

	type wrap struct {
		Data string `json:"data"`
	}
	var out wrap

	endpoint := fmt.Sprintf("/nodes/%s/vzdump", node)
	
	// Use doRequestWithData for POST with form data
	err := c.doRequestWithData("POST", endpoint, data, &out)
	if err != nil {
		return "", err
	}

	// Return the task ID (UPID)
	return out.Data, nil
}

// GetTasks returns recent tasks from a node
func (c *Client) GetTasks(node string, limit int) ([]map[string]interface{}, error) {
	type wrap struct {
		Data []map[string]interface{} `json:"data"`
	}
	var out wrap

	endpoint := fmt.Sprintf("/nodes/%s/tasks", node)
	if limit > 0 {
		endpoint = fmt.Sprintf("%s?limit=%d", endpoint, limit)
	}
	
	err := c.doRequest("GET", endpoint, &out)
	if err != nil {
		return nil, err
	}

	return out.Data, nil
}

// GetTaskStatus returns the status of a task
func (c *Client) GetTaskStatus(node, upid string) (map[string]interface{}, error) {
	type wrap struct {
		Data map[string]interface{} `json:"data"`
	}
	var out wrap

	endpoint := fmt.Sprintf("/nodes/%s/tasks/%s/status", node, upid)
	err := c.doRequest("GET", endpoint, &out)
	if err != nil {
		return nil, err
	}

	return out.Data, nil
}

// GetTaskLog returns the log of a task
func (c *Client) GetTaskLog(node, upid string, lines int) ([]string, error) {
	type logEntry struct {
		N int    `json:"n"`
		T string `json:"t"`
	}
	type wrap struct {
		Data []logEntry `json:"data"`
	}
	var out wrap

	endpoint := fmt.Sprintf("/nodes/%s/tasks/%s/log", node, upid)
	if lines > 0 {
		endpoint = fmt.Sprintf("%s?limit=%d", endpoint, lines)
	}
	
	err := c.doRequest("GET", endpoint, &out)
	if err != nil {
		return nil, err
	}

	// Extract log text
	var log []string
	for _, entry := range out.Data {
		log = append(log, entry.T)
	}

	return log, nil
}

// RebootInstance reboots a VM or container
func (c *Client) RebootInstance(node, vmid, instanceType string) error {
	var endpoint string
	if instanceType == "vm" {
		endpoint = fmt.Sprintf("/nodes/%s/qemu/%s/status/reboot", node, vmid)
	} else {
		endpoint = fmt.Sprintf("/nodes/%s/lxc/%s/status/reboot", node, vmid)
	}
	return c.doRequestWithData("POST", endpoint, map[string]string{}, nil)
}

// ShutdownInstance shuts down a VM or container
func (c *Client) ShutdownInstance(node, vmid, instanceType string, force bool) error {
	var endpoint string
	if instanceType == "vm" {
		endpoint = fmt.Sprintf("/nodes/%s/qemu/%s/status/shutdown", node, vmid)
	} else {
		endpoint = fmt.Sprintf("/nodes/%s/lxc/%s/status/shutdown", node, vmid)
	}
	
	data := map[string]string{}
	if force {
		data["forceStop"] = "1"
	}
	
	return c.doRequestWithData("POST", endpoint, data, nil)
}

// ListBackups lists backups on a storage
func (c *Client) ListBackups(node, storage, vmid string) ([]map[string]interface{}, error) {
	type wrap struct {
		Data []map[string]interface{} `json:"data"`
	}
	var out wrap

	endpoint := fmt.Sprintf("/nodes/%s/storage/%s/content", node, storage)
	err := c.doRequest("GET", endpoint, &out)
	if err != nil {
		return nil, err
	}

	// Filter backups
	var backups []map[string]interface{}
	for _, item := range out.Data {
		if content, ok := item["content"].(string); ok && content == "backup" {
			if vmid != "" {
				if itemVMID, ok := item["vmid"].(float64); ok && fmt.Sprintf("%.0f", itemVMID) == vmid {
					backups = append(backups, item)
				}
			} else {
				backups = append(backups, item)
			}
		}
	}

	return backups, nil
}

// DeleteBackup deletes a backup
func (c *Client) DeleteBackup(node, storage, volid string) error {
	endpoint := fmt.Sprintf("/nodes/%s/storage/%s/content/%s", node, storage, volid)
	return c.doRequest("DELETE", endpoint, nil)
}

// ResizeDisk resizes a VM disk
func (c *Client) ResizeDisk(node, vmid, disk, size string) error {
	endpoint := fmt.Sprintf("/nodes/%s/qemu/%s/resize", node, vmid)
	data := map[string]string{
		"disk": disk,
		"size": size,
	}
	return c.doRequestWithData("PUT", endpoint, data, nil)
}

// CreateTemplate converts a VM to a template
func (c *Client) CreateTemplate(node, vmid string) error {
	endpoint := fmt.Sprintf("/nodes/%s/qemu/%s/template", node, vmid)
	return c.doRequestWithData("POST", endpoint, map[string]string{}, nil)
}

// SuspendVM suspends a VM
func (c *Client) SuspendVM(node, vmid string) error {
	endpoint := fmt.Sprintf("/nodes/%s/qemu/%s/status/suspend", node, vmid)
	return c.doRequestWithData("POST", endpoint, map[string]string{}, nil)
}

// ResumeVM resumes a suspended VM
func (c *Client) ResumeVM(node, vmid string) error {
	endpoint := fmt.Sprintf("/nodes/%s/qemu/%s/status/resume", node, vmid)
	return c.doRequestWithData("POST", endpoint, map[string]string{}, nil)
}

// StartVM starts a VM or container
func (c *Client) StartVM(node string, vmid int, vmType string) error {
	endpoint := fmt.Sprintf("/nodes/%s/%s/%d/status/start", node, vmType, vmid)
	return c.doRequestWithData("POST", endpoint, map[string]string{}, nil)
}
