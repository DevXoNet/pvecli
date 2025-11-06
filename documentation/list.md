# list

List all virtual machines and LXC containers across the Proxmox cluster.

## Usage

```bash
# List all VMs and containers
pvecli list

# Filter by node
pvecli list --node node01

# Show only running instances
pvecli list --running

# Filter by type
pvecli list --type vm    # VMs only
pvecli list --type ct    # Containers only

# Force table output
pvecli list --table

# Combine filters
pvecli list --node node01 --running --type vm
```

## Parameters

| Flag | Type | Description |
|------|------|-------------|
| `--node` | string | Filter by specific node |
| `--running` | bool | Show only running instances |
| `--type` | string | Filter by type: `vm` or `ct` |
| `--table` | bool | Force table output format |

## Output Formats

### JSON (default)

```json
{
  "vms": [{
    "VMID": 230,
    "Name": "k8s-master",
    "Status": "running",
    "Node": "node04",
    "Type": "vm",
    "CPUs": 2,
    "Memory": 4294967296,
    "Disk": 34359738368,
    "GuestAgent": true
  }]
}
```

### Table

```
+------+-------------+---------+--------+------+------+--------+------+-------+
| VMID |    NAME     | STATUS  |  NODE  | TYPE | CPUS | MEMORY | DISK | AGENT |
+------+-------------+---------+--------+------+------+--------+------+-------+
|  230 | k8s-master  | running | node04 | vm   |    2 | 4.0GB  | 32GB | YES   |
+------+-------------+---------+--------+------+------+--------+------+-------+
```

Also supports `yaml` and `text` formats via config file.

## API Endpoints

```
GET /api2/json/nodes
GET /api2/json/nodes/{node}/qemu
GET /api2/json/nodes/{node}/lxc
GET /api2/json/nodes/{node}/qemu/{vmid}/config
```

## Features

### Auto-discovery
- Automatically discovers all nodes in cluster
- Retrieves VMs/containers from all nodes
- No need to specify node

### Guest Agent Detection
Shows whether VM/container has Proxmox Guest Agent installed:
- **YES** - Agent configured and active
- **NO** - Agent not configured

## Examples

### Find all running VMs on node01
```bash
pvecli list --node node01 --running --type vm
```

### Export to JSON for scripting
```bash
pvecli list | jq '.vms[] | select(.Status == "running") | .VMID'
```

### Check Guest Agent status
```bash
pvecli list --table | grep "NO"
```

### Filter large VMs
```bash
pvecli list | jq '.vms[] | select(.CPUs > 4)'
```

## Required Permissions

- `VM.Audit` - View VM/container information

## See Also

- [info](info.md) - Detailed VM/container information
- [status](status.md) - VM/container status
