# Smart Commit

Create a well-formatted commit based on staged changes.

Optional message hint: $ARGUMENTS

## Context

Current status:
!`git status --short`

Staged changes:
!`git diff --cached --stat`

## Instructions

1. **Analyze the staged changes**
   ```bash
   git diff --cached
   ```

2. **Determine commit type**
   - `feat`: New feature
   - `fix`: Bug fix
   - `docs`: Documentation only
   - `test`: Adding/updating tests
   - `refactor`: Code change that neither fixes bug nor adds feature
   - `chore`: Maintenance tasks

3. **Pre-commit checks**
   ```bash
   # Verify no secrets
   git diff --cached | grep -iE "(api_key|password|secret|token|nyk_v0)\s*[=:]" && echo "⛔ SECRET DETECTED" || echo "✓ No secrets"

   # Verify no sensitive files
   git diff --cached --name-only | grep -E "\.(env|pem|key)$" && echo "⛔ SENSITIVE FILE" || echo "✓ No sensitive files"
   ```

4. **Create commit message**

   Format:
   ```
   <type>: <short description (50 chars max)>

   [optional body with more details]
   ```

5. **Execute commit**
   ```bash
   git commit -m "<message>"
   ```

## Rules

- Keep subject line under 50 characters
- Use imperative mood ("add" not "added")
- Don't end subject with period
- Separate subject from body with blank line
- Body should explain WHAT and WHY, not HOW
- Reference issue numbers if applicable

## Examples

Good:
- `feat: add calendar availability check command`
- `fix: resolve nil pointer in email send`
- `docs: update COMMANDS.md with webhook examples`
- `test: add unit tests for contacts API`
- `refactor: extract common HTTP client logic`

Bad:
- `Updated stuff` (vague)
- `feat: add calendar availability check command.` (period)
- `Added new feature for checking calendar` (past tense, too long)
