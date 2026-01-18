# OpenCode Configuration

This directory contains OpenCode-specific configuration files for the Nylas CLI project.

## Directory Structure

```
.opencode/
├── command/        # Custom commands (markdown files)
├── agent/          # Custom agents (markdown files)
└── README.md       # This file
```

## Custom Commands

Custom commands are defined as markdown files in `command/`. Each file becomes a `/command-name` in the TUI.

### Available Commands

- `/debug` - Debug a specific issue with detailed analysis
- `/benchmark` - Run Go benchmarks and analyze performance
- `/coverage` - Generate and analyze test coverage report
- `/api-check` - Verify code uses Nylas API v3 only
- `/commit` - Review staged changes and suggest commit message

Also see commands defined in `../opencode.json`:
- `/test` - Run unit tests with coverage
- `/lint` - Run linter and fix errors
- `/feature` - Add new feature following architecture
- `/review` - Review recent code changes
- `/docs` - Update project documentation
- `/refactor` - Refactor code following project patterns

### Creating New Commands

Create a markdown file in `command/`:

```markdown
---
description: Brief description shown in command palette
agent: build              # Optional: which agent to use
subtask: true            # Optional: run as subtask
---

Your command template here.
Use $ARGUMENTS for command arguments.
Use $1, $2, etc. for positional arguments.
Use !`command` to execute shell commands.
Use @filename to reference files.
```

## Custom Agents

Custom agents are defined as markdown files in `agent/`. Each file becomes an `@agent-name` you can mention.

### Available Agents

Defined in `../opencode.json`:
- `@code-reviewer` - Reviews code for best practices and security (read-only)
- `@test-writer` - Writes table-driven tests
- `@doc-writer` - Writes and maintains documentation

### Creating New Agents

Create a markdown file in `agent/`:

```markdown
---
description: What this agent does
mode: subagent           # primary or subagent
temperature: 0.1        # Optional: 0.0-1.0 (default: model default)
tools:
  write: false          # Optional: tool restrictions
  bash: false
---

System prompt for the agent.
Explain what the agent should focus on.
```

**Note:** Agents and commands don't specify models - they use whatever model the user has configured. This ensures compatibility with any LLM provider.

## Configuration Files

- `../opencode.json` - Main OpenCode configuration
- `../AGENTS.md` - Project-specific instructions loaded automatically

## Learn More

- OpenCode Docs: https://opencode.ai/docs
- Commands: https://opencode.ai/docs/commands
- Agents: https://opencode.ai/docs/agents
- LSP: https://opencode.ai/docs/lsp
