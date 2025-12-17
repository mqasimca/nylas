# Security

This project follows security best practices for handling credentials and sensitive data.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [TUI](TUI.md) | [Architecture](ARCHITECTURE.md) | [Development](DEVELOPMENT.md)

---

## Security Principles

- **No hardcoded credentials**: All API keys and secrets are stored in the system keyring
- **Comprehensive .gitignore**: Prevents accidental commit of sensitive files
- **Environment variables for testing**: Integration tests use environment variables for credentials
- **No credential files in repository**: The `.gitignore` blocks all common credential file patterns

## Credential Storage

Credentials are stored securely in your system keyring:

| Platform | Backend |
|----------|---------|
| Linux | Secret Service (GNOME Keyring, KWallet) |
| macOS | Keychain |
| Windows | Windows Credential Manager |

Config file location: `~/.config/nylas/config.yaml`

---

## Running Integration Tests

Integration tests require Nylas API credentials. Set them via environment variables:

```bash
# Required
export NYLAS_API_KEY="your-api-key"
export NYLAS_GRANT_ID="your-grant-id"

# Optional
export NYLAS_CLIENT_ID="your-client-id"

# Run integration tests
go test -tags=integration ./internal/cli/...

# Run with verbose output
go test -tags=integration -v ./internal/cli/...
```

## Destructive Test Operations

Some tests can modify data (send emails, delete messages). These require explicit opt-in:

```bash
# Enable send email tests
export NYLAS_TEST_SEND_EMAIL=true
export NYLAS_TEST_EMAIL="your-test-email@example.com"

# Enable delete message tests
export NYLAS_TEST_DELETE_MESSAGE=true
```

---

## Protected File Patterns

The `.gitignore` blocks these sensitive patterns:

### Environment Files
- `.env`, `.env.*`, `.env.local`
- `*.env`

### Credential Files
- `credentials.json`, `credentials.yaml`
- `*credentials*`, `*credential*`

### API Keys and Tokens
- `*.key`, `*.pem`, `*.p12`, `*.pfx`
- `api_key*`, `*api_key*`, `*token*`

### Secret Files
- `secrets.json`, `secrets.yaml`
- `*secrets*`, `*secret*`

### OAuth Tokens
- `oauth_token*`, `access_token*`, `refresh_token*`
- `*.token`

### SSH/GPG Keys
- `id_rsa*`, `id_dsa*`, `id_ecdsa*`, `id_ed25519*`
- `*.gpg`, `*.asc`, `secring.*`

---

## Best Practices

1. **Never commit credentials** - Always use environment variables or the system keyring
2. **Use `--yes` flag carefully** - It skips confirmation prompts for destructive operations
3. **Review before sending** - Always preview emails before sending
4. **Rotate API keys** - Regularly rotate your Nylas API keys
5. **Use test accounts** - Run destructive tests against test email accounts
