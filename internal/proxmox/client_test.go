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
	"testing"

	"pvecli/config"
)

func TestNewClient(t *testing.T) {
	cfg := &config.Config{
		APIURL:             "https://test.example.com:8006/api2/json",
		TokenID:            "test@pam!test",
		TokenSecret:        "test-secret",
		InsecureSkipVerify: true,
		Debug:              false,
	}

	// Initialize config to avoid nil pointer
	config.SetDebug(false)

	client := NewClient(cfg)

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.BaseURL != cfg.APIURL {
		t.Errorf("Expected BaseURL '%s', got '%s'", cfg.APIURL, client.BaseURL)
	}

	if client.TokenID != cfg.TokenID {
		t.Errorf("Expected TokenID '%s', got '%s'", cfg.TokenID, client.TokenID)
	}

	if client.Secret != cfg.TokenSecret {
		t.Errorf("Expected Secret '%s', got '%s'", cfg.TokenSecret, client.Secret)
	}

	if client.Insecure != cfg.InsecureSkipVerify {
		t.Errorf("Expected Insecure %v, got %v", cfg.InsecureSkipVerify, client.Insecure)
	}
}

func TestFriendlyError(t *testing.T) {
	err := &FriendlyError{Msg: "test error message"}

	if err.Error() != "test error message" {
		t.Errorf("Expected error message 'test error message', got '%s'", err.Error())
	}
}

func TestNodeStruct(t *testing.T) {
	node := Node{
		Node: "node01",
		IP:   "192.168.1.1",
	}

	if node.Node != "node01" {
		t.Errorf("Expected Node 'node01', got '%s'", node.Node)
	}

	if node.IP != "192.168.1.1" {
		t.Errorf("Expected IP '192.168.1.1', got '%s'", node.IP)
	}
}

func TestInstanceStruct(t *testing.T) {
	instance := Instance{
		VMID:      100,
		Name:      "test-vm",
		Status:    "running",
		MaxMem:    4096,
		MaxDisk:   32768,
		CPUs:      2,
		Type:      "vm",
		Template:  false,
		CloudInit: true,
	}

	if instance.VMID != 100 {
		t.Errorf("Expected VMID 100, got %d", instance.VMID)
	}

	if instance.Name != "test-vm" {
		t.Errorf("Expected Name 'test-vm', got '%s'", instance.Name)
	}

	if instance.Status != "running" {
		t.Errorf("Expected Status 'running', got '%s'", instance.Status)
	}

	if instance.Type != "vm" {
		t.Errorf("Expected Type 'vm', got '%s'", instance.Type)
	}

	if instance.Template {
		t.Error("Expected Template to be false")
	}

	if !instance.CloudInit {
		t.Error("Expected CloudInit to be true")
	}
}
