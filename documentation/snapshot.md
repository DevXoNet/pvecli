# snapshot

Create backup snapshots of VMs and containers to PBS (Proxmox Backup Server) or NFS storage.

## Usage

```bash
pvecli snapshot <vmid> --pbs
pvecli snapshot <vmid> --nfs
```

## Description

The snapshot command creates a backup of a VM or container using Proxmox vzdump API. The backup is created in snapshot mode for minimal downtime and automatically includes metadata notes with VM information.

## Features

- Auto-detects PBS or NFS storage on the target node
- Creates backups in snapshot mode (minimal downtime)
- Automatically adds descriptive notes to each backup
- Supports both VMs (QEMU) and containers (LXC)
- Returns task ID for monitoring backup progress
- Respects configured output format (JSON, YAML, or Text)

## Flags

- `--pbs` - Use PBS (Proxmox Backup Server) storage (auto-detected)
- `--nfs` - Use NFS storage (auto-detected)

Note: Exactly one of `--pbs` or `--nfs` must be specified.

## Examples

### Create PBS backup

```bash
pvecli snapshot 230 --pbs
```

Output (JSON format):
```json
{
  "node": "node04",
  "status": "started",
  "storage": "PBS_HomeLAB",
  "storage_type": "pbs",
  "task_id": "UPID:node04:001B123A:02102141:690CB36E:vzdump:230:root@pam!pvecli:",
  "type": "vm",
  "vmid": "230"
}
```

### Create NFS backup

```bash
pvecli snapshot 202 --nfs
```

Output (JSON format):
```json
{
  "node": "node02",
  "status": "started",
  "storage": "nfs-backup",
  "storage_type": "nfs",
  "task_id": "UPID:node02:001B123A:02102141:690CB36E:vzdump:202:root@pam!pvecli:",
  "type": "vm",
  "vmid": "202"
}
```

### Container backup

```bash
pvecli snapshot 100 --pbs
```

Output (JSON format):
```json
{
  "node": "node01",
  "status": "started",
  "storage": "PBS_HomeLAB",
  "storage_type": "pbs",
  "task_id": "UPID:node01:0020B483:02108968:690CB478:vzdump:100:root@pam!pvecli:",
  "type": "ct",
  "vmid": "100"
}
```

## How It Works

### 1. Instance Location

The command automatically locates which node the VM/container is running on by querying the Proxmox API.

### 2. Storage Auto-Detection

Based on the flag provided (`--pbs` or `--nfs`), the command queries the node's storage configuration and automatically selects the first available storage of the requested type.

If no storage of the requested type is found, an error is returned:

```
Error: no pbs storage found on node node04
```

### 3. Backup Creation

The backup is created using the Proxmox vzdump API with the following parameters:

- **Mode**: `snapshot` - Creates a snapshot for minimal downtime
- **Storage**: Auto-detected PBS or NFS storage
- **Notes Template**: Automatically includes descriptive metadata

### 4. Backup Notes

Each backup automatically includes notes with the following template:

```
Backup of {{guestname}} (VMID: {{vmid}}) from node {{node}} in cluster {{cluster}}
```

Proxmox automatically replaces the template variables:
- `{{guestname}}` - VM/CT name (e.g., "k8s-master")
- `{{vmid}}` - VM ID (e.g., "230")
- `{{node}}` - Node name (e.g., "node04")
- `{{cluster}}` - Cluster name

Example result:
```
Backup of k8s-master (VMID: 230) from node node04 in cluster production
```

## Error Handling

### No flag specified

```bash
pvecli snapshot 230
```

```
Error: either --pbs or --nfs flag is required
```

### Both flags specified

```bash
pvecli snapshot 230 --pbs --nfs
```

```
Error: cannot use both --pbs and --nfs flags together
```

### Storage not found

```bash
pvecli snapshot 230 --pbs
```

```
Error: no pbs storage found on node node04
```

### Wrong storage type

If a storage exists but is not the correct type, the command will skip it and continue searching. If no matching storage is found, an error is returned.

## Output Formats

The command respects the configured output format setting and outputs only the final result without intermediate messages.

### JSON Format (default)

```json
{
  "node": "node04",
  "status": "started",
  "storage": "PBS_HomeLAB",
  "storage_type": "pbs",
  "task_id": "UPID:node04:001B123A:02102141:690CB36E:vzdump:230:root@pam!pvecli:",
  "type": "vm",
  "vmid": "230"
}
```

### YAML Format

```yaml
node: node04
status: started
storage: PBS_HomeLAB
storage_type: pbs
task_id: 'UPID:node04:001B123A:02102141:690CB36E:vzdump:230:root@pam!pvecli:'
type: vm
vmid: "230"
```

### Text Format

```
Backup task started: UPID:node04:001B123A:02102141:690CB36E:vzdump:230:root@pam!pvecli:
Use 'pvecli task <task_id>' to monitor progress
```

## Monitoring Backup Progress

The command returns a task ID (UPID) that can be used to monitor the backup progress. While pvecli does not currently have a built-in task monitoring command, you can check the task status in the Proxmox web interface under:

```
Datacenter > Tasks
```

Or use the Proxmox API directly:

```bash
# Example using curl (requires authentication)
curl -k "https://proxmox-host:8006/api2/json/nodes/node04/tasks/UPID:..."
```

## Technical Details

### API Endpoint

```
POST /api2/json/nodes/{node}/vzdump
```

### Parameters

- `vmid` - VM/CT ID to backup
- `storage` - Target storage name (auto-detected)
- `mode` - Set to "snapshot" for minimal downtime
- `remove` - Set to "0" to keep old backups
- `notes-template` - Template for backup notes

### Storage Types

The command supports two storage types:

1. **PBS (Proxmox Backup Server)** - Deduplicated backup storage
   - Type identifier: `pbs`
   - Recommended for production backups
   - Supports incremental backups and deduplication

2. **NFS (Network File System)** - Network-attached storage
   - Type identifier: `nfs`
   - Standard file-based backups
   - Compatible with any NFS server

## Best Practices

1. **Use PBS for production** - PBS provides better deduplication and incremental backups
2. **Monitor backup tasks** - Check task completion in Proxmox web interface
3. **Verify backups** - Periodically test backup restoration
4. **Schedule regular backups** - Use Proxmox backup scheduler for automated backups
5. **Keep multiple backup copies** - Store backups on different storage systems

## Limitations

- Only creates single VM/CT backups (no bulk operations)
- Requires at least one PBS or NFS storage configured on the node
- Does not wait for backup completion (returns immediately with task ID)
- Cannot specify custom backup retention policies (use Proxmox backup jobs for this)

## Related Commands

- `pvecli list` - List all VMs and containers
- `pvecli status <vmid>` - Check VM/CT status
- `pvecli cluster` - View cluster storage information

## See Also

- [Proxmox Backup Server Documentation](https://pbs.proxmox.com/docs/)
- [Proxmox VE Backup and Restore](https://pve.proxmox.com/wiki/Backup_and_Restore)
