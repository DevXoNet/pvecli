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

package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client connection
type Client struct {
	client *ssh.Client
	host   string
}

// NewClient creates a new SSH client
func NewClient(host, user, keyPath string, port int) (*Client, error) {
	var authMethods []ssh.AuthMethod

	// Try to read and parse private key
	if keyPath != "" {
		key, err := os.ReadFile(keyPath)
		if err == nil {
			// Try parsing without passphrase first
			signer, err := ssh.ParsePrivateKey(key)
			if err == nil {
				authMethods = append(authMethods, ssh.PublicKeys(signer))
			}
		}
	}

	// If no auth methods available, return error
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no valid authentication methods available (key: %s)", keyPath)
	}

	// SSH client config
	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	// Connect to SSH server
	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &Client{
		client: client,
		host:   host,
	}, nil
}

// RunCommand executes a command on the remote host
func (c *Client) RunCommand(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

// Close closes the SSH connection
func (c *Client) Close() error {
	return c.client.Close()
}

// GetDefaultSSHKeyPath returns the default SSH key path
func GetDefaultSSHKeyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	
	// Try common SSH key locations
	keys := []string{
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
	}
	
	for _, key := range keys {
		if _, err := os.Stat(key); err == nil {
			return key
		}
	}
	
	return filepath.Join(home, ".ssh", "id_rsa")
}
