# Command Reference

Technical documentation for all `pvecli` commands.

## Commands

| Command | Description |
|---------|-------------|
| [backup](backup.md) | Create and manage backups |
| [clone](clone.md) | Clone VM templates (full/linked) |
| [cluster](cluster.md) | Cluster operations |
| [cluster update](cluster-update.md) | Update all nodes ⚠️ |
| [config](config.md) | Configure pvecli |
| [console](console.md) | Interactive console ⚠️ |
| [disk](disk.md) | Disk management operations |
| [info](info.md) | Show VM/container configuration |
| [list](list.md) | List all VMs and containers |
| [migrate](migrate.md) | Migrate VM/container to another node |
| [reboot](reboot.md) | Reboot a VM/container |
| [shutdown](shutdown.md) | Shutdown a VM/container |
| [snapshot](snapshot.md) | Create and manage snapshots |
| [start](start.md) | Start a VM/container |
| [status](status.md) | Check VM/container status |
| [stop](stop.md) | Stop a VM/container |
| [suspend](suspend-resume.md) | Suspend and resume VMs |
| [task](task.md) | Monitor and manage tasks |
| [template](template.md) | Convert VM to template |
| [top](top.md) | Real-time monitoring with graphs |

⚠️ = Known limitations with API tokens

## Authentication

### Single Cluster (Legacy)

Configure API token in `~/.devxo/pve.yaml`:

```yaml
api_url: "https://proxmox.example.com:8006/api2/json"
token_id: "root@pam!cli"
token_secret: "your-secret-here"
```

### Multiple Clusters

Configure multiple environments in `~/.devxo/pve.yaml`:

```yaml
environments:
  prod:
    api_url: "https://pve-prod.example.com:8006/api2/json"
    token_id: "root@pam!prod-token"
    token_secret: "prod-secret"
  dev:
    api_url: "https://pve-dev.example.com:8006/api2/json"
    token_id: "root@pam!dev-token"
    token_secret: "dev-secret"
default_env: prod
```

Use `--env` or `-e` flag to select environment:

```bash
pvecli -e prod list
pvecli -e dev start 100
```

See [config.md](config.md) for detailed multi-cluster setup.

## Output Formats

- `json` - Default, machine-readable
- `yaml` - Human-readable
- `text` - Simple key-value pairs
- `table` - Formatted table (use `--table` flag)

## Permissions

| Permission | Required For |
|------------|--------------|
| `VM.Audit` | list, info, status |
| `VM.PowerMgmt` | start, stop |
| `Sys.Audit` | cluster |

---

**Note**: Console command has known limitations with API tokens. See [console.md](console.md) for details.
