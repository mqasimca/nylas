# Documentation Index

Quick navigation guide to find the right documentation for your needs.

---

## 🎯 I want to...

### Get Started

- **Learn about Nylas CLI** → [README.md](../README.md)
- **Quick command reference** → [COMMANDS.md](COMMANDS.md)
- **See examples** → [EXAMPLES.md](EXAMPLES.md) or [examples/](examples/)

### Understand the Project

- **Architecture overview** → [ARCHITECTURE.md](ARCHITECTURE.md)
- **File structure** → [CLAUDE.md](../CLAUDE.md#file-structure)

### Development

- **Set up development environment** → [DEVELOPMENT.md](DEVELOPMENT.md)
- **Testing guidelines** → [.claude/rules/testing.md](../.claude/rules/testing.md)
- **Go quality & linting** → [.claude/rules/go-quality.md](../.claude/rules/go-quality.md)
- **Contributing guidelines** → [CONTRIBUTING.md](../CONTRIBUTING.md)

### Add Features

- **Add a CLI command** → [.claude/commands/add-command.md](../.claude/commands/add-command.md)
- **Add an API method** → [.claude/commands/add-api-method.md](../.claude/commands/add-api-method.md)
- **Add a domain type** → [.claude/commands/add-domain-type.md](../.claude/commands/add-domain-type.md)
- **Add a command flag** → [.claude/commands/add-flag.md](../.claude/commands/add-flag.md)
- **Generate CRUD command** → [.claude/commands/generate-crud-command.md](../.claude/commands/generate-crud-command.md)

### Testing

- **Run tests** → [.claude/commands/run-tests.md](../.claude/commands/run-tests.md)
- **Add integration test** → [.claude/commands/add-integration-test.md](../.claude/commands/add-integration-test.md)
- **Debug test failure** → [.claude/commands/debug-test-failure.md](../.claude/commands/debug-test-failure.md)
- **Analyze coverage** → [.claude/commands/analyze-coverage.md](../.claude/commands/analyze-coverage.md)
- **Testing guidelines** → [.claude/rules/testing.md](../.claude/rules/testing.md)

### Fix Issues

- **Fix build errors** → [.claude/commands/fix-build.md](../.claude/commands/fix-build.md)
- **Debug test failure** → [.claude/commands/debug-test-failure.md](../.claude/commands/debug-test-failure.md)
- **Troubleshooting guide** → [TROUBLESHOOTING.md](TROUBLESHOOTING.md) or [troubleshooting/](troubleshooting/)

### Quality & Security

- **Security scan** → [.claude/commands/security-scan.md](../.claude/commands/security-scan.md)
- **Security guidelines** → [SECURITY.md](SECURITY.md)
- **Code review** → [.claude/commands/review-pr.md](../.claude/commands/review-pr.md)
- **Go quality & linting** → [.claude/rules/go-quality.md](../.claude/rules/go-quality.md)
- **File size limits** → [.claude/rules/file-size-limits.md](../.claude/rules/file-size-limits.md)

### Maintenance

- **Update documentation** → [.claude/commands/update-docs.md](../.claude/commands/update-docs.md)
- **Documentation rules** → [.claude/rules/documentation-maintenance.md](../.claude/rules/documentation-maintenance.md)
- **Go quality rules** → [.claude/rules/go-quality.md](../.claude/rules/go-quality.md)

### Specific Features

- **AI features** → [AI.md](AI.md) or [ai/](ai/)
- **Timezone handling** → [TIMEZONE.md](TIMEZONE.md) or [timezone/](timezone/)
- **Calendar commands** → [commands/calendar.md](commands/calendar.md)
- **Email commands** → [commands/email.md](commands/email.md)
- **Slack integration** → [COMMANDS.md#slack-integration](COMMANDS.md#slack-integration)
- **Webhooks** → [WEBHOOKS.md](WEBHOOKS.md) or [commands/webhooks.md](commands/webhooks.md)
- **TUI (Terminal UI)** → [TUI.md](TUI.md)

---

## 📂 Documentation Structure

```
docs/
├── *.md                    # Main documentation (auto-loaded by Claude)
├── INDEX.md               # This file
├── ai/                    # AI feature details (load on-demand)
├── commands/              # Detailed command guides (load on-demand)
├── development/           # Development guides (load on-demand)
├── examples/              # Usage examples (load on-demand)
├── timezone/              # Detailed timezone docs (load on-demand)
└── troubleshooting/       # Debug guides (load on-demand)

.claude/
├── commands/              # 20 actionable skills
├── rules/                 # 6 development rules (auto-loaded)
├── agents/                # 6 specialized agents
└── hooks/                 # 6 automation hooks
```

---

## 🔍 By Role

### **New Contributors**
1. [README.md](../README.md) - Project overview
2. [CONTRIBUTING.md](../CONTRIBUTING.md) - How to contribute
3. [DEVELOPMENT.md](DEVELOPMENT.md) - Setup instructions
4. [CLAUDE.md](../CLAUDE.md#file-structure) - Code navigation

### **Developers Adding Features**
1. [ARCHITECTURE.md](ARCHITECTURE.md) - Understand the design
2. [.claude/commands/add-command.md](../.claude/commands/add-command.md) - Add CLI commands
3. [.claude/rules/testing.md](../.claude/rules/testing.md) - Testing requirements
4. [.claude/rules/documentation-maintenance.md](../.claude/rules/documentation-maintenance.md) - Doc updates

### **Bug Fixers**
1. [.claude/commands/debug-test-failure.md](../.claude/commands/debug-test-failure.md) - Test debugging
2. [TROUBLESHOOTING.md](TROUBLESHOOTING.md) - Common issues
3. [.claude/commands/fix-build.md](../.claude/commands/fix-build.md) - Fix build errors

### **Maintainers**
1. [.claude/commands/security-scan.md](../.claude/commands/security-scan.md) - Security checks
2. [.claude/commands/review-pr.md](../.claude/commands/review-pr.md) - PR review
3. [.claude/rules/go-quality.md](../.claude/rules/go-quality.md) - Go quality & linting

### **Users**
1. [README.md](../README.md) - Getting started
2. [COMMANDS.md](COMMANDS.md) - Command reference
3. [EXAMPLES.md](EXAMPLES.md) - Usage examples
4. [FAQ.md](FAQ.md) - Common questions

---

## 💡 Quick Tips

- **For AI (Claude):** Most docs are in CLAUDE.md and .claude/ directory
- **For humans:** Start with README.md and COMMANDS.md
- **Need help?** Check FAQ.md or TROUBLESHOOTING.md
- **Adding code?** Follow workflows in .claude/commands/
- **Security concern?** See SECURITY.md

---

**Last Updated:** December 30, 2024
