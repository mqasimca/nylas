# Troubleshooting

Quick diagnostics and solutions for common issues.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [FAQ](FAQ.md)

---

## Quick Diagnostics

```bash
nylas doctor    # Run health checks
```

This will check:
- API connectivity
- Authentication status
- Configuration validity
- System requirements

---

## Common Issues

| Issue | Quick Fix | Detailed Guide |
|-------|-----------|----------------|
| **Authentication fails** | `nylas auth config` | [Auth Guide](troubleshooting/auth.md) |
| **No emails shown** | Check grant ID | [Email Guide](troubleshooting/email.md) |
| **Calendar events missing** | Verify calendar access | [Email Guide](troubleshooting/email.md) |
| **Timezone incorrect** | Use `--timezone` flag | [Timezone Guide](troubleshooting/timezone.md) |
| **API errors (401, 403, 404)** | Check credentials | [API Guide](troubleshooting/api.md) |
| **API rate limits** | Reduce request frequency | [API Guide](troubleshooting/api.md#rate-limits) |
| **TUI not loading** | Check terminal size | [TUI Guide](TUI.md) |
| **Webhooks not working** | Verify URL is public | [Webhook Guide](troubleshooting/webhooks.md) |

---

## Quick Solutions

### Authentication Issues
```bash
# Reconfigure
nylas auth config

# Check current status
nylas auth status

# Re-login
nylas auth login
```
**→ [Detailed auth troubleshooting](troubleshooting/auth.md)**

### Email Issues
```bash
# Verify credentials
nylas email list --limit 1

# Try different grant
nylas email list <grant-id>
```
**→ [Detailed email troubleshooting](troubleshooting/email.md)**

### Timezone Issues
```bash
# Verify timezone exists
nylas timezone info America/New_York

# List available timezones
nylas timezone list --filter America
```
**→ [Detailed timezone troubleshooting](troubleshooting/timezone.md)**

### API Issues
```bash
# Test API connectivity
nylas doctor

# Check API key validity
nylas auth status
```
**→ [Detailed API troubleshooting](troubleshooting/api.md)**

---

## Detailed Guides

**Comprehensive troubleshooting with examples, edge cases, and solutions:**

- **[Authentication Issues](troubleshooting/auth.md)** - Login, credentials, OAuth
- **[Email Problems](troubleshooting/email.md)** - Missing emails, send failures
- **[API Errors](troubleshooting/api.md)** - Rate limits, permissions, connectivity
- **[Timezone Issues](troubleshooting/timezone.md)** - DST, conversions, parsing
- **[Complete FAQ](troubleshooting/faq.md)** - 50+ questions with detailed answers

---

## Still Need Help?

1. Check the [FAQ](FAQ.md)
2. Review [Command Documentation](COMMANDS.md)
3. Report issue: https://github.com/mqasimca/nylas/issues
