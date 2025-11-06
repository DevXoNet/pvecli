# console

⚠️ **Status**: Not working with API tokens

Interactive TTY console access to VMs and containers via Proxmox API (similar to AWS SSM).

## Usage

```bash
pvecli console <vmid> [--source <node>] [--target <node>] [--node <node>]
```

### Flags

- `--source NODE` - Source node to create the VNC ticket from (where the API call is made)
- `--target NODE` - Target node to connect the WebSocket to (where the console connection is established)
- `--node NODE` - Legacy flag for node override (kept for backward compatibility)

### Examples

```bash
# Auto-detect node where VM is located
pvecli console 202

# Specify source node (searches VM only on that node)
pvecli console 202 --source node04

# Specify both source and target nodes
pvecli console 202 --source node04 --target node04

# Create ticket from one node, connect to another
pvecli console 202 --source node01 --target node04
```

**Important Notes:**
- Use **node names** (e.g., `node04`), not IP addresses (e.g., `192.168.99.7`)
- IP addresses will cause "proxy loop detected" errors in Proxmox API
- When `--source` is specified, the VM is searched only on that node
- When `--source` is not specified, all nodes are searched automatically

## Intended Functionality

- Open interactive console to VM/container
- No SSH required
- No agent installation needed
- Automatic terminal resizing
- Works with QEMU VMs and LXC containers

## Technical Implementation

### Architecture

```
pvecli → Proxmox API (termproxy) → WebSocket → VM/LXC Console
```

### Process Flow

1. Locate node where VM/container is running
2. Create termproxy session via POST to `/nodes/{node}/{type}/{vmid}/termproxy`
3. Receive VNC ticket and port from Proxmox
4. Attempt WebSocket connection to `/nodes/{node}/vncwebsocket`
5. Send authentication message `username:ticket`

### API Endpoints

**Step 1**: Create termproxy session ✅ Works
```
POST /api2/json/nodes/{node}/qemu/{vmid}/termproxy
```

**Step 2**: WebSocket connection ❌ Fails with 401
```
wss://{host}:8006/api2/json/nodes/{node}/vncwebsocket?port=5900&vncticket=...
```

## Problem

### Error

```bash
Error: websocket dial failed (status 401): websocket: bad handshake
```

### Root Cause

**Proxmox WebSocket termproxy endpoint does not support API token authentication.**

WebSocket connections require full session authentication (session cookie from login), not API tokens.

### Why It Doesn't Work

1. **API Tokens vs Session Cookies**
   - API tokens work for REST API requests
   - WebSocket connections require session cookies
   - Session cookies obtained only via username/password login

2. **Security Model**
   - WebSocket connections are more sensitive than REST API
   - Proxmox requires full session authentication for WebSocket
   - This is a security feature, not a bug

### Attempted Solutions

All attempts failed with 401 Unauthorized:

- ❌ Cookie header with VNC ticket
- ❌ Authentication message after connection
- ❌ Different WebSocket paths
- ❌ Connecting to different nodes
- ❌ Empty headers

## Alternatives

### Option 1: Username/Password Login (Not Recommended)

Login to get session cookie, then use for WebSocket.

**Issues**: Requires storing passwords, no 2FA support, less secure than API tokens.

### Option 2: SSH to Proxmox Node (Recommended)

```bash
ssh root@node04 "qm terminal 230"
```

**Pros**: Works reliably, uses SSH keys, secure.  
**Cons**: Requires SSH access to nodes.

### Option 3: Proxmox Web UI

Use the built-in console in Proxmox web interface.

### Option 4: Feature Request

Request Proxmox team to add API token support for WebSocket connections.

## Implementation

### Files

- `/opt/pvectl/cmd/console.go` - Command implementation
- `/opt/pvectl/internal/console/vnc.go` - WebSocket VNC client
- `/opt/pvectl/internal/proxmox/client.go` - Proxmox API methods

### Dependencies

```go
github.com/gorilla/websocket v1.5.3  // WebSocket client
golang.org/x/term v0.27.0            // Terminal control
golang.org/x/crypto v0.31.0          // Crypto functions
```

## Testing Results

All configurations tested resulted in 401 Unauthorized:

- API Token + cluster IP
- API Token + VM node
- API Token + node IP address
- Different WebSocket paths
- With/without Cookie headers
- Root token without Privilege Separation

## References

- [Proxmox VE API Documentation](https://pve.proxmox.com/pve-docs/api-viewer/)
- [Proxmox pve-xtermjs](https://github.com/proxmox/pve-xtermjs)
- [Forum: termproxy via Python](https://forum.proxmox.com/threads/try-to-use-termproxy-via-python-and-websockets.89194/)

---

**Conclusion**: The console function is technically implemented correctly but cannot work due to Proxmox API limitations regarding WebSocket authentication with API tokens. Use SSH or Proxmox Web UI for console access.
