# disk

Manage VM disks.

## Usage

```bash
pvecli disk resize <vmid> --disk <disk> --size <size>
```

## Description

The disk command provides tools to manage VM disks. Currently supports resizing disks. Note: Disk resize is only supported for VMs, not containers.

## Commands

### disk resize

Resize a VM disk. Can only increase size, not decrease.

```bash
pvecli disk resize 230 --disk scsi0 --size +10G
pvecli disk resize 230 --disk virtio0 --size 50G
```

**Flags:**
- `--disk` - Disk name (required, e.g., scsi0, virtio0, sata0)
- `--size` - New size (required, e.g., +10G for increase, 50G for absolute)

**Output (JSON):**
```json
{
  "vmid": "230",
  "node": "node04",
  "disk": "scsi0",
  "size": "+10G",
  "action": "resize",
  "status": "success"
}
```

## Disk Names

Common disk names:
- `scsi0`, `scsi1`, `scsi2`, ... - SCSI disks
- `virtio0`, `virtio1`, `virtio2`, ... - VirtIO disks
- `sata0`, `sata1`, `sata2`, ... - SATA disks
- `ide0`, `ide1`, `ide2`, `ide3` - IDE disks

## Size Format

Size can be specified in two ways:

### Relative (increase)
```bash
+10G   # Increase by 10 GB
+5T    # Increase by 5 TB
+512M  # Increase by 512 MB
```

### Absolute (set to specific size)
```bash
50G    # Set to 50 GB total
1T     # Set to 1 TB total
```

Units:
- `K` or `KB` - Kilobytes
- `M` or `MB` - Megabytes
- `G` or `GB` - Gigabytes
- `T` or `TB` - Terabytes

## Examples

### Increase disk by 10GB

```bash
pvecli disk resize 230 --disk scsi0 --size +10G
```

### Set disk to 100GB total

```bash
pvecli disk resize 230 --disk virtio0 --size 100G
```

### Resize multiple disks

```bash
pvecli disk resize 230 --disk scsi0 --size +10G
pvecli disk resize 230 --disk scsi1 --size +20G
```

## Output Formats

### JSON Format
```json
{
  "vmid": "230",
  "node": "node04",
  "disk": "scsi0",
  "size": "+10G",
  "action": "resize",
  "status": "success"
}
```

### YAML Format
```yaml
vmid: "230"
node: node04
disk: scsi0
size: +10G
action: resize
status: success
```

### Text Format
```
Disk scsi0 on VM 230 resized to +10G
```

## Important Notes

1. **Cannot decrease size** - Disk resize only supports increasing size
2. **VM only** - Containers (LXC) do not support disk resize via this method
3. **Guest OS** - After resize, you may need to resize the partition/filesystem inside the guest OS
4. **Online resize** - Some guest OSes support online resize, others require reboot

## Guest OS Steps

After resizing the disk, you typically need to:

### Linux
```bash
# Rescan disk
echo 1 > /sys/class/block/sda/device/rescan

# Resize partition (if using partitions)
growpart /dev/sda 1

# Resize filesystem
resize2fs /dev/sda1  # ext4
xfs_growfs /         # xfs
```

### Windows
1. Open Disk Management
2. Right-click the volume
3. Select "Extend Volume"
4. Follow the wizard

## Use Cases

1. **Running out of space** - Increase disk when VM needs more storage
2. **Database growth** - Expand database volumes
3. **Log files** - Increase space for growing log files
4. **Application data** - Expand application storage

## Limitations

- Cannot shrink disks
- Only works for VMs, not containers
- Requires sufficient space on underlying storage
- May require guest OS actions to use new space

## API Endpoints

- `PUT /nodes/{node}/qemu/{vmid}/resize` - Resize VM disk

## See Also

- [info](info.md) - View VM disk configuration
- [status](status.md) - Check VM status
