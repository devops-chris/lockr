# Examples

This directory contains example configurations and scripts for using lockr.

## Files

| File | Description |
|------|-------------|
| `config.yaml` | Example configuration file for `~/.config/lockr/config.yaml` |
| `github-actions.yml` | Example GitHub Actions workflow for deploying secrets |
| `iam-policy.json` | Full IAM policy for lockr access |
| `iam-policy-scoped.json` | IAM policy scoped to a specific path/team |

## Scripts

| Script | Description |
|--------|-------------|
| `scripts/batch-import.sh` | Import multiple secrets from a directory of files |
| `scripts/export-env.sh` | Export secrets as environment variables |

## Usage

### Configuration

Copy the example config to your home directory:

```bash
mkdir -p ~/.config/lockr
cp config.yaml ~/.config/lockr/config.yaml
# Edit as needed
```

### IAM Policy

Apply the IAM policy to your users/roles:

```bash
# Full access
aws iam put-user-policy --user-name myuser --policy-name lockr-access \
    --policy-document file://iam-policy.json

# Scoped access (edit the policy first to set your paths)
aws iam put-user-policy --user-name myuser --policy-name lockr-access \
    --policy-document file://iam-policy-scoped.json
```

### Batch Import

```bash
# Make executable
chmod +x scripts/batch-import.sh

# Import all files from a directory
./scripts/batch-import.sh ./my-secrets /myapp/prod
```

