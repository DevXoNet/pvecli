# info

Display detailed configuration and status information for a specific VM or container.

## Usage

```bash
# Show VM/container information
pvecli info <vmid>

# Examples
pvecli info 230
pvecli info 500
```

## Output Formats

### JSON (default)

```json
{
  "vmid": 230,
  "name": "k8s-master",
  "status": "running",
  "node": "node04",
  "type": "vm",
  "cpus": 2,
  "memory": 4294967296,
  "maxdisk": 34359738368,
  "uptime": 86400,
  "template": false,
  "agent": "1,fstrim_cloned_disks=1"
}
```

### Table

```
╔════════════════════════════════════════╗
║         VM/Container Information        ║
╚════════════════════════════════════════╝

VMID:     230
Name:     k8s-master
Status:   running
Node:     node04
Type:     vm
CPUs:     2
Memory:   4.0 GB
Disk:     32 GB
Uptime:   1d 0h 0m
Template: false
Agent:    enabled
```

Also supports `yaml` and `text` formats via config file.

## API Endpoints

```
GET /api2/json/nodes/{node}/qemu/{vmid}/status/current
GET /api2/json/nodes/{node}/qemu/{vmid}/config
GET /api2/json/nodes/{node}/lxc/{vmid}/status/current
GET /api2/json/nodes/{node}/lxc/{vmid}/config
```

## Features

- Automatic node detection
- Shows current status and configuration
- Displays uptime in human-readable format
- Shows Guest Agent status
- Works with both VMs and containers

## Examples

### Get VM information as JSON
```bash
pvecli info 230 | jq '.memory'
```

### Check if VM is a template
```bash
pvecli info 230 | jq '.template'
```

### Get uptime
```bash
pvecli info 230 | jq '.uptime'
```

## Required Permissions

- `VM.Audit` - View VM/container information

## See Also

- [list](list.md) - List all VMs and containers
- [status](status.md) - Quick status check
