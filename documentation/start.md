# start

Start a stopped VM or container.

## Usage

```bash
# Start VM/container
pvecli start <vmid>

# Examples
pvecli start 230
pvecli start 500
```

## Output

```
Starting VM 230...
VM 230 started successfully
```

## API Endpoints

```
POST /api2/json/nodes/{node}/qemu/{vmid}/status/start
POST /api2/json/nodes/{node}/lxc/{vmid}/status/start
```

## Features

- Automatic node detection
- Works with both VMs and containers
- Confirms successful start

## Examples

### Start multiple VMs
```bash
for vmid in 230 500 502; do
  pvecli start $vmid
done
```

### Start and wait for running status
```bash
pvecli start 230 && sleep 5 && pvecli status 230
```

### Conditional start
```bash
if pvecli status 230 | grep -q "stopped"; then
  pvecli start 230
fi
```

## Error Handling

If VM/container is already running:
```
Error: VM is already running
```

If VM/container doesn't exist:
```
Error: VM 999 not found
```

## Required Permissions

- `VM.PowerMgmt` - Start/stop VMs and containers

## See Also

- [stop](stop.md) - Stop VM/container
- [status](status.md) - Check status
- [info](info.md) - Detailed information
