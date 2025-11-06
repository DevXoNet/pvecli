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

```bash
./pvecli config
```

Enter when prompted:
- API URL: `https://your-proxmox:8006/api2/json`
- Token ID: `root@pam!cli`
- Token Secret: `<your-secret>`
- Skip TLS verification: `yes` (for self-signed certificates)

Configuration is stored in `~/.devxo/pve.yaml`

### 3. Configure Output Format

pvecli supports multiple output formats that can be configured in `~/.devxo/pve.yaml`:

```yaml
# Output format: json, yaml, or text
output_format: json
```

**Available formats:**
- **json** (default): Structured JSON output, ideal for scripting and parsing
- **yaml**: YAML format, human-readable and easy to edit
- **text**: Plain text format, simple key-value pairs


## License

Licensed under the [Apache License 2.0](./LICENSE).

---

## Disclaimer

`pvecli` is an independent open-source project and is **not affiliated with, endorsed, or sponsored by Proxmox Server Solutions GmbH**.  
All trademarks and product names, including **Proxmox**, are the property of their respective owners.