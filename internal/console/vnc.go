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

// TODO: Console provides a Proxmox Serial/VNC console connection (termproxy)

package console

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

// PVEConsole represents a Proxmox Serial/VNC console connection (termproxy)
type PVEConsole struct {
	conn         *websocket.Conn
	originalMode *term.State
	width        int
	height       int
	debug        bool
}

// Connect establishes a WebSocket connection to the Proxmox termproxy console
// baseURL should be the resolvable API endpoint (e.g., https://proxmox-host:8006/api2/json)
func Connect(baseURL, node, vmid, instanceType string, port int, ticket string, insecure bool, debug bool) (*PVEConsole, error) {
	// Parse base URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}

	// Determine WebSocket scheme based on HTTP/HTTPS
	scheme := "wss"
	if u.Scheme == "http" {
		scheme = "ws"
	}

	// Build WebSocket path using node name
	wsPath := fmt.Sprintf("/api2/json/nodes/%s/vncwebsocket", node)

	// Build complete WebSocket URL with VNC ticket
	wsURL := fmt.Sprintf("%s://%s%s?port=%d&vncticket=%s", scheme, u.Host, wsPath, port, url.QueryEscape(ticket))

	if debug {
		fmt.Printf("DEBUG: WebSocket URL: %s\n", wsURL)
		fmt.Printf("DEBUG: Ticket: %s\n", ticket)
	}

	// Configure WebSocket dialer with TLS settings
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	// Use empty headers - authentication is done via vncticket parameter and auth message
	headers := http.Header{}

	if debug {
		fmt.Printf("DEBUG: WebSocket headers: %v\n", headers)
	}

	// Connect to WebSocket
	conn, resp, err := dialer.Dial(wsURL, headers)
	if err != nil {
		if resp != nil {
			// Read response body for more details
			body := ""
			if resp.Body != nil {
				bodyBytes := make([]byte, 1024)
				n, _ := resp.Body.Read(bodyBytes)
				body = string(bodyBytes[:n])
				resp.Body.Close()
			}
			if debug {
				fmt.Printf("DEBUG: WebSocket error response: %s\n", body)
				fmt.Printf("DEBUG: Response headers: %v\n", resp.Header)
			}
			// Status 401 indicates authentication failure.
			return nil, fmt.Errorf("websocket dial failed (status %d): %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("websocket dial: %w", err)
	}

	// Send authentication message in format "username:ticket"
	authMsg := fmt.Sprintf("%s:%s", "root@pam", ticket)
	if debug {
		fmt.Printf("DEBUG: Sending auth message: %s\n", authMsg)
	}

	err = conn.WriteMessage(websocket.TextMessage, []byte(authMsg))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to send auth message: %w", err)
	}

	// Read authentication response from Proxmox
	_, authResp, err := conn.ReadMessage()
	if err != nil {
		if debug {
			fmt.Printf("DEBUG: Failed to read auth response: %v\n", err)
		}
	} else if debug {
		fmt.Printf("DEBUG: Auth response: %s\n", string(authResp))
	}

	console := &PVEConsole{
		conn:  conn,
		debug: debug,
	}

	return console, nil
}

// Start begins the interactive console session
func (c *PVEConsole) Start() error {
	// Get terminal size
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		width, height = 80, 24 // Default size
	}
	c.width = width
	c.height = height

	// Set terminal to raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("set raw mode: %w", err)
	}
	c.originalMode = oldState

	// Ensure we restore terminal on exit
	defer c.restoreTerminal()

	// Send initial resize
	if err := c.sendResize(width, height); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to set terminal size: %v\n", err)
	}

	// Setup signal handler for window resize
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGWINCH)
	go c.handleResize(sigCh)

	// Setup channels for communication
	done := make(chan struct{})
	errCh := make(chan error, 2)

	// Start reading from WebSocket and writing to stdout
	go func() {
		err := c.readFromWebSocket()
		if err != nil {
			errCh <- fmt.Errorf("read from websocket: %w", err)
		}
		close(done)
	}()

	// Start reading from stdin and writing to WebSocket
	go func() {
		err := c.writeToWebSocket()
		if err != nil {
			errCh <- fmt.Errorf("write to websocket: %w", err)
		}
	}()

	// Wait for completion or error
	select {
	case <-done:
		return nil
	case err := <-errCh:
		return err
	}
}

// readFromWebSocket reads data from WebSocket and writes to stdout
func (c *PVEConsole) readFromWebSocket() error {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			return err
		}

		// Write to stdout
		if _, err := os.Stdout.Write(message); err != nil {
			return err
		}
	}
}

// writeToWebSocket reads from stdin and writes to WebSocket
func (c *PVEConsole) writeToWebSocket() error {
	buf := make([]byte, 1024)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		if n > 0 {
			// Send data to WebSocket
			if err := c.conn.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				return err
			}
		}
	}
}

// handleResize handles terminal resize signals
func (c *PVEConsole) handleResize(sigCh chan os.Signal) {
	for range sigCh {
		width, height, err := term.GetSize(int(os.Stdin.Fd()))
		if err != nil {
			continue
		}

		if width != c.width || height != c.height {
			c.width = width
			c.height = height
			if err := c.sendResize(width, height); err != nil {
				fmt.Fprintf(os.Stderr, "\rWarning: failed to resize terminal: %v\n", err)
			}
		}
	}
}

// sendResize sends a resize command to the console
func (c *PVEConsole) sendResize(width, height int) error {
	resizeMsg := map[string]interface{}{
		"command": "resize",
		"width":   width,
		"height":  height,
	}

	data, err := json.Marshal(resizeMsg)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(websocket.TextMessage, data)
}

// restoreTerminal restores the terminal to its original state
func (c *PVEConsole) restoreTerminal() {
	if c.originalMode != nil {
		term.Restore(int(os.Stdin.Fd()), c.originalMode)
	}
}

// Close closes the WebSocket connection and restores terminal
func (c *PVEConsole) Close() error {
	c.restoreTerminal()
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
