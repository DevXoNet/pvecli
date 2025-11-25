// Copyright 2025 DevXo part of vByte Ltd //
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package instance

import (
	"fmt"
	"pvecli/internal/pve"
)

// FindInstance finds the node and instance type for a given VMID
func FindInstance(client *pve.Client, vmid string) (node, instanceType string, err error) {
	return client.FindNodeByVMID(vmid)
}

// ValidateVMOnly checks if the instance is a VM (qemu), returns error if it's LXC
func ValidateVMOnly(instanceType string) error {
	if instanceType != "qemu" {
		return fmt.Errorf("this command only works with VMs (qemu), not LXC containers")
	}
	return nil
}

// ValidateLXCOnly checks if the instance is an LXC container, returns error if it's VM
func ValidateLXCOnly(instanceType string) error {
	if instanceType != "lxc" {
		return fmt.Errorf("this command only works with LXC containers, not VMs")
	}
	return nil
}
