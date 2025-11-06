# status

Quick status check for a VM or container.

## Usage

```bash
# Check VM/container status
pvecli status <vmid>

# Examples
pvecli status 230
pvecli status 500
```

## Output

Simple status output:

```
VM 230: k8s-master - running
```

Possible statuses:
- `running` - VM/container is running
- `stopped` - VM/container is stopped
- `paused` - VM is paused

## API Endpoints

```
GET /api2/json/nodes/{node}/qemu/{vmid}/status/current
GET /api2/json/nodes/{node}/lxc/{vmid}/status/current
```

## Features

- Fast status check
- Automatic node detection
- Works with both VMs and containers
- Simple output format

## Examples

### Check multiple VMs
```bash
for vmid in 230 500 502; do
  pvecli status $vmid
done
```

### Use in scripts
```bash
if pvecli status 230 | grep -q "running"; then
  echo "VM is running"
else
  echo "VM is not running"
fi
```

### Get just the status
```bash
pvecli status 230 | awk '{print $NF}'
```

## Required Permissions

- `VM.Audit` - View VM/container status

## See Also

- [info](info.md) - Detailed information
- [start](start.md) - Start VM/container
- [stop](stop.md) - Stop VM/container
