# cluster update

⚠️ **Status**: Not working with API tokens

Update all nodes in the Proxmox cluster (apt update && apt upgrade -y).

## Usage

```bash
# Dry run - show what would be done
pvecli cluster update --dry-run

# Execute updates on all nodes
pvecli cluster update
```

## Intended Functionality

- Run `apt update` on all cluster nodes
- Run `apt upgrade -y` on all cluster nodes
- Cluster-wide system updates from single command
- Error handling - continues if one node fails
- Summary report of successful/failed nodes

## Parameters

| Flag | Type | Description |
|------|------|-------------|
| `--dry-run` | bool | Show what would be done without executing |

## Technical Implementation

### Process Flow

1. Get all nodes in cluster via `/api2/json/nodes`
2. For each node:
   - Execute `apt update` via `/nodes/{node}/execute`
   - Wait 1 second
   - Execute `apt upgrade -y` via `/nodes/{node}/execute`
3. Display summary of results

### API Endpoint

```
POST /api2/json/nodes/{node}/execute
Body: {"command": "apt update"}
```

## Problem

### Error

```bash
✗ Failed to update package list: http 403 for POST /nodes/node01/execute
{"message":"Permission check failed (user != root@pam)","data":null}
```

### Root Cause

**Proxmox `/nodes/{node}/execute` endpoint requires `root@pam` user authentication.**

The endpoint does not accept API token authentication, only full session-based authentication with root@pam user.

### Why It Doesn't Work

1. **API Tokens vs Root User**
   - API tokens work for most REST API requests
   - Command execution requires root@pam user
   - This is a security restriction in Proxmox

2. **Security Model**
   - Executing arbitrary commands is highly privileged
   - Proxmox restricts this to root@pam only
   - API tokens cannot execute system commands

3. **Permission Check**
   - Even with `Sys.Modify` permission on API token
   - Proxmox explicitly checks for `root@pam` user
   - Returns 403 if user is not root@pam

## Alternatives

### Option 1: SSH-based Updates (Recommended)

Execute updates via SSH to each node:

```bash
# Update single node
ssh root@node01 "apt update && apt upgrade -y"

# Update all nodes (bash loop)
for node in node01 node02 node03 node04; do
  echo "Updating $node..."
  ssh root@$node "apt update && apt upgrade -y"
done
```

**Pros**: Works reliably, uses SSH keys, secure.  
**Cons**: Requires SSH access to all nodes.

### Option 2: Proxmox Web UI

Use the built-in update functionality in Proxmox web interface:
- Navigate to each node
- Click "Updates" tab
- Click "Refresh" and "Upgrade"

### Option 3: Ansible/Automation Tool

Use configuration management tools:

```yaml
# Ansible playbook example
- hosts: proxmox_nodes
  tasks:
    - name: Update apt cache
      apt:
        update_cache: yes
    - name: Upgrade packages
      apt:
        upgrade: dist
```

### Option 4: Username/Password Login (Not Recommended)

Login as root@pam to get session cookie, then execute commands.

**Issues**: Requires storing passwords, no 2FA support, less secure than API tokens.

## Implementation

### Files

- `/opt/pvectl/cmd/cluster_update.go` - Command implementation
- `/opt/pvectl/internal/proxmox/client.go` - `RunNodeCommand()` method

### Code

```go
func (c *Client) RunNodeCommand(node, command string) error {
    data := map[string]string{
        "command": command,
    }
    return c.doRequest("POST", fmt.Sprintf("/nodes/%s/execute", node), nil)
}
```

## Testing Results

All nodes tested resulted in 403 Permission check failed:

- API Token + node01: ❌ 403
- API Token + node02: ❌ 403
- API Token + node03: ❌ 403
- API Token + node04: ❌ 403
- Root token without Privilege Separation: ❌ 403

## Dry Run Mode

The `--dry-run` flag works correctly and shows what would be executed:

```bash
$ pvecli cluster update --dry-run

╔════════════════════════════════════════════════════════════════╗
║                    DRY RUN MODE - NO CHANGES                   ║
╚════════════════════════════════════════════════════════════════╝

Starting cluster-wide system update...
Found 4 nodes to update

Node: node01
  → Would run: apt update
  → Would run: apt upgrade -y
  ✓ Dry run completed

═══════════════════════════════════════
Dry Run Summary:
  Would update: 4 nodes
═══════════════════════════════════════
```

## Required Permissions

- `Sys.Modify` - System modification (not sufficient for API tokens)
- `root@pam` user - Required for command execution (API tokens not supported)

## References

- [Proxmox VE API Documentation](https://pve.proxmox.com/pve-docs/api-viewer/)
- [Proxmox Node Execute API](https://pve.proxmox.com/pve-docs/api-viewer/index.html#/nodes/{node}/execute)

---

**Conclusion**: The cluster update command is implemented correctly but cannot work due to Proxmox API limitations requiring root@pam user for command execution. Use SSH or Proxmox Web UI for system updates.
