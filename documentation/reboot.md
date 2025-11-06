# reboot

Reboot a VM or container.

## Usage

```bash
pvecli reboot <vmid>
```

## Description

Reboots a running VM or container. The command automatically detects which node the instance is running on.

## Examples

### Reboot a VM

```bash
pvecli reboot 230
```

**Output (JSON):**
```json
{
  "vmid": "230",
  "node": "node04",
  "type": "vm",
  "action": "reboot",
  "status": "success"
}
```

### Reboot a container

```bash
pvecli reboot 100
```

**Output (JSON):**
```json
{
  "vmid": "100",
  "node": "node01",
  "type": "ct",
  "action": "reboot",
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
  "action": "reboot",
  "status": "success"
}
```

### YAML Format
```yaml
vmid: "230"
node: node04
type: vm
action: reboot
status: success
```

### Text Format
```
Instance 230 rebooted successfully
```

## Behavior

- Sends a reboot signal to the guest OS
- Works for both VMs (QEMU) and containers (LXC)
- Requires guest to be running
- Node is automatically detected

## API Endpoints

- `POST /nodes/{node}/qemu/{vmid}/status/reboot` - Reboot VM
- `POST /nodes/{node}/lxc/{vmid}/status/reboot` - Reboot container

## See Also

- [shutdown](shutdown.md) - Graceful shutdown
- [start](start.md) - Start instance
- [stop](stop.md) - Stop instance
