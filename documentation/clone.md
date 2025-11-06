# clone

Clone a VM template to create a new VM. Default mode is Full Clone. Linked Clone is available only on the same node as the template.

## Usage

```bash
# Full clone (default - independent copy)
pvecli clone <template-vmid> --new-id <new-vmid> --name <vm-name>

# Linked clone (same node only, fast and space-efficient)
pvecli clone <template-vmid> --new-id <new-vmid> --name <vm-name> --full=false

# Clone and start immediately
pvecli clone <template-vmid> --new-id <new-vmid> --name <vm-name> --start
```

## Parameters

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--new-id` | int | Yes | New VM ID |
| `--name` | string | No | Name for the new VM |
| `--description` | string | No | Description for the new VM |
| `--pool` | string | No | Add to resource pool |
| `--storage` | string | No | Target storage (for full clone) |
| `--full` | bool | No | Full clone (default). Set `--full=false` for linked clone |
| `--start` | bool | No | Start VM after cloning |

## Clone Modes

### Linked Clone (Optional)

**Fast and space-efficient** - Creates a new VM that uses the template as a base.

**Advantages:**
- Very fast (seconds)
- Minimal disk space usage
- Instant deployment
- Perfect for testing and development

**Limitations:**
- Must be on same node as template
- Cannot use `--target-node` (use `--full` for cross-node cloning)

**How it works:**
- Creates a snapshot of the template
- New VM uses template disks as backing storage
- Only changes are stored in new VM

**Example (same node only):**
```bash
pvecli clone 9005 --new-id 107 --name web-server-01
```

### Full Clone

**Complete independent copy** - Creates a fully independent VM.

**Advantages:**
- ✅ Completely independent
- ✅ Can delete template after cloning
- ✅ Better for production
- ✅ No dependency on template

**Disadvantages:**
- ❌ Slower (copies all disks)
- ❌ Uses more disk space
- ❌ Takes longer to complete

**Example:**
```bash
pvecli clone 9005 --new-id 107 --name web-server-01
```

## Examples

### Basic Linked Clone (same node)
```bash
pvecli clone 9005 --new-id 107 --name web-server-01 --full=false
```

### Full Clone with Custom Storage (default)
```bash
pvecli clone 9005 --new-id 107 --name web-server-01 \
  --full \
  --storage local-lvm
```

### Cross-node cloning
Not supported by this command flow. Clone occurs on the template's node.

### Clone and Start Immediately
```bash
pvecli clone 9005 --new-id 107 --name web-server-01 --start
```

### Clone with Description and Pool
```bash
pvecli clone 9005 --new-id 107 \
  --name web-server-01 \
  --description "Production web server" \
  --pool production
```

### Batch Clone Multiple VMs
```bash
# Clone 3 web servers
for i in {1..3}; do
  pvecli clone 9005 --new-id $((100 + i)) --name web-server-0$i
done
```

## Output

### Clone Process
```
Looking for template VM 9005...

╔════════════════════════════════════════╗
║         Cloning VM Template            ║
╚════════════════════════════════════════╝

Template ID:   9005
New VM ID:     107
Name:          web-server-01
Source Node:   node04
Target Node:   node04
Clone Mode:    full

Cloning VM 9005 -> 107...
Clone task started: UPID:node04:00001234:...
Waiting for clone to complete...

✓ VM 9005 cloned successfully to VM 107
```

### Completion Summary
```
╔════════════════════════════════════════╗
║           Clone Complete               ║
╚════════════════════════════════════════╝

New VM ID: 107
Name:      web-server-01
Node:      node04
Mode:      full clone
Status:    stopped

To start the VM, run:
  pvecli start 107
```

## Workflow

### 1. Create Template

First, create a template VM:
```bash
# Install and configure a VM
# Then convert to template in Proxmox Web UI
# Or via API
```

### 2. Clone Template

Clone the template to create new VMs:
```bash
pvecli clone 9005 --new-id 107 --name web-server-01
```

### 3. Customize (Optional)

Modify the cloned VM if needed:
```bash
# Update configuration
# Change network settings
# Resize disks
```

### 4. Start VM

```bash
pvecli start 107
```

## Best Practices

### Use Linked Clones For:
- ✅ Development environments
- ✅ Testing
- ✅ Temporary VMs
- ✅ Quick deployments
- ✅ Lab environments

### Use Full Clones For:
- ✅ Production VMs
- ✅ Long-term deployments
- ✅ When template will be deleted
- ✅ Independent VMs
- ✅ Critical workloads

### Template Management
- Keep templates updated
- Use cloud-init for customization
- Document template configurations
- Version your templates (e.g., ubuntu-22.04-v1, v2, etc.)

### Naming Conventions
```bash
# Good naming
pvecli clone 9005 --new-id 107 --name web-prod-01
pvecli clone 9005 --new-id 108 --name web-prod-02

# Include environment
pvecli clone 9005 --new-id 201 --name web-dev-01
pvecli clone 9005 --new-id 301 --name web-test-01
```

## API Endpoints

```
GET  /api2/json/nodes/{node}/qemu/{vmid}/config
POST /api2/json/nodes/{node}/qemu/{vmid}/clone
GET  /api2/json/nodes/{node}/tasks/{taskid}/status
POST /api2/json/nodes/{node}/qemu/{vmid}/status/start
```

## Required Permissions

- `VM.Clone` - Clone VMs
- `VM.Allocate` - Create new VMs
- `Datastore.AllocateSpace` - Allocate disk space

## Comparison

| Feature | Linked Clone | Full Clone |
|---------|--------------|------------|
| Speed |  Fast (seconds) | Slow (minutes) |
| Disk Space |  Minimal |  Full copy |
| Independence |  Depends on template |  Fully independent |
| Template Deletion |  Cannot delete |  Can delete |
| Use Case | Dev/Test | Production |

## Troubleshooting

### Parameter verification failed
```
Error: clone failed: Parameter verification failed.
```
**Common causes:**
1. **Attempting cross-node clone**
   - Cross-node cloning is not supported in this command
   - Solution: Clone on the same node as the template, then migrate

2. **Missing required parameters**
   - Ensure `--new-id` is provided
   - VM ID must be unique

### VM ID already exists
```
Error: VM 107 already exists
```
Solution: Use a different `--new-id`

### Template not found
```
Error: template VM 9005 not found
```
Solution: Verify template ID with `pvecli list`

### Not a template
```
Error: VM 9005 is not a template
```
Solution: Convert VM to template first

### Linked clone across nodes
Linked clone requires same node. This command does not support cross-node linked clones.

### Insufficient storage
```
Error: not enough space on storage
```
Solution: 
- Use linked clone instead of full
- Specify different storage with `--storage`
- Free up space on target storage

### Clone task timeout
```
Error: task timeout after 300 seconds
```
Solution: Large VMs may take longer, this is normal for full clones

## See Also

- [list](list.md) - List templates
- [start](start.md) - Start cloned VM
- [info](info.md) - View VM configuration
- [stop](stop.md) - Stop VM
