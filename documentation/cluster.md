# cluster

Display cluster overview and resource information.

## Usage

```bash
# Show cluster overview
pvecli cluster

# Show cluster resources
pvecli cluster resources
```

## Output

### Cluster Overview

```
╔════════════════════════════════════════╗
║         Proxmox Cluster Overview        ║
╚════════════════════════════════════════╝

Cluster: production-cluster

Nodes:
  • node01 (192.168.99.4) - online
  • node02 (192.168.99.5) - online
  • node03 (192.168.99.6) - online
  • node04 (192.168.99.7) - online

Resources:
  VMs:        15
  Containers: 8
  Total:      23

Status: All nodes online
```

### Cluster Resources

Shows detailed resource usage per node:

```
Node: node01
  CPU:    25% (4/16 cores)
  Memory: 45% (18GB/40GB)
  VMs:    4
  Status: online

Node: node02
  CPU:    15% (2/16 cores)
  Memory: 30% (12GB/40GB)
  VMs:    3
  Status: online
```

## Subcommands

| Subcommand | Description |
|------------|-------------|
| `cluster` | Show cluster overview |
| `cluster resources` | Show detailed resources |
| `cluster update` | Update all nodes (see [cluster-update.md](cluster-update.md)) |

## API Endpoints

```
GET /api2/json/nodes
GET /api2/json/cluster/resources
GET /api2/json/nodes/{node}/status
```

## Features

- Cluster-wide overview
- Node status monitoring
- Resource usage statistics
- VM/container count per node
- Online/offline status

## Examples

### Check cluster health
```bash
pvecli cluster | grep -i "status"
```

### Get node count
```bash
pvecli cluster | grep "Nodes:" -A 10 | grep "•" | wc -l
```

### Monitor resources
```bash
watch -n 5 pvecli cluster resources
```

### Export cluster info
```bash
pvecli cluster > cluster-status.txt
```

## Required Permissions

- `Sys.Audit` - View cluster information

## See Also

- [cluster update](cluster-update.md) - Update all nodes
- [list](list.md) - List VMs and containers
