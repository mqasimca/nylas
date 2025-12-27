# Claude Code Configuration

This directory contains skills, workflows, rules, and agents for AI-assisted development with Claude Code.

---

## 📂 Directory Structure

```
.claude/
├── commands/              # 13 actionable skills (invokable workflows)
├── rules/                 # 4 development rules (auto-applied)
├── agents/                # 3 specialized agents
├── settings.json          # Security hooks & permissions
└── README.md              # This file
```

---

## 🛠️ Skills (13 Total)

Skills are actionable workflows that guide AI through common development tasks.

### Feature Development (5 skills)

| Skill | Purpose | Dependencies |
|-------|---------|--------------|
| `add-command` | New CLI command | → add-flag (optional) |
| `add-api-method` | Extend API client | - |
| `add-domain-type` | New domain models | - |
| `add-flag` | Add command flags | - |
| `generate-crud-command` | Auto-generate CRUD operations | → add-command |

### Testing (4 skills)

| Skill | Purpose | Dependencies |
|-------|---------|--------------|
| `run-tests` | Execute test suite | - |
| `add-integration-test` | Create integration tests | - |
| `debug-test-failure` | Debug failing tests | → run-tests |
| `analyze-coverage` | Coverage analysis | → run-tests |

### Quality Assurance (3 skills)

| Skill | Purpose | Dependencies |
|-------|---------|--------------|
| `fix-build` | Resolve build errors | - |
| `security-scan` | Security audit | - |
| `review-pr` | Code review checklist | - |

### Maintenance (1 skill)

| Skill | Purpose | Dependencies |
|-------|---------|--------------|
| `update-docs` | Documentation updates | - |

---

## 📋 Rules (4 Files)

Rules are automatically applied to all code changes.

| Rule | Purpose | Applies To |
|------|---------|-----------|
| `testing.md` | Testing requirements & patterns | All new code |
| `linting.md` | Mandatory linting workflow | All Go code |
| `go-best-practices.md` | Modern Go patterns (1.21+) | All Go code |
| `documentation-maintenance.md` | Doc update requirements | Code + doc changes |

---

---

## 🤖 Agents (3 Specialized)

Pre-configured agents for specific tasks.

| Agent | Purpose | Tools Available |
|-------|---------|-----------------|
| `code-reviewer` | Independent code review | Read, Grep, Glob, git diff/log |
| `security-auditor` | Deep security analysis | Read, Grep, Glob, git log |
| `test-writer` | Generate comprehensive tests | Read, Grep, Glob, Write |

---

## 🔒 Security (settings.json)

**Pre-commit Hooks:**
- Check for sensitive files (.env, .pem, .key)
- Scan for secrets (api_key, password, token)

**Post-commit Hooks:**
- Run security scan

**Permissions:**
- ✅ Allowed: go, golangci-lint, make, git (except push), gh CLI
- ❌ Denied: git push, destructive operations
- 🔐 Protected: .env, .pem/.key, secrets/, credentials

---

## 🎯 Common Workflows

### Add a New Command

```bash
# 1. Use add-command skill
add-command          # Create CLI command structure
add-flag             # Add flags (if needed)
add-integration-test # Add tests
update-docs          # Update COMMANDS.md
```

### Run Security Scan

```bash
# Use security-scan skill before commits
security-scan        # Check for secrets and vulnerabilities
```

---

## 📊 Skill Usage Statistics

**Most Used:**
1. `add-command` - New CLI commands
2. `run-tests` - Test execution
3. `review-pr` - Code review
4. `security-scan` - Security checks
5. `debug-test-failure` - Bug debugging

**Best for AI Collaboration:**
- `generate-crud-command` - Auto-generates complete CRUD operations
- `debug-test-failure` - Systematic debugging
- `security-scan` - Prevents common mistakes

---

## 💡 Tips for Using Skills

1. **Use granular skills**: Each skill handles a specific task
2. **Follow dependencies**: Skills may require others to complete
3. **Check rules**: All skills must comply with rules in `rules/`
4. **Security first**: Always run `security-scan` before commits
5. **Test everything**: Use `run-tests` and `add-integration-test` skills

---

## 🔗 Related Documentation

- **Main Guide:** [`CLAUDE.md`](../CLAUDE.md) - AI assistant quick reference
- **Doc Index:** [`docs/INDEX.md`](../docs/INDEX.md) - Documentation decision tree
- **Architecture:** [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) - System design
- **Commands:** [`docs/COMMANDS.md`](../docs/COMMANDS.md) - CLI reference

---

## 📈 Metrics

- **Total Skills:** 13
- **Total Rules:** 4
- **Total Agents:** 3
- **Coverage:** Full development lifecycle
- **Last Updated:** December 27, 2024

---

**For AI Assistants:** This directory contains your primary configuration and workflow definitions. Follow skills for structured tasks, apply rules to all changes, and use agents for specialized analysis.

**For Developers:** Use skills via `/skill-name` or reference workflows directly. The `.claude/settings.json` file enforces security policies automatically.
