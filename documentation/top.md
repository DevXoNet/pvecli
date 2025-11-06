# top

Real-time monitoring of VMs and containers with live statistics and visual graphs.

## Usage

```bash
# Start monitoring with default settings (2s refresh)
pvecli top

# Custom refresh interval
pvecli top --interval 5

# Sort by different metrics
pvecli top --sort-by cpu    # Sort by CPU usage (default)
pvecli top --sort-by mem    # Sort by memory usage
pvecli top --sort-by disk   # Sort by disk I/O
pvecli top --sort-by net    # Sort by network I/O
pvecli top --sort-by id     # Sort by VM ID
```

## Parameters

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--interval` | `-i` | int | 2 | Refresh interval in seconds |
| `--sort-by` | `-s` | string | cpu | Sort by: cpu, mem, disk, net, id |

## Display

### Header
```
╔════════════════════════════════════════════════════════════════════════════════════════════════╗
║  PVECLI TOP - Real-time VM/Container Monitor                                                   ║
║  Refresh: 2s | Sort: cpu        | Time: 2025-11-06 12:20:15                                   ║
╚════════════════════════════════════════════════════════════════════════════════════════════════╝
```

### Summary
```
VMs/CTs: 15 total, 12 running | Avg CPU: 25.3% | Total Memory: 48.5GB / 128GB (37.9%)
```

### Statistics Table

| Column | Description |
|--------|-------------|
| ID | VM/Container ID |
| NAME | VM/Container name |
| NODE | Proxmox node |
| TYPE | VM or CT (container) |
| STATUS | RUN, STOP, PAUSE |
| CPU% | CPU usage percentage |
| CPU | Visual bar graph of CPU usage |
| MEMORY | Used/Total memory |
| MEM% | Visual bar graph of memory usage |
| DISK I/O | Read/Write disk I/O |
| NET I/O | Download/Upload network traffic |

### Example Output

```
ID    NAME           NODE    TYPE  STATUS  CPU%   CPU                    MEMORY           MEM%                   DISK I/O              NET I/O
230   k8s-master     node04  VM    ✓ RUN   45.2   ████████░░ 45.2%      4.2GB/8.0GB      █████░░░░░ 52.5%      R:1.2GB W:850MB      ↓2.5GB ↑1.8GB
500   web-server     node01  VM    ✓ RUN   12.5   ██░░░░░░░░ 12.5%      2.1GB/4.0GB      █████░░░░░ 52.5%      R:450MB W:120MB      ↓850MB ↑420MB
502   db-primary     node02  VM    ✓ RUN   78.9   ███████░░░ 78.9%      15.2GB/16.0GB    █████████░ 95.0%      R:8.5GB W:6.2GB      ↓1.2GB ↑980MB
105   nginx-proxy    node03  CT    ✓ RUN   5.2    █░░░░░░░░░ 5.2%       512MB/2.0GB      ██░░░░░░░░ 25.0%      R:85MB W:42MB        ↓450MB ↑320MB
231   test-vm        node04  VM    ✗ STOP  -      -                     -                -                      -                    -
```

## Features

### Real-time Updates
- Auto-refresh at specified interval
- Live statistics from Proxmox API
- No caching - always current data

### Visual Graphs
- **CPU bars**: `████████░░ 45.2%`
- **Memory bars**: `█████░░░░░ 52.5%`
- Color-coded status indicators

### Sorting Options
- **CPU**: Highest CPU usage first
- **Memory**: Highest memory usage first
- **Disk**: Highest disk I/O first
- **Network**: Highest network traffic first
- **ID**: Numerical order by VM ID

### Metrics Tracked

#### CPU
- Current CPU usage percentage
- Number of CPU cores allocated
- Visual bar graph

#### Memory
- Used memory
- Total allocated memory
- Usage percentage with bar graph

#### Disk I/O
- Total bytes read
- Total bytes written
- Cumulative since VM start

#### Network I/O
- Total bytes received (download)
- Total bytes sent (upload)
- Cumulative since VM start

## Examples

### Monitor with 5-second refresh
```bash
pvecli top --interval 5
```

### Sort by memory usage
```bash
pvecli top --sort-by mem
```

### Sort by network traffic
```bash
pvecli top --sort-by net
```

### Quick CPU check
```bash
pvecli top --interval 1 --sort-by cpu
```

## Keyboard Controls

- **Ctrl+C** - Exit monitoring

## Notes

### Performance
- Queries all nodes in cluster
- Fetches stats for all VMs/containers
- Network overhead depends on cluster size
- Recommended interval: 2-5 seconds

### Stopped VMs
- Shown with status `✗ STOP`
- No statistics displayed
- Sorted to bottom of list

### Data Accuracy
- Statistics are real-time from Proxmox
- Disk/Network I/O are cumulative totals
- CPU/Memory are current values

## API Endpoints

```
GET /api2/json/nodes
GET /api2/json/nodes/{node}/qemu
GET /api2/json/nodes/{node}/lxc
GET /api2/json/nodes/{node}/qemu/{vmid}/status/current
GET /api2/json/nodes/{node}/lxc/{vmid}/status/current
```

## Required Permissions

- `VM.Audit` - View VM/container information and statistics

## Comparison with Other Tools

| Feature | pvecli top | Proxmox Web UI | htop |
|---------|-----------|----------------|------|
| Real-time monitoring | ✅ | ✅ | ✅ |
| Cluster-wide view | ✅ | ❌ | ❌ |
| CLI-based | ✅ | ❌ | ✅ |
| Network I/O | ✅ | ✅ | ❌ |
| Disk I/O | ✅ | ✅ | ❌ |
| Visual graphs | ✅ | ✅ | ✅ |
| Sorting options | ✅ | ✅ | ✅ |

## Troubleshooting

### High CPU usage
If `pvecli top` uses too much CPU, increase refresh interval:
```bash
pvecli top --interval 10
```

### Connection timeouts
For large clusters, increase timeout in config or reduce refresh rate.

### Missing statistics
- Ensure VMs are running
- Check API token permissions
- Verify network connectivity

## See Also

- [list](list.md) - List all VMs and containers
- [info](info.md) - Detailed VM information
- [status](status.md) - Quick status check
- [cluster](cluster.md) - Cluster overview
