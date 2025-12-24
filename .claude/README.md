# Claude Code Configuration

This directory contains skills, workflows, rules, and agents for AI-assisted development with Claude Code.

---

## 📂 Directory Structure

```
.claude/
├── commands/              # 22 actionable skills (invokable workflows)
├── rules/                 # 4 development rules (auto-applied)
├── workflows/             # 2 multi-step workflows
├── agents/                # 3 specialized agents
├── settings.json          # Security hooks & permissions
└── README.md              # This file
```

---

## 🛠️ Skills (22 Total)

Skills are actionable workflows that guide AI through common development tasks.

### Feature Development (6 skills)

| Skill | Purpose | Dependencies |
|-------|---------|--------------|
| `add-feature` | Complete feature implementation | → add-domain-type<br>→ add-api-method<br>→ add-command<br>→ add-integration-test<br>→ update-docs |
| `add-command` | New CLI command | → add-flag (optional) |
| `add-api-method` | Extend API client | - |
| `add-domain-type` | New domain models | - |
| `add-flag` | Add command flags | - |
| `generate-crud-command` | Auto-generate CRUD operations | → add-command |

**Dependency Graph:**
```
add-feature
├─ add-domain-type
├─ add-api-method
├─ add-command
│  └─ add-flag (optional)
├─ add-integration-test
└─ update-docs
```

### Testing (4 skills)

| Skill | Purpose | Dependencies |
|-------|---------|--------------|
| `run-tests` | Execute test suite | - |
| `add-integration-test` | Create integration tests | - |
| `debug-test-failure` | Debug failing tests | → run-tests |
| `analyze-coverage` | Coverage analysis | → run-tests |

### Quality Assurance (4 skills)

| Skill | Purpose | Dependencies |
|-------|---------|--------------|
| `fix-bug` | Bug fix workflow | → debug-test-failure (optional)<br>→ add-integration-test |
| `fix-build` | Resolve build errors | - |
| `security-scan` | Security audit | - |
| `review-pr` | Code review checklist | - |

### Maintenance (5 skills)

| Skill | Purpose | Dependencies |
|-------|---------|--------------|
| `smart-commit` | Commit workflow | → security-scan |
| `update-docs` | Documentation updates | - |
| `go-modernize` | Upgrade to modern Go | → fix-build<br>→ run-tests |
| `sync-mock-implementations` | Keep mocks synchronized | - |
| `validate-api-signatures` | API contract validation | - |

### Specialized (3 skills)

| Skill | Purpose | Dependencies |
|-------|---------|--------------|
| `add-domain-type` | Domain model creation | - |
| `add-api-method` | API client extension | - |
| `add-flag` | Command flag addition | - |

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

## 🔄 Workflows (2 Files)

Multi-step workflows for complex tasks.

| Workflow | Purpose |
|----------|---------|
| `testing.md` | Comprehensive testing guide |
| `code-quality-checklist.md` | Quality gates before commit |

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

### Add a Complete Feature

```bash
# 1. Use add-feature skill
# This runs in sequence:
add-domain-type      # Create domain models
add-api-method       # Implement API client
add-command          # Create CLI commands
add-integration-test # Add tests
update-docs          # Update documentation
```

### Fix a Bug

```bash
# 1. Use fix-bug skill
debug-test-failure   # Understand the issue (if test exists)
# 2. Fix the code
add-integration-test # Add regression test
update-docs          # Update docs if behavior changed
```

### Add a New Command

```bash
# 1. Use add-command skill
add-command          # Create CLI command structure
add-flag             # Add flags (if needed)
add-integration-test # Add tests
update-docs          # Update COMMANDS.md
```

### Commit Changes

```bash
# 1. Use smart-commit skill
# This runs:
security-scan        # Check for secrets
# 2. Create commit with message
# 3. Verify success
```

---

## 📊 Skill Usage Statistics

**Most Used:**
1. `add-feature` - Complete feature development
2. `fix-bug` - Bug resolution
3. `smart-commit` - Safe commits
4. `run-tests` - Test execution
5. `review-pr` - Code review

**Best for AI Collaboration:**
- `add-feature` - Comprehensive, step-by-step
- `debug-test-failure` - Systematic debugging
- `security-scan` - Prevents common mistakes

---

## 💡 Tips for Using Skills

1. **Start with high-level skills**: Use `add-feature` instead of individual skills
2. **Follow dependencies**: Skills may require others to complete
3. **Check rules**: All skills must comply with rules in `rules/`
4. **Security first**: `smart-commit` always runs `security-scan`
5. **Test everything**: Most skills include testing steps

---

## 🔗 Related Documentation

- **Main Guide:** [`CLAUDE.md`](../CLAUDE.md) - AI assistant quick reference
- **Doc Index:** [`docs/INDEX.md`](../docs/INDEX.md) - Documentation decision tree
- **Architecture:** [`docs/ARCHITECTURE.md`](../docs/ARCHITECTURE.md) - System design
- **Commands:** [`docs/COMMANDS.md`](../docs/COMMANDS.md) - CLI reference

---

## 📈 Metrics

- **Total Skills:** 22
- **Total Rules:** 4
- **Total Workflows:** 2
- **Total Agents:** 3
- **Coverage:** Full development lifecycle
- **Last Updated:** December 23, 2024

---

**For AI Assistants:** This directory contains your primary configuration and workflow definitions. Follow skills for structured tasks, apply rules to all changes, and use agents for specialized analysis.

**For Developers:** Use skills via `/skill-name` or reference workflows directly. The `.claude/settings.json` file enforces security policies automatically.
