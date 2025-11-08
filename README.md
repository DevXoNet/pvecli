# pvecli

`pvecli` is a lightweight, secure, and extensible command-line interface (CLI) designed for efficient management of Proxmox Virtual Environment (PVE) clusters. Built in Go, it provides a unified, developer-friendly way to interact with multiple Proxmox nodes via a REST API, making it ideal for automation, scripting, and daily operations.

## Why pvecli?

Traditional Proxmox cluster management presents several challenges. Administrators often need SSH access to individual nodes, must manually track which VMs reside on which nodes, and rely on separate commands for different operations. While the web UI provides cluster-wide visibility, it's not ideal for automation or scripting workflows.

pvecli addresses these pain points by offering a unified command-line interface that works across your entire cluster. Using Proxmox's REST API and token-based authentication, it eliminates the need for SSH access while providing powerful automation capabilities.

## Key Features

**Cluster-Wide Operations**
- Manage VMs and containers across all nodes from a single command
- Automatic node detection - no need to specify where a VM is located
- Real-time cluster health monitoring with resource usage metrics
- Batch operations for cluster-wide updates

**Flexible Output**
- Multiple output formats: JSON, YAML, plain text, and formatted tables
- Structured output perfect for scripting and integration with other tools
- Configurable default format per user preference

**Backup Management**
- Create snapshots to PBS (Proxmox Backup Server) or NFS storage
- Auto-detection of available backup storage
- List and manage existing backups with filtering options

**Advanced Capabilities**
- Guest Agent detection and status reporting
- Pure API-based operation - no direct node access needed
- Built for automation, scripting, and daily operations

## Installation

### Prerequisites
- Go 1.22 or later
- Proxmox VE 9.0+ cluster with API access
- API token with appropriate permissions

### Build from source

```bash
git clone https://github.com/DevXoNet/pvecli.git
cd pvecli
make setup
make build
```

The binary will be created as `pvecli` in the current directory.

### Cross-platform builds

```bash
make build-all
```

Binaries for all platforms will be in `_build/` directory.

## Initial Setup

### 1. Create Proxmox API Token

In Proxmox Web UI:
1. Navigate to Datacenter → Permissions → API Tokens
2. Click "Add"
3. Select user (e.g., `root@pam`)
4. Set Token ID (e.g., `cli`)
5. Uncheck "Privilege Separation" for full permissions
6. Copy the generated secret

### 2. Configure pvecli

Run the interactive configuration wizard:

```bash
pvecli config
```

Enter when prompted:
- **Environment name**: Name for this cluster (e.g., `prod`, `dev`, `staging`)
- **API URL**: `https://your-proxmox:8006/api2/json`
- **Token ID**: `root@pam!cli`
- **Token Secret**: `<your-secret>`
- **Skip TLS verification**: `yes` (for self-signed certificates)

Configuration is stored in `~/.devxo/pve.yaml`

### 3. Multi-Cluster Setup (Optional)

To manage multiple Proxmox clusters, add additional environments:

```bash
# Add production cluster
pvecli config
# Enter: prod, https://pve-prod.example.com:8006/api2/json, credentials...

# Add development cluster
pvecli config
# Enter: dev, https://pve-dev.example.com:8006/api2/json, credentials...

# List all configured environments
pvecli config list
```

**Using different environments:**

```bash
# Use default environment
pvecli list

# Use specific environment
pvecli --env prod list
pvecli -e dev start 100
pvecli -e staging cluster
```

See [Multi-Cluster Guide](./MULTI_CLUSTER.md) for detailed setup instructions.

### 4. Configure Output Format

pvecli supports multiple output formats that can be configured in `~/.devxo/pve.yaml`:

```yaml
# Output format: json, yaml, or text
output_format: json
```

**Available formats:**
- **json** (default): Structured JSON output, ideal for scripting and parsing
- **yaml**: YAML format, human-readable and easy to edit
- **text**: Plain text format, simple key-value pairs

## Quick Start Examples

### Single Cluster

```bash
# List all VMs and containers
pvecli list --table

# Start a VM
pvecli start 100

# Check cluster status
pvecli cluster

# Real-time monitoring
pvecli top
```

### Multi-Cluster Management

```bash
# List VMs in production cluster
pvecli -e prod list --table

# Start VM in development cluster
pvecli -e dev start 100

# Migrate VM in staging cluster
pvecli -e staging migrate 200 node02

# Check status across different clusters
pvecli -e prod cluster
pvecli -e dev cluster
pvecli -e staging cluster
```

## Available Commands

### VM/Container Management
- **list** - List all VMs and containers across the cluster
- **info** - Show detailed VM/container configuration
- **status** - Check VM/container status
- **start** - Start a VM/container
- **stop** - Stop a VM/container
- **reboot** - Reboot a VM/container
- **shutdown** - Gracefully shutdown a VM/container
- **suspend** - Suspend and resume VMs

### Advanced Operations
- **clone** - Clone VM templates (full/linked clones)
- **migrate** - Migrate VM/container to another node
- **backup** - Create and manage backups (PBS/NFS)
- **snapshot** - Create and manage VM snapshots
- **template** - Convert VM to template
- **disk** - Disk management operations

### Monitoring & Management
- **top** - Real-time resource monitoring with graphs
- **task** - Monitor and manage Proxmox tasks
- **cluster** - View cluster information and health
- **console** - Interactive console access (VNC/SPICE)

### Configuration
- **config** - Configure pvecli settings and manage multiple cluster environments

For detailed documentation on each command, see the [wiki](https://github.com/DevXoNet/pvecli/wiki).


## Disclaimer

`pvecli` is an independent open-source project and is **not affiliated with, endorsed, or sponsored by Proxmox Server Solutions GmbH**.  
All trademarks and product names, including **Proxmox**, are the property of their respective owners.


## License

Licensed under the [Apache License 2.0](./LICENSE).


---
