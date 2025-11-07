# config

Configure pvecli settings and manage multiple Proxmox cluster environments.

## Overview

The `config` command allows you to set up and manage multiple Proxmox cluster environments (prod, dev, staging, etc.) from a single configuration file. This enables seamless switching between different clusters using the `--env` flag.

## Usage

```bash
# Interactive configuration wizard (add/update environment)
pvecli config

# List all configured environments
pvecli config list
```

## Configuration File

**Location:** `~/.devxo/pve.yaml`

### Multi-Cluster Configuration Format

```yaml
# Multiple environments/clusters
environments:
  # Production cluster
  prod:
    api_url: "https://pve-prod.example.com:8006/api2/json"
    token_id: "root@pam!prod-token"
    token_secret: "your-prod-secret"
    insecure_skip_verify: false
  
  # Development cluster
  dev:
    api_url: "https://pve-dev.example.com:8006/api2/json"
    token_id: "root@pam!dev-token"
    token_secret: "your-dev-secret"
    insecure_skip_verify: true
  
  # Staging cluster
  staging:
    api_url: "https://pve-staging.example.com:8006/api2/json"
    token_id: "root@pam!staging-token"
    token_secret: "your-staging-secret"
    insecure_skip_verify: false

# Default environment (used when --env flag is not specified)
default_env: prod

# Global settings
output_format: json
debug: false
```

### Legacy Single-Cluster Format (Deprecated)

For backward compatibility, the old format is still supported:

```yaml
api_url: "https://proxmox.example.com:8006/api2/json"
token_id: "root@pam!cli"
token_secret: "your-secret-here"
insecure_skip_verify: false
output_format: "json"
```

This format is automatically converted to an environment named `default`.

## Configuration Options

### Global Options

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `environments` | map | Map of environment configurations | Required |
| `default_env` | string | Default environment name | First environment |
| `output_format` | string | Output format (json/yaml/text) | `json` |
| `debug` | bool | Enable debug mode | `false` |

### Per-Environment Options

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `api_url` | string | Proxmox API URL | Required |
| `token_id` | string | API token ID | Required |
| `token_secret` | string | API token secret | Required |
| `insecure_skip_verify` | bool | Skip TLS verification | `false` |

## Commands

### Interactive Configuration

Add or update an environment using the interactive wizard:

```bash
pvecli config
```

**Prompts:**
1. **Environment name** - Name for this cluster (e.g., `prod`, `dev`, `staging`)
2. **API URL** - Proxmox API endpoint
3. **Token ID** - API token identifier (format: `user@realm!tokenid`)
4. **Token Secret** - API token secret
5. **Skip SSL verification** - Whether to skip TLS certificate verification
6. **Set as default** - Whether to make this the default environment

**Example session:**
```
=== Proxmox CLI Configuration ===

Environment name (e.g., prod, dev, staging): prod
API URL (e.g. https://proxmox.example.com:8006/api2/json): https://pve-prod.example.com:8006/api2/json
Token ID (e.g. root@pam!mytoken): root@pam!prod-token
Token Secret: ********************************
Skip SSL verification (yes/no, default: no): no

Setting 'prod' as default environment.

✓ Configuration saved successfully to ~/.devxo/pve.yaml
✓ Environment 'prod' configured

Use: pvecli --env prod <command>
Or simply: pvecli <command> (default environment)
```

### List Environments

View all configured environments:

```bash
pvecli config list
```

**Output:**
```
=== Configured Environments ===

• prod (default)
  API URL: https://pve-prod.example.com:8006/api2/json
  Token ID: root@pam!prod-token

• dev
  API URL: https://pve-dev.example.com:8006/api2/json
  Token ID: root@pam!dev-token

• staging
  API URL: https://pve-staging.example.com:8006/api2/json
  Token ID: root@pam!staging-token

Default environment: prod
Output format: json
```

## Using Multiple Environments

### Default Environment

Commands without `--env` flag use the default environment:

```bash
# Uses default environment (prod)
pvecli list
pvecli start 100
pvecli cluster
```

### Specific Environment

Use `--env` or `-e` flag to target a specific environment:

```bash
# Production cluster
pvecli --env prod list
pvecli -e prod start 100

# Development cluster
pvecli --env dev list
pvecli -e dev stop 200

# Staging cluster
pvecli --env staging list --table
pvecli -e staging migrate 100 node02
```

## Examples

### Setting Up Multiple Clusters

```bash
# Configure production cluster
pvecli config
# Enter: prod, https://pve-prod.example.com:8006/api2/json, credentials...

# Configure development cluster
pvecli config
# Enter: dev, https://pve-dev.example.com:8006/api2/json, credentials...

# Configure staging cluster
pvecli config
# Enter: staging, https://pve-staging.example.com:8006/api2/json, credentials...

# List all environments
pvecli config list
```

### Working with Different Environments

```bash
# List VMs in production
pvecli -e prod list --table

# Start VM in development
pvecli -e dev start 100

# Check cluster status in staging
pvecli -e staging cluster

# Migrate VM in production
pvecli -e prod migrate 200 node02

# Create backup in development
pvecli -e dev backup create 100
```

### Updating an Existing Environment

```bash
# Run config command
pvecli config

# Enter existing environment name
Environment name: prod

# Confirm overwrite
Environment 'prod' already exists. Overwrite? (yes/no): yes

# Enter new credentials...
```

## Output Formats

Available formats (set in configuration):
- `json` - JSON format (default, machine-readable)
- `yaml` - YAML format (human-readable)
- `text` - Plain text format

Example:
```yaml
output_format: yaml
```

## Security Best Practices

### File Permissions
- Configuration file is automatically created with 600 permissions (owner read/write only)
- Contains sensitive API credentials
- Never commit `~/.devxo/pve.yaml` to version control

### Token Management
- Create separate API tokens for each environment
- Use least-privilege principle - grant only necessary permissions
- Rotate tokens regularly
- Use `insecure_skip_verify: false` in production

### Environment Isolation
- Use different tokens for prod/dev/staging
- Consider using different Proxmox users per environment
- Enable privilege separation for non-admin tokens

## Troubleshooting

### Configuration File Not Found

```bash
Error: read ~/.devxo/pve.yaml: no such file or directory
```

**Solution:** Run `pvecli config` to create initial configuration.

### Environment Not Found

```bash
Error: environment 'prod' not found in config
```

**Solution:** 
1. Check available environments: `pvecli config list`
2. Add missing environment: `pvecli config`

### Invalid Configuration Format

```bash
Error: parse yaml: yaml: unmarshal errors
```

**Solution:** 
1. Check YAML syntax in `~/.devxo/pve.yaml`
2. Refer to [config.example.yaml](../config.example.yaml) for correct format
3. Backup and recreate: `pvecli config`

### Permission Denied

```bash
Error: create config file: permission denied
```

**Solution:** Ensure `~/.devxo/` directory is writable by your user.

## Migration from Legacy Format

If you have an existing single-cluster configuration, it will be automatically converted to the new format with environment name `default`:

**Old format:**
```yaml
api_url: "https://proxmox.example.com:8006/api2/json"
token_id: "root@pam!cli"
token_secret: "secret"
```

**Automatically becomes:**
```yaml
environments:
  default:
    api_url: "https://proxmox.example.com:8006/api2/json"
    token_id: "root@pam!cli"
    token_secret: "secret"
default_env: default
```

## Required Permissions

No special Proxmox permissions required for configuration management - it's a local operation.

For API token creation, you need:
- Access to Proxmox Web UI
- Permission to create API tokens (typically admin/root)

## See Also

- [README](../README.md) - Getting started guide
- [config.example.yaml](../config.example.yaml) - Example configuration file
- [Multi-Cluster Guide](../MULTI_CLUSTER.md) - Detailed multi-cluster setup guide
