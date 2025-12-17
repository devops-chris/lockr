# Roadmap

Future enhancements and feature ideas for lockr.

## Planned Features

### Team/SSO Integration
- Auto-detect team from AWS SSO session tags
- `LOCKR_TEAM` environment variable for shared-account organizations
- Path template support: `/{team}/{env}/{service}/{key}`
- IAM policy generation based on team membership

### Additional Backends
- **AWS Secrets Manager** - Support for Secrets Manager alongside SSM Parameter Store
- **Azure Key Vault** - Cross-cloud support for Azure environments
- **HashiCorp Vault** - Integration with Vault for hybrid setups

### Enhanced Features
- **Secret rotation** - Built-in support for rotating secrets
- **Diff command** - Compare secret values between environments
- **Copy command** - Copy secrets between paths or environments
- **History** - View version history of a secret
- **Bulk operations** - Import/export secrets from JSON/YAML files
- **Secret templates** - Generate secrets from templates (e.g., connection strings)

### Developer Experience
- **Shell completions** - Enhanced tab completion for paths
- **TUI mode** - Full terminal UI for browsing and managing secrets
- **VS Code extension** - View and manage secrets from the IDE
- **Secret linting** - Detect accidentally committed secrets

### Enterprise Features
- **Audit logging** - Local audit log of all operations
- **Policy enforcement** - Enforce naming conventions and required tags
- **Secret expiration** - Warn about secrets that haven't been rotated

## Contributing

Have an idea? Open an issue or discussion on GitHub! Pull requests welcome.

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

