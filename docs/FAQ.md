# Frequently Asked Questions

Quick answers to common questions. For detailed solutions, see the troubleshooting guides.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [Troubleshooting](TROUBLESHOOTING.md)

---

## Quick Answers

**Q: How do I get started?**
See [Quick Start](../README.md#quick-start)

**Q: How do I authenticate?**
```bash
nylas auth config    # Configure API credentials
nylas auth login     # Login with email provider
```

**Q: Authentication not working?**
Run diagnostics: `nylas doctor` → See [Auth troubleshooting](troubleshooting/auth.md)

**Q: No emails showing up?**
Check your grant ID is correct → See [Email troubleshooting](troubleshooting/email.md)

**Q: Timezone conversion issues?**
See [Timezone troubleshooting](troubleshooting/timezone.md)

**Q: API rate limit errors?**
See [API troubleshooting](troubleshooting/api.md#rate-limits)

**Q: How do I use the TUI?**
```bash
nylas tui           # Launch interactive interface
nylas tui --demo    # Demo mode (no credentials)
```
See [TUI Documentation](TUI.md)

**Q: How do I schedule emails?**
```bash
nylas email send --to "user@example.com" --subject "Hello" --schedule 2h
```
See [Email examples](examples/email-workflows.md)

**Q: How do I find meeting times across timezones?**
```bash
nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"
```
See [Scheduling examples](examples/scheduling.md)

**Q: Can I use this offline?**
Yes! Timezone utilities work offline: `nylas timezone --help`

**Q: How do I set up webhooks?**
See [Webhook examples](examples/webhooks.md)

**Q: Where are my credentials stored?**
Securely in your system keyring (Keychain/GNOME Keyring/Windows Credential Manager)

---

## Full FAQ

**For comprehensive answers with examples and edge cases:**
→ [Complete FAQ Guide](troubleshooting/faq.md)

---

## More Help

- **Troubleshooting:** [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- **Examples:** [EXAMPLES.md](EXAMPLES.md)
- **Commands:** [COMMANDS.md](COMMANDS.md)
- **Report Issues:** https://github.com/mqasimca/nylas/issues
