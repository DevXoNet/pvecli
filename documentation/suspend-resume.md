# suspend / resume

Suspend and resume VMs.

## Usage

```bash
pvecli suspend <vmid>
pvecli resume <vmid>
```

## Description

Suspend pauses a running VM by saving its state to disk. Resume restores the VM from the suspended state. This is similar to hibernation on a laptop.

## Commands

### suspend

Suspend (pause) a running VM.

```bash
pvecli suspend 230
```

**Output (JSON):**
```json
{
  "vmid": "230",
  "node": "node04",
  "action": "suspend",
  "status": "success"
}
```

### resume

Resume a suspended VM.

```bash
pvecli resume 230
```

**Output (JSON):**
```json
{
  "vmid": "230",
  "node": "node04",
  "action": "resume",
  "status": "success"
}
```

## Output Formats

### JSON Format (suspend)
```json
{
  "vmid": "230",
  "node": "node04",
  "action": "suspend",
  "status": "success"
}
```

### YAML Format (suspend)
```yaml
vmid: "230"
node: node04
action: suspend
status: success
```

### Text Format (suspend)
```
VM 230 suspended successfully
```

### Text Format (resume)
```
VM 230 resumed successfully
```

## Behavior

### Suspend
- Saves VM state to disk
- Frees RAM but keeps disk space
- VM appears as "paused" in Proxmox
- Faster than shutdown/start cycle
- Preserves exact VM state

### Resume
- Restores VM from saved state
- VM continues exactly where it left off
- All processes and connections resume
- Faster than cold boot

## Use Cases

1. **Resource management** - Free RAM temporarily
2. **Maintenance** - Pause VMs during host maintenance
3. **Quick recovery** - Resume VMs faster than cold boot
4. **Testing** - Pause VMs between test runs

## Limitations

- Only works for VMs, not containers
- Requires disk space for state file
- Some applications may not handle suspension well
- Network connections may timeout during suspension

## Differences from Other Operations

### Suspend vs Stop
- **Suspend** - Saves state, fast resume
- **Stop** - Full shutdown, cold boot required

### Suspend vs Snapshot
- **Suspend** - Temporary state save
- **Snapshot** - Permanent disk state backup

### Suspend vs Pause (QEMU)
- **Suspend** - Saves to disk, frees RAM
- **Pause** - Keeps in RAM, instant resume

## Examples

### Suspend VM for maintenance

```bash
# Suspend VM
pvecli suspend 230

# Perform maintenance
# ...

# Resume VM
pvecli resume 230
```

### Free resources temporarily

```bash
# Suspend multiple VMs
pvecli suspend 230
pvecli suspend 240
pvecli suspend 241

# Do resource-intensive work
# ...

# Resume VMs
pvecli resume 230
pvecli resume 240
pvecli resume 241
```

## Important Notes

1. **Disk space** - Suspended VMs require disk space equal to RAM size
2. **Time** - Long-suspended VMs may have stale time/date on resume
3. **Network** - Network connections will likely be dropped
4. **Applications** - Some applications may not handle suspension gracefully

## API Endpoints

- `POST /nodes/{node}/qemu/{vmid}/status/suspend` - Suspend VM
- `POST /nodes/{node}/qemu/{vmid}/status/resume` - Resume VM

## See Also

- [start](start.md) - Start VM
- [stop](stop.md) - Stop VM
- [reboot](reboot.md) - Reboot VM
- [snapshot](snapshot.md) - Create backup snapshot
