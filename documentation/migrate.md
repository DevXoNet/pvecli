# migrate

Migrate a VM or container to another node in the cluster.

## Usage

```bash
pvecli migrate <vmid> --target <node> [flags]
```

## Description

Migrates a running or stopped VM or container from one node to another within the Proxmox cluster. Supports both offline and online (live) migration. The command automatically detects the current node and instance type.

## Flags

- `--target <node>` - Target node name (required)
- `--online` - Perform live (online) migration (default: false)
- `--wait` - Wait for the migration task to complete (default: true)
- `--timeout <seconds>` - Wait timeout in seconds (default: 600)

## Examples

### Offline migration

```bash
pvecli migrate 230 --target node02
```

**Output:**
```
Migrating vm/230 from node04 to node02 (online=false) ...
Task started: UPID:node04:00001234:00000000:00000000:qmmigrate:230:root@pam:
Waiting for migration to complete...
✓ Migration completed successfully
```

### Online (live) migration

```bash
pvecli migrate 230 --target node02 --online
```

**Output:**
```
Migrating vm/230 from node04 to node02 (online=true) ...
Task started: UPID:node04:00001234:00000000:00000000:qmmigrate:230:root@pam:
Waiting for migration to complete...
✓ Migration completed successfully
```

### Migrate without waiting

```bash
pvecli migrate 230 --target node02 --wait=false
```

**Output:**
```
Migrating vm/230 from node04 to node02 (online=false) ...
Task started: UPID:node04:00001234:00000000:00000000:qmmigrate:230:root@pam:
Hint: use --wait to wait for completion
```

### Migrate container

```bash
pvecli migrate 100 --target node03
```

**Output:**
```
Migrating lxc/100 from node01 to node03 (online=false) ...
Task started: UPID:node01:00005678:00000000:00000000:vzmigrate:100:root@pam:
Waiting for migration to complete...
✓ Migration completed successfully
```

## Migration Types

### Offline Migration (default)
- VM/container is stopped during migration
- Faster and more reliable
- Recommended for most use cases
- No service interruption if properly scheduled

### Online Migration (--online)
- VM/container continues running during migration
- Minimal downtime (usually < 1 second)
- Requires shared storage
- Only works for running instances
- More resource intensive

## Behavior

- Automatically detects source node and instance type
- Validates target node exists in cluster
- Creates migration task and monitors progress
- Returns success/failure status
- Supports both VMs (QEMU) and containers (LXC)

## Use Cases

1. **Load balancing** - Distribute VMs across cluster nodes
2. **Maintenance** - Move VMs before node maintenance
3. **Resource optimization** - Move VMs to nodes with more resources
4. **High availability** - Relocate VMs from failing nodes
5. **Storage migration** - Move VMs to different storage

## API Endpoints

- `POST /nodes/{node}/qemu/{vmid}/migrate` - Migrate VM
- `POST /nodes/{node}/lxc/{vmid}/migrate` - Migrate container
- `GET /nodes/{node}/tasks/{upid}/status` - Check task status

## Error Handling

If target node doesn't exist:
```
Error: target node not found
```

If VM/container doesn't exist:
```
Error: VM 999 not found
```

If migration fails:
```
Error: migration failed: insufficient resources on target node
```

## Notes

- Online migration requires:
  - Shared storage between nodes
  - Instance must be running
  - Sufficient resources on target node
- Offline migration works with local storage
- Migration time depends on:
  - VM/container size
  - Network speed
  - Storage performance
  - Migration type (online/offline)

## Required Permissions

- `VM.Migrate` - Migrate VMs and containers
- `VM.Allocate` - Allocate resources on target node

## See Also

- [info](info.md) - View VM/container details
- [status](status.md) - Check instance status
- [cluster](cluster.md) - View cluster information
- [task](task.md) - Monitor task progress
