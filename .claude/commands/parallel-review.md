# Parallel Review

Review code changes using multiple parallel code-reviewer agents for thorough, fast reviews.

Files to review: $ARGUMENTS

## When to Use

- Large PRs with many changed files
- Changes spanning multiple directories
- Pre-commit review of staged changes
- Reviewing changes across different layers (cli, adapters, air)
- When you want multiple "fresh eyes" perspectives

## Instructions

### 1. Get Files to Review

```bash
# Option A: Staged changes
git diff --staged --name-only

# Option B: All uncommitted changes
git diff --name-only

# Option C: PR changes
git diff main...HEAD --name-only

# Option D: Specific files
# (provided as arguments)
```

### 2. Group Files by Directory

Organize files into review groups:

| Group | Files | Reviewer Focus |
|-------|-------|----------------|
| CLI | `internal/cli/**/*.go` | Command structure, flags, UX |
| Adapters | `internal/adapters/**/*.go` | API calls, error handling, retries |
| Air | `internal/air/**/*.go` | Handlers, security, templates |
| Domain | `internal/domain/*.go` | Types, validation, invariants |
| Tests | `*_test.go` | Coverage, edge cases, mocks |
| Frontend | `*.js`, `*.css`, `*.gohtml` | Accessibility, XSS, patterns |

### 3. Launch Parallel Reviewers

Spawn 2-4 code-reviewer agents based on file count:

| Changed Files | Reviewers | Strategy |
|---------------|-----------|----------|
| 1-3 files | 1 | Single thorough review |
| 4-8 files | 2 | Split by directory |
| 9-15 files | 3 | Split by layer |
| 16+ files | 4 | Max parallel |

```
Launch parallel review:

Reviewer 1 (CLI): "Review these CLI files for quality, patterns, and issues:
- internal/cli/email/send.go
- internal/cli/email/helpers.go
Focus: Command structure, error messages, flag handling."

Reviewer 2 (Adapters): "Review these adapter files:
- internal/adapters/nylas/messages.go
Focus: API integration, error handling, retries."

Reviewer 3 (Air): "Review these handler files:
- internal/air/handlers_email.go
Focus: Security, response handling, templates."

Reviewer 4 (Tests): "Review these test files:
- internal/cli/email/send_test.go
Focus: Coverage, edge cases, mock setup."
```

### 4. Reviewer Prompt Template

Each reviewer receives:

```markdown
## Review Task
Review the following files for the Nylas CLI project.

## Files
- `path/to/file1.go`
- `path/to/file2.go`

## Review Checklist
1. **Code Quality** - Functions <50 lines, clear naming, no dead code
2. **Error Handling** - Wrapped with context, user-friendly messages
3. **Security** - No secrets, input validation, no injection risks
4. **Architecture** - Hexagonal compliance, proper layer placement
5. **Testing** - Tests exist, edge cases covered

## Output Format
### Summary
[2-3 sentence overview]

### Issues
| Severity | Location | Issue | Fix |
|----------|----------|-------|-----|
| [emoji] | file:line | Problem | Solution |

### Positive Notes
[What's done well]

### Verdict
[APPROVE / CHANGES NEEDED / DISCUSS]
```

### 5. Consolidate Reviews

Merge all reviewer findings:

```markdown
## Parallel Review Results

### Overview
- **Files reviewed:** [N]
- **Reviewers:** [N]
- **Issues found:** [N critical, N warnings, N info]

### Consolidated Issues

#### Critical (Must Fix)
| Location | Issue | Fix | Reviewer |
|----------|-------|-----|----------|
| file:line | Problem | Solution | Reviewer N |

#### Warnings (Should Fix)
| Location | Issue | Fix | Reviewer |
|----------|-------|-----|----------|
| file:line | Problem | Solution | Reviewer N |

#### Info (Consider)
| Location | Issue | Fix | Reviewer |
|----------|-------|-----|----------|
| file:line | Problem | Solution | Reviewer N |

### Positive Notes
[Consolidated good practices observed]

### Final Verdict
- [ ] ✅ APPROVE - All reviewers approve
- [ ] ⚠️ CHANGES NEEDED - Issues must be fixed
- [ ] ❓ DISCUSS - Conflicting opinions, needs discussion
```

## Review Focus by Layer

| Layer | Key Concerns |
|-------|--------------|
| **CLI** | Flag naming, help text, error messages, command structure |
| **Adapters** | API error handling, retries, rate limiting, context usage |
| **Air** | XSS prevention, CSRF, response handling, template escaping |
| **Domain** | Type safety, validation, invariants, no external deps |
| **Tests** | Coverage, table-driven, mocks, cleanup, independence |

## Examples

### Example 1: Review Staged Changes
```
/parallel-review staged
```
Gets `git diff --staged --name-only`, groups by directory, launches reviewers.

### Example 2: Review Specific Files
```
/parallel-review internal/cli/email/send.go internal/adapters/nylas/messages.go
```
Reviews the two specified files with appropriate focus.

### Example 3: Review PR
```
/parallel-review pr
```
Gets changes from `git diff main...HEAD`, groups and reviews.

## Parallelization Benefits

| Metric | Single Reviewer | 4 Parallel Reviewers |
|--------|-----------------|---------------------|
| Context per file | Shared (degraded) | Fresh per group |
| Review depth | Shallow for later files | Consistent depth |
| Specialization | General | Layer-specific focus |
| Time | Sequential | ~4x faster |

## Conflict Resolution

If reviewers disagree:

| Conflict | Resolution |
|----------|------------|
| Style preference | Follow existing codebase pattern |
| Architecture concern | Escalate to human decision |
| Security issue | Most conservative opinion wins |
| Test coverage | Higher coverage requirement wins |

## Quick Reference

```bash
# Before review: see what changed
git status
git diff --stat

# After review: verify fixes
go build ./...
go vet ./...
golangci-lint run
go test ./... -short
```

## Related Commands

- `/parallel-explore` - Explore codebase in parallel
- `/analyze-coverage` - Check test coverage
- `/security-scan` - Security-focused review
- `/review-pr` - Traditional single-reviewer PR review
