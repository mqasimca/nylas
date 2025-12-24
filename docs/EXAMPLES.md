# Usage Examples

Real-world workflows and automation examples.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [FAQ](FAQ.md)

---

## Quick Examples

### Email Automation
```bash
# Check for urgent emails
nylas email list --unread | grep -i "urgent"

# Send scheduled reminder
nylas email send --to "team@example.com" --subject "Meeting in 1 hour" --schedule 1h
```
**→ [Full email workflows](examples/email-workflows.md)**

### Team Scheduling
```bash
# Find meeting time across timezones
nylas calendar find-time --participants "alice@team.com,bob@team.com" --duration 1h

# Check team availability
nylas calendar availability check
```
**→ [Full scheduling guide](examples/scheduling.md)**

### Webhook Integration
```bash
# Create webhook for new emails
nylas webhook create --url "https://myapp.com/webhook" --triggers "message.created"

# Test webhook
nylas webhook test <webhook-id>
```
**→ [Full webhook guide](examples/webhooks.md)**

### Timezone Tools
```bash
# Convert meeting time to colleague's timezone
nylas timezone convert --from "America/New_York" --to "Europe/London"

# Find best meeting time across 3 timezones
nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"
```
**→ [Timezone examples](commands/timezone.md)**

---

## Common Workflows

| Workflow | Command | Full Guide |
|----------|---------|------------|
| **Daily email check** | `nylas email list --unread` | [Email workflows](examples/email-workflows.md#daily-check) |
| **Send bulk emails** | `nylas email send --to "..."` | [Email workflows](examples/email-workflows.md#bulk-send) |
| **Team scheduling** | `nylas calendar find-time` | [Scheduling](examples/scheduling.md#team-meetings) |
| **Multi-timezone events** | `nylas calendar events list --timezone` | [Scheduling](examples/scheduling.md#timezones) |
| **Webhook automation** | `nylas webhook create` | [Webhooks](examples/webhooks.md) |
| **OTP extraction** | `nylas otp get` | [Automation](examples/automation.md#otp) |

---

## Detailed Examples

**Comprehensive guides with code, explanations, and best practices:**

- **[Email Workflows](examples/email-workflows.md)** - Daily automation, filtering, scheduling
- **[Scheduling Guide](examples/scheduling.md)** - Team meetings, timezone coordination
- **[Webhook Integration](examples/webhooks.md)** - Setup, testing, event handling
- **[Advanced Automation](examples/automation.md)** - Scripts, pipelines, integrations

---

## Interactive Mode

```bash
# Launch TUI for visual workflow
nylas tui

# Demo mode (no credentials needed)
nylas tui --demo
```

See [TUI Documentation](TUI.md) for keyboard shortcuts and features.

---

## More Resources

- **Commands:** [COMMANDS.md](COMMANDS.md)
- **Troubleshooting:** [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- **FAQ:** [FAQ.md](FAQ.md)
