# template

Convert a VM to a template.

## Usage

```bash
pvecli template <vmid>
```

## Description

Converts a VM into a template. Templates are used as base images for cloning new VMs. This operation cannot be easily undone.

## Important Warning

Converting a VM to a template is a one-way operation. Once converted:
- The VM cannot be started
- The VM can only be cloned
- Reverting requires manual intervention in Proxmox

Only convert VMs that are specifically prepared as templates.

## Examples

### Convert VM to template

```bash
pvecli template 9000
```

**Output (JSON):**
```json
{
  "vmid": "9000",
  "node": "node04",
  "action": "template",
  "status": "success"
}
```

## Output Formats

### JSON Format
```json
{
  "vmid": "9000",
  "node": "node04",
  "action": "template",
  "status": "success"
}
```

### YAML Format
```yaml
vmid: "9000"
node: node04
action: template
status: success
```

### Text Format
```
VM 9000 converted to template successfully
```

## Template Preparation

Before converting a VM to a template, you should:

1. **Clean the VM**
   ```bash
   # Remove SSH host keys
   rm /etc/ssh/ssh_host_*
   
   # Clear machine ID
   truncate -s 0 /etc/machine-id
   
   # Clear bash history
   history -c
   ```

2. **Install cloud-init** (for cloud-init templates)
   ```bash
   apt-get install cloud-init  # Debian/Ubuntu
   yum install cloud-init      # RHEL/CentOS
   ```

3. **Remove user data**
   ```bash
   # Clear logs
   find /var/log -type f -exec truncate -s 0 {} \;
   
   # Remove temporary files
   rm -rf /tmp/*
   ```

4. **Shutdown the VM**
   ```bash
   pvecli shutdown 9000
   ```

5. **Convert to template**
   ```bash
   pvecli template 9000
   ```

## Use Cases

1. **Base images** - Create standardized VM templates
2. **Rapid deployment** - Clone VMs from templates quickly
3. **Consistency** - Ensure all VMs start from same base
4. **Cloud-init** - Use with cloud-init for automated configuration

## Limitations

- Only works for VMs, not containers
- Cannot be reversed easily
- Template VMs cannot be started
- Requires VM to be stopped first

## After Creating Template

Once you have a template, you can clone it:

```bash
pvecli clone 9000 --new-id 230 --name web-server-01
```

## Common Template Naming

Convention for template VMIDs:
- `9000-9999` - Templates
- `100-8999` - Regular VMs/containers

Example templates:
- `9000` - Ubuntu 22.04 cloud template
- `9001` - Debian 12 cloud template
- `9002` - Rocky Linux 9 cloud template
- `9003` - Windows Server 2022 template

## API Endpoints

- `POST /nodes/{node}/qemu/{vmid}/template` - Convert VM to template

## See Also

- [clone](clone.md) - Clone VMs from templates
- [info](info.md) - View template configuration
- [list](list.md) - List templates
