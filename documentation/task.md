# task

Manage and monitor Proxmox tasks (backup, migration, reboot, etc.).

## Usage

```bash
pvecli task list [flags]
pvecli task status <upid>
pvecli task log <upid>
pvecli task wait <upid>
```

## Description

The task command provides tools to monitor and manage asynchronous tasks in Proxmox. Tasks are created by operations like backup, migration, reboot, and other long-running operations.

## Commands

### task list

List recent tasks from all nodes or a specific node.

```bash
pvecli task list
pvecli task list --limit 10
pvecli task list --node node04
```

**Flags:**
- `--limit, -l` - Number of tasks to show (default: 5)
- `--node, -n` - Filter by specific node

**Output (JSON):**
```json
[
  {
    "upid": "UPID:node04:001B2849:0211CDCC:690CB7B8:qmreboot:230:root@pam!pvecli:",
    "type": "qmreboot",
    "status": "stopped",
    "node": "node04",
    "user": "root@pam",
    "starttime": 1762441144,
    "endtime": 1762441145,
    "exitstatus": "OK"
  }
]
```

### task status

Get the current status of a task by its UPID.

```bash
pvecli task status 'UPID:node04:001B2849:0211CDCC:690CB7B8:qmreboot:230:root@pam!pvecli:'
```

**Important:** Use single quotes around the UPID to prevent shell interpretation of special characters.

**Output (JSON):**
```json
{
  "exitstatus": "OK",
  "id": "230",
  "node": "node04",
  "pid": 1779785,
  "pstart": 34721228,
  "starttime": 1762441144,
  "status": "stopped",
  "tokenid": "pvecli",
  "type": "qmreboot",
  "upid": "UPID:node04:001B2849:0211CDCC:690CB7B8:qmreboot:230:root@pam!pvecli:",
  "user": "root@pam"
}
```

### task log

Get the log output of a task.

```bash
pvecli task log 'UPID:node04:001B2849:0211CDCC:690CB7B8:qmreboot:230:root@pam!pvecli:'
pvecli task log 'UPID:...' --lines 50
```

**Flags:**
- `--lines, -n` - Number of lines to show (0 = all, default: 0)

**Output (Text format):**
```
starting qemu
QEMU[230]: starting VM
...
```

### task wait

Wait for a task to complete and return its final status.

```bash
pvecli task wait 'UPID:node04:001B2849:0211CDCC:690CB7B8:qmreboot:230:root@pam!pvecli:'
```

This command blocks until the task completes and returns the final status.

**Output (Text format):**
```
Task completed successfully
```

Or if failed:
```
Task failed with exit status: ERROR
```

## UPID Format

UPID (Unique Process ID) format:
```
UPID:node:PID:PSTART:STARTTIME:TYPE:ID:USER:
```

Example:
```
UPID:node04:001B2849:0211CDCC:690CB7B8:qmreboot:230:root@pam!pvecli:
```

Components:
- `node` - Node name (e.g., node04)
- `PID` - Process ID
- `PSTART` - Process start counter
- `STARTTIME` - Unix timestamp
- `TYPE` - Task type (vzdump, qmreboot, migrate, etc.)
- `ID` - Resource ID (VMID)
- `USER` - User who started the task

## Shell Quoting

Always use single quotes when passing UPID as argument to prevent shell interpretation:

```bash
# Correct
pvecli task status 'UPID:node04:...:root@pam!pvecli:'

# Wrong (will fail in zsh/bash)
pvecli task status UPID:node04:...:root@pam!pvecli:
```

The exclamation mark (!) is a special character in most shells and must be quoted.

## Task Types

Common task types:
- `vzdump` - Backup operation
- `qmreboot` - VM reboot
- `qmstart` - VM start
- `qmstop` - VM stop
- `qmshutdown` - VM shutdown
- `qmmigrate` - VM migration
- `qmclone` - VM clone
- `vzmigrate` - Container migration

## Output Formats

All task commands respect the configured output format.

### JSON Format
```json
{
  "status": "stopped",
  "exitstatus": "OK",
  "upid": "UPID:..."
}
```

### YAML Format
```yaml
status: stopped
exitstatus: OK
upid: UPID:...
```

### Text Format
```
Task: UPID:...
Status: stopped
Exit Status: OK
```

## Examples

### Monitor a backup task

```bash
# Start backup
pvecli snapshot 230 --pbs

# Output shows task ID
{
  "task_id": "UPID:node04:001B0C7C:020FBD99:690CB26F:vzdump:230:root@pam!pvecli:",
  ...
}

# Check status
pvecli task status 'UPID:node04:001B0C7C:020FBD99:690CB26F:vzdump:230:root@pam!pvecli:'

# View log
pvecli task log 'UPID:node04:001B0C7C:020FBD99:690CB26F:vzdump:230:root@pam!pvecli:'

# Wait for completion
pvecli task wait 'UPID:node04:001B0C7C:020FBD99:690CB26F:vzdump:230:root@pam!pvecli:'
```

### List recent tasks

```bash
# Show last 5 tasks (default)
pvecli task list

# Show last 20 tasks
pvecli task list --limit 20

# Show tasks from specific node
pvecli task list --node node04

# Show tasks in YAML format
pvecli task list --limit 10
```

## Use Cases

1. **Monitor backup progress** - Check if backup completed successfully
2. **Debug failed operations** - View logs of failed tasks
3. **Track cluster activity** - See what operations are running
4. **Automation** - Wait for task completion in scripts

## API Endpoints

- `GET /nodes/{node}/tasks` - List tasks
- `GET /nodes/{node}/tasks/{upid}/status` - Get task status
- `GET /nodes/{node}/tasks/{upid}/log` - Get task log

## See Also

- [snapshot](snapshot.md) - Create backups (returns task ID)
- [migrate](migrate.md) - Migrate VMs (returns task ID)
- [clone](clone.md) - Clone VMs (returns task ID)
