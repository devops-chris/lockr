# lockr

[![Release](https://img.shields.io/github/v/release/devops-chris/lockr)](https://github.com/devops-chris/lockr/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/devops-chris/lockr)](https://goreportcard.com/report/github.com/devops-chris/lockr)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A simple, elegant CLI for managing secrets in AWS SSM Parameter Store.

## Features

- **Zero config** - Works out of the box with AWS managed KMS
- **Interactive fuzzy search** - Find secrets instantly with `lockr list`
- **Secure input** - Values never appear in shell history
- **Pretty output** - Tables, spinners, and colors
- **Scriptable** - Quiet mode, JSON output, and proper exit codes
- **Tagging support** - Organize secrets with metadata
- **Path templating** - Configure prefix and environment patterns
- **Cross-platform** - Works on macOS, Linux, and Windows

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap devops-chris/tap
brew install lockr
```

### Scoop (Windows)

```powershell
scoop bucket add devops-chris https://github.com/devops-chris/scoop-bucket
scoop install lockr
```

### Go Install

```bash
go install github.com/devops-chris/lockr@latest
```

### Download Binary

Download the appropriate binary for your platform from [Releases](https://github.com/devops-chris/lockr/releases):

| Platform | Architecture | Download |
|----------|--------------|----------|
| macOS | Intel | `lockr_*_darwin_amd64.tar.gz` |
| macOS | Apple Silicon | `lockr_*_darwin_arm64.tar.gz` |
| Linux | x86_64 | `lockr_*_linux_amd64.tar.gz` |
| Linux | ARM64 | `lockr_*_linux_arm64.tar.gz` |
| Windows | x86_64 | `lockr_*_windows_amd64.zip` |
| Windows | ARM64 | `lockr_*_windows_arm64.zip` |

**Windows manual install:**
1. Download the `.zip` file
2. Extract `lockr.exe`
3. Add to your PATH or move to a directory in your PATH

**macOS/Linux manual install:**
```bash
# Example for macOS ARM64
curl -LO https://github.com/devops-chris/lockr/releases/latest/download/lockr_*_darwin_arm64.tar.gz
tar xzf lockr_*_darwin_arm64.tar.gz
sudo mv lockr /usr/local/bin/
```

## Quick Start

```bash
# Write a secret (prompts for value securely)
lockr write /myapp/prod/api-key

# Read it back
lockr read /myapp/prod/api-key

# List all secrets (interactive fuzzy search)
lockr list

# Delete a secret
lockr delete /myapp/prod/old-key
```

## Usage

### Writing Secrets

```bash
# Interactive prompt (secure, recommended)
lockr write /myapp/prod/db-password

# From value flag
lockr write /myapp/prod/api-key --value "sk_live_xxx"

# From file (great for certs, keys, JSON)
lockr write /myapp/prod/tls-cert --file ./cert.pem

# From stdin (for piping)
cat cert.pem | lockr write /myapp/prod/tls-cert --value -

# With tags
lockr write /myapp/prod/api-key --tag owner=platform --tag env=prod
```

**Windows PowerShell:**
```powershell
# From file
lockr write /myapp/prod/tls-cert --file .\cert.pem

# From value
lockr write /myapp/prod/api-key --value "sk_live_xxx"
```

### Reading Secrets

```bash
# Interactive search, then read
lockr read

# Read specific secret
lockr read /myapp/prod/api-key

# Value only (for scripts)
lockr read /myapp/prod/api-key --quiet

# JSON output
lockr read /myapp/prod/api-key --output json
```

### Listing Secrets

```bash
# All secrets with fuzzy search (interactive)
lockr list

# Secrets at a path (table view)
lockr list /myapp/prod

# Recursive listing
lockr list /myapp --recursive

# Interactive mode on specific path
lockr list /myapp -i
```

### Deleting Secrets

```bash
# With confirmation
lockr delete /myapp/prod/old-key

# Skip confirmation (for scripts)
lockr delete /myapp/prod/old-key --force
```

## Configuration

**Works with zero config!** Customize only if needed.

### Precedence

```
CLI flags > Environment variables > Config file > Defaults
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LOCKR_PREFIX` | (none) | Path prefix for relative paths |
| `LOCKR_ENV` | (none) | Environment added to path (prod, staging, etc.) |
| `LOCKR_OUTPUT` | `text` | Output format: `text`, `json` |
| `LOCKR_KMS_KEY` | `alias/aws/ssm` | KMS key for encryption |
| `LOCKR_REGION` | (AWS default) | AWS region |

### Path Templating

Configure prefix and environment to simplify paths:

**macOS/Linux:**
```bash
export LOCKR_PREFIX=/infra/saas
export LOCKR_ENV=prod

lockr write datadog/api-key
# Creates: /infra/saas/prod/datadog/api-key
```

**Windows PowerShell:**
```powershell
$env:LOCKR_PREFIX = "/infra/saas"
$env:LOCKR_ENV = "prod"

lockr write datadog/api-key
# Creates: /infra/saas/prod/datadog/api-key
```

**Windows CMD:**
```cmd
set LOCKR_PREFIX=/infra/saas
set LOCKR_ENV=prod

lockr write datadog/api-key
```

Full paths still work and bypass prefix/env:
```bash
lockr write /other/path/key
```

### Config File (Optional)

**macOS/Linux:** `~/.config/lockr/config.yaml`  
**Windows:** `%USERPROFILE%\.config\lockr\config.yaml`

```yaml
prefix: /infra/saas
env: prod
output: text
kms_key: alias/aws/ssm
region: us-east-1
```

## Scripting & Automation

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (secret not found, permission denied, etc.) |

### Bash Examples

```bash
# Get value for use in scripts
API_KEY=$(lockr read /myapp/prod/api-key --quiet)

# Export as environment variable
export DB_PASSWORD=$(lockr read /myapp/prod/db-password -q)

# Check if secret exists
if lockr read /myapp/prod/key -q > /dev/null 2>&1; then
  echo "Secret exists"
fi

# Write from environment variable (safe)
lockr write /myapp/prod/api-key --value "$API_KEY"

# Pipe from another command
aws secretsmanager get-secret-value --secret-id foo --query SecretString --output text \
  | lockr write /myapp/prod/migrated --value -
```

### PowerShell Examples

```powershell
# Get value for use in scripts
$ApiKey = lockr read /myapp/prod/api-key --quiet

# Set environment variable
$env:DB_PASSWORD = lockr read /myapp/prod/db-password -q

# Check if secret exists
if (lockr read /myapp/prod/key -q 2>$null) {
    Write-Host "Secret exists"
}

# Write from variable
lockr write /myapp/prod/api-key --value $ApiKey
```

### CI/CD

**GitHub Actions:**
```yaml
- name: Deploy secret
  run: echo "${{ secrets.API_KEY }}" | lockr write /myapp/prod/api-key --value -
  env:
    AWS_REGION: us-east-1
```

**Azure DevOps:**
```yaml
- script: |
    echo $(API_KEY) | lockr write /myapp/prod/api-key --value -
  env:
    AWS_REGION: us-east-1
```

## IAM Permissions

Minimum required policy:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ssm:PutParameter",
        "ssm:GetParameter",
        "ssm:GetParametersByPath",
        "ssm:DeleteParameter",
        "ssm:ListTagsForResource",
        "ssm:AddTagsToResource"
      ],
      "Resource": "arn:aws:ssm:*:*:parameter/*"
    },
    {
      "Effect": "Allow",
      "Action": ["kms:Encrypt", "kms:Decrypt"],
      "Resource": "*",
      "Condition": {
        "StringEquals": {
          "kms:ViaService": "ssm.*.amazonaws.com"
        }
      }
    }
  ]
}
```

### Scoped Access

Restrict users to specific paths:

```json
{
  "Resource": "arn:aws:ssm:*:*:parameter/myteam/*"
}
```

When running `lockr list`, users only see secrets they have access to.

## Prerequisites

- **AWS credentials** configured via:
  - Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
  - AWS credentials file (`~/.aws/credentials`)
  - IAM role (EC2, ECS, Lambda)
  - AWS SSO (`aws sso login`)

## Roadmap

See [ROADMAP.md](ROADMAP.md) for planned features including:
- AWS Secrets Manager & Azure Key Vault support
- Team/SSO integration for shared accounts
- Secret rotation and history
- And more!

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT License - see [LICENSE](LICENSE) for details.
