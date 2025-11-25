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

// GetClusterStatus returns cluster status information
func (c *Client) GetClusterStatus() ([]interface{}, error) {
	type wrap struct {
		Data []interface{} `json:"data"`
	}
	var out wrap
	return out.Data, c.doRequest("GET", "/cluster/status", &out)
}

// GetClusterResources returns all cluster resources
func (c *Client) GetClusterResources() ([]interface{}, error) {
	type wrap struct {
		Data []interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", "/cluster/resources", &out)
	return out.Data, err
}

// GetAllStorages returns all storages from datacenter
func (c *Client) GetAllStorages() ([]interface{}, error) {
	type wrap struct {
		Data []interface{} `json:"data"`
	}
	var out wrap
	err := c.doRequest("GET", "/storage", &out)
	return out.Data, err
}
