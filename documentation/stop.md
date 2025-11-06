# stop

Stop a running VM or container.

## Usage

```bash
# Stop VM/container
pvecli stop <vmid>

# Examples
pvecli stop 230
pvecli stop 500
```

## Output

```
Stopping VM 230...
VM 230 stopped successfully
```

## API Endpoints

```
POST /api2/json/nodes/{node}/qemu/{vmid}/status/stop
POST /api2/json/nodes/{node}/lxc/{vmid}/status/stop
```

## Features

- Automatic node detection
- Works with both VMs and containers
- Graceful shutdown (ACPI shutdown for VMs)
- Confirms successful stop

## Examples

### Stop multiple VMs
```bash
for vmid in 230 500 502; do
  pvecli stop $vmid
done
```

### Stop and verify
```bash
pvecli stop 230 && sleep 5 && pvecli status 230
```

### Conditional stop
```bash
if pvecli status 230 | grep -q "running"; then
  pvecli stop 230
fi
```

## Error Handling

If VM/container is already stopped:
```
Error: VM is already stopped
```

If VM/container doesn't exist:
```
Error: VM 999 not found
```

## Notes

- For VMs: Sends ACPI shutdown signal (graceful)
- For containers: Stops the container gracefully
- VM/container may take time to fully stop
- Use `status` command to verify shutdown

## Required Permissions

- `VM.PowerMgmt` - Start/stop VMs and containers

## See Also

- [start](start.md) - Start VM/container
- [status](status.md) - Check status
- [info](info.md) - Detailed information
