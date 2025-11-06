# shutdown

Shutdown a VM or container gracefully.

## Usage

```bash
pvecli shutdown <vmid> [flags]
```

## Description

Shuts down a running VM or container gracefully. The command sends a shutdown signal to the guest OS and waits for it to power off. Use --force flag for immediate hard stop.

## Flags

- `--force` - Force shutdown (hard stop, equivalent to pulling power)

## Examples

### Graceful shutdown

```bash
pvecli shutdown 230
```

**Output (JSON):**
```json
{
  "vmid": "230",
  "node": "node04",
  "type": "vm",
  "action": "shutdown",
  "force": false,
  "status": "success"
}
```

### Force shutdown

```bash
pvecli shutdown 230 --force
```

**Output (JSON):**
```json
{
  "vmid": "230",
  "node": "node04",
  "type": "vm",
  "action": "shutdown",
  "force": true,
  "status": "success"
}
```

## Output Formats

### JSON Format
```json
{
  "vmid": "230",
  "node": "node04",
  "type": "vm",
  "action": "shutdown",
  "force": false,
  "status": "success"
}
```

### YAML Format
```yaml
vmid: "230"
node: node04
type: vm
action: shutdown
force: false
status: success
```

### Text Format
```
Instance 230 shutdown gracefully
```

Or with --force:
```
Instance 230 forcefully shutdown
```

## Behavior

### Graceful Shutdown (default)
- Sends ACPI shutdown signal to guest OS
- Guest OS performs clean shutdown
- Waits for guest to power off
- Recommended for production use

### Force Shutdown (--force)
- Immediately stops the VM/container
- Equivalent to pulling the power cord
- May cause data corruption
- Use only when graceful shutdown fails

## Use Cases

1. **Maintenance** - Shutdown before hardware maintenance
2. **Resource management** - Free up resources
3. **Emergency** - Force shutdown unresponsive instances
4. **Automation** - Scheduled shutdowns

## API Endpoints

- `POST /nodes/{node}/qemu/{vmid}/status/shutdown` - Shutdown VM
- `POST /nodes/{node}/lxc/{vmid}/status/shutdown` - Shutdown container

## See Also

- [reboot](reboot.md) - Reboot instance
- [start](start.md) - Start instance
- [stop](stop.md) - Stop instance (alias for force shutdown)
