# config

Configure pvecli settings and manage configuration file.

## Usage

```bash
# Show current configuration
pvecli config show

# Set configuration values
pvecli config set <key> <value>

# Initialize configuration
pvecli config init
```

## Configuration File

Location: `~/.pvecli/config.yaml`

### Example Configuration

```yaml
api_url: "https://proxmox.example.com:8006/api2/json"
token_id: "root@pam!cli"
token_secret: "your-secret-here"
insecure_skip_verify: false
output_format: "json"
```

## Configuration Options

| Option | Type | Description | Default |
|--------|------|-------------|---------|
| `api_url` | string | Proxmox API URL | Required |
| `token_id` | string | API token ID | Required |
| `token_secret` | string | API token secret | Required |
| `insecure_skip_verify` | bool | Skip TLS verification | `false` |
| `output_format` | string | Output format | `json` |

## Output Formats

Available formats:
- `json` - JSON format (default, machine-readable)
- `yaml` - YAML format (human-readable)
- `text` - Plain text format
- `table` - Table format (for list command)

## Commands

### Show Configuration

```bash
pvecli config show
```

Output:
```yaml
api_url: https://192.168.99.4:8006/api2/json
token_id: root@pam!pvecli
insecure_skip_verify: true
output_format: json
```

### Set Configuration Value

```bash
# Set output format
pvecli config set output_format yaml

# Set API URL
pvecli config set api_url "https://proxmox.example.com:8006/api2/json"

# Enable insecure mode
pvecli config set insecure_skip_verify true
```

### Initialize Configuration

```bash
pvecli config init
```

Creates `~/.pvecli/config.yaml` with default values and prompts for required settings.

## Examples

### Change output format to YAML
```bash
pvecli config set output_format yaml
pvecli list
```

### Enable insecure mode for self-signed certificates
```bash
pvecli config set insecure_skip_verify true
```

### View current configuration
```bash
cat ~/.pvecli/config.yaml
```

### Backup configuration
```bash
cp ~/.pvecli/config.yaml ~/.pvecli/config.yaml.backup
```

## Security Notes

- Configuration file contains sensitive credentials
- File permissions are automatically set to 600 (owner read/write only)
- Never commit config.yaml to version control
- Use environment variables for CI/CD pipelines

## Environment Variables

Override config file settings:

```bash
export PVECLI_API_URL="https://proxmox.example.com:8006/api2/json"
export PVECLI_TOKEN_ID="root@pam!cli"
export PVECLI_TOKEN_SECRET="your-secret"
```

## Required Permissions

No special permissions required - configuration is local.

## See Also

- [README](../README.md) - Getting started guide
- Example configuration: [config.example.yaml](../config.example.yaml)
