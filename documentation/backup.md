# backup

Manage backups on PBS or NFS storage.

## Usage

```bash
pvecli backup list --storage <storage> [flags]
pvecli backup delete <volid> --storage <storage>
```

## Description

The backup command provides tools to list and delete backups stored on PBS (Proxmox Backup Server) or NFS storage. Works with both VM (QEMU) and container (LXC) backups.

## Commands

### backup list

List all backups on specified storage. You can specify either a storage name or storage type (pbs/nfs).

```bash
# Using storage type (auto-detects first matching storage)
pvecli backup list -s pbs
pvecli backup list --storage nfs

# Using specific storage name
pvecli backup list -s PBS_HomeLAB
pvecli backup list --storage PBS_HomeLAB --vmid 230
```

**Flags:**
- `--storage, -s` - Storage name or type (pbs/nfs) - required
- `--vmid` - Filter by VMID (optional)

**Output (JSON):**
```json
[
  {
    "volid": "PBS_HomeLAB:backup/vm/230/2025-11-06T14:45:00Z",
    "vmid": 230,
    "format": "pbs-vm",
    "size": 5368709120,
    "ctime": 1730905500,
    "notes": "Backup of k8s-master (VMID: 230) from node node04 in cluster production",
    "content": "backup"
  },
  {
    "volid": "PBS_HomeLAB:backup/ct/100/2025-11-06T14:45:00Z",
    "vmid": 100,
    "format": "pbs-ct",
    "size": 1073741824,
    "ctime": 1730905600,
    "notes": "Backup of docker-host (VMID: 100) from node node01 in cluster production",
    "content": "backup"
  }
]
```

### backup delete

Delete a backup by its volume ID. You can specify either a storage name or storage type (pbs/nfs).

```bash
# Using storage type
pvecli backup delete 'PBS_HomeLAB:backup/vm/230/2025-11-06T14:45:00Z' -s pbs

# Using specific storage name
pvecli backup delete 'PBS_HomeLAB:backup/vm/230/2025-11-06T14:45:00Z' --storage PBS_HomeLAB
```

**Flags:**
- `--storage, -s` - Storage name or type (pbs/nfs) - required

**Output (JSON):**
```json
{
  "volid": "PBS_HomeLAB:backup/vm/230/2025-11-06T14:45:00Z",
  "storage": "PBS_HomeLAB",
  "action": "delete",
  "status": "success"
}
```

## Output Formats

### JSON Format (list)
```json
[
  {
    "volid": "PBS_HomeLAB:backup/vm/230/2025-11-06T14:45:00Z",
    "vmid": 230,
    "format": "pbs-vm",
    "size": 5368709120
  }
]
```

### YAML Format (list)
```yaml
- volid: PBS_HomeLAB:backup/vm/230/2025-11-06T14:45:00Z
  vmid: 230
  format: pbs-vm
  size: 5368709120
```

### Text Format (list)
```
PBS_HomeLAB:backup/vm/230/2025-11-06T14:45:00Z
PBS_HomeLAB:backup/vm/230/2025-11-05T14:45:00Z
PBS_HomeLAB:backup/vm/240/2025-11-06T14:45:00Z
```

## Volume ID Format

Volume IDs have the format:
```
<storage>:backup/<type>/<vmid>/<timestamp>
```

Examples:
```
# VM backup
PBS_HomeLAB:backup/vm/230/2025-11-06T14:45:00Z

# Container backup
PBS_HomeLAB:backup/ct/100/2025-11-06T14:45:00Z
```

Components:
- `storage` - Storage name (e.g., PBS_HomeLAB)
- `type` - Instance type (vm for VMs, ct for containers)
- `vmid` - VM/Container ID
- `timestamp` - Backup timestamp in ISO format

## Examples

### List all backups using storage type

```bash
# List all PBS backups (auto-detects first PBS storage)
pvecli backup list -s pbs

# List all NFS backups
pvecli backup list -s nfs
```

### List all backups using storage name

```bash
pvecli backup list --storage PBS_HomeLAB
```

### List backups for specific VM

```bash
# Using storage type
pvecli backup list -s pbs --vmid 230

# Using storage name
pvecli backup list --storage PBS_HomeLAB --vmid 230
```

### List backups for specific container

```bash
# Using storage type
pvecli backup list -s pbs --vmid 100

# Using storage name
pvecli backup list --storage PBS_HomeLAB --vmid 100
```

### Delete old backup

```bash
# First, list backups to find the volid
pvecli backup list -s pbs --vmid 230

# Then delete specific backup using storage type
pvecli backup delete 'PBS_HomeLAB:backup/vm/230/2025-11-05T14:45:00Z' -s pbs

# Or using storage name
pvecli backup delete 'PBS_HomeLAB:backup/vm/230/2025-11-05T14:45:00Z' --storage PBS_HomeLAB
```

### Automation example

```bash
# Keep only last 3 backups for VM 230 (using storage type)
pvecli backup list -s pbs --vmid 230 | \
  jq -r '.[3:] | .[] | .volid' | \
  while read volid; do
    pvecli backup delete "$volid" -s pbs
  done

# Or using specific storage name
pvecli backup list --storage PBS_HomeLAB --vmid 230 | \
  jq -r '.[3:] | .[] | .volid' | \
  while read volid; do
    pvecli backup delete "$volid" --storage PBS_HomeLAB
  done
```

## Storage Specification

You can specify storage in two ways:

### 1. Storage Type (Auto-Detection)

Use `pbs` or `nfs` to automatically detect the first available storage of that type:

```bash
pvecli backup list -s pbs
pvecli backup list -s nfs
```

The command will:
- Query all available storages on the cluster
- Find the first storage matching the specified type
- Use that storage for the operation
- Return an error if no matching storage is found

### 2. Storage Name (Explicit)

Use the exact storage name for precise control:

```bash
pvecli backup list -s PBS_HomeLAB
pvecli backup list -s my-nfs-backup
```

## Storage Types

Supported storage types for auto-detection:
- **pbs** - Proxmox Backup Server (deduplicated backup storage)
- **nfs** - Network File System storage

Other storage types (must use explicit name):
- **CIFS** - Windows file shares
- **Directory** - Local directory storage

## Important Notes

1. **VMs and Containers** - The command works with both VM (QEMU) and container (LXC) backups
2. **Backup format** - VMs use `pbs-vm` format, containers use `pbs-ct` format
3. **VMID filtering** - The `--vmid` flag works for both VMs and containers
4. **Volume ID** - Always includes the type (vm or ct) in the path

## Use Cases

1. **Cleanup old backups** - Remove outdated backups to free space
2. **Audit backups** - List all backups for compliance
3. **Backup verification** - Check if backups exist
4. **Automation** - Script backup retention policies

## API Endpoints

- `GET /nodes/{node}/storage/{storage}/content` - List backups
- `DELETE /nodes/{node}/storage/{storage}/content/{volid}` - Delete backup

## See Also

- [snapshot](snapshot.md) - Create backups
- [task](task.md) - Monitor backup tasks
