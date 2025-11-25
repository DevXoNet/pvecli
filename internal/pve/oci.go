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

import "fmt"

// OCI Image support for Proxmox VE 9.1+
// NOTE: In Proxmox VE 9.1, OCI support is implemented via LXC containers
// that use OCI images. There is no separate /nodes/{node}/oci endpoint.
// OCI containers are managed through the standard LXC API endpoints.
// Only the query-oci-repo-tags endpoint is available for searching registry tags.

// QueryOCIRepoTags queries available tags for an OCI image in a registry (Proxmox VE 9.1+)
func (c *Client) QueryOCIRepoTags(node, repository string) ([]string, error) {
	type wrap struct {
		Data []string `json:"data"`
	}
	var out wrap
	path := fmt.Sprintf("/nodes/%s/query-oci-repo-tags?reference=%s", node, repository)
	err := c.doRequest("GET", path, &out)
	return out.Data, err
}

