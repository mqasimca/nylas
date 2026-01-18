---
description: Review staged changes and suggest commit message
agent: code-reviewer
subtask: true
---

Review the staged changes and suggest a commit message:
!`git diff --cached --stat`
!`git diff --cached`

Provide:
1. A concise, descriptive commit message following conventional commits format
2. Brief summary of the changes
3. Any concerns or suggestions before committing

Commit message format:
- feat: for new features
- fix: for bug fixes
- refactor: for code refactoring
- test: for test additions/changes
- docs: for documentation updates
- chore: for maintenance tasks
