# Parallel Explore

Explore the codebase using multiple parallel agents for faster, context-efficient exploration.

Search scope: $ARGUMENTS

## When to Use

- Large codebase exploration (745+ Go files)
- Feature search across multiple directories
- Understanding how a concept is implemented across layers
- Pre-exploration before writing code
- Context preservation (each agent gets fresh context window)

## Directory Structure

| Directory | Files | Best For |
|-----------|-------|----------|
| `internal/cli/` | 268 | CLI commands, flags, user interactions |
| `internal/adapters/` | 158 | API integrations, external services |
| `internal/air/` | 117 | Web UI handlers, templates, static files |
| `internal/tui2/` | 81 | Bubble Tea TUI components |
| `internal/domain/` | 19 | Core types, business logic |
| `internal/ports/` | 6 | Interface definitions |

## Instructions

### 1. Determine Exploration Scope

Based on the search query, identify which directories are relevant:

| Query Type | Directories to Search |
|------------|----------------------|
| "Where is X implemented?" | All 5 directories |
| "How does feature Y work?" | cli + adapters + domain |
| "Find all Z handlers" | cli + air + tui2 |
| "API integration for W" | adapters + ports |
| "UI components for V" | air + tui2 |

### 2. Launch Parallel Agents

Spawn 4-5 codebase-explorer agents simultaneously:

```
Launch parallel exploration:

Agent 1 (cli): "Search internal/cli/ for [query]. Report: file paths, function names, key patterns."

Agent 2 (adapters): "Search internal/adapters/ for [query]. Report: file paths, function names, API methods."

Agent 3 (air): "Search internal/air/ for [query]. Report: handlers, templates, routes."

Agent 4 (tui2): "Search internal/tui2/ for [query]. Report: components, models, views."

Agent 5 (domain+ports): "Search internal/domain/ and internal/ports/ for [query]. Report: types, interfaces."
```

### 3. Agent Prompt Template

Each agent receives:

```markdown
## Task
Search [directory] for: [user's query]

## Scope
- Directory: [specific path]
- File types: *.go, *.gohtml, *.js, *.css (as appropriate)

## Report Format
### Summary
[2-3 sentences answering the query for this directory]

### Key Files
- `path/file.go:line` - [description]

### Patterns Found
- [Pattern observed]

### Related
- [Other relevant files]
```

### 4. Consolidate Results

After all agents complete, merge their findings:

```markdown
## Exploration Results: [query]

### Summary
[Combined answer from all agents]

### By Layer

#### CLI Layer (internal/cli/)
[Agent 1 findings]

#### Adapter Layer (internal/adapters/)
[Agent 2 findings]

#### Web UI Layer (internal/air/)
[Agent 3 findings]

#### TUI Layer (internal/tui2/)
[Agent 4 findings]

#### Domain Layer (internal/domain/ + ports/)
[Agent 5 findings]

### Cross-Cutting Patterns
[Patterns observed across multiple layers]

### Key Files to Review
1. `path/to/most/relevant.go` - [why]
2. `path/to/second.go` - [why]
3. `path/to/third.go` - [why]
```

## Parallelization Benefits

| Metric | Single Agent | 5 Parallel Agents |
|--------|--------------|-------------------|
| Context usage | 100k+ tokens (degraded) | 20k each (fresh) |
| Exploration depth | Shallow (context exhaustion) | Deep per directory |
| Time | Sequential | ~5x faster |
| Quality | Degrades late-session | Consistent |

## Examples

### Example 1: Feature Search
```
/parallel-explore "email sending functionality"
```
Result: Finds send.go in cli, messages_send.go in adapters, handlers_email.go in air, etc.

### Example 2: Pattern Search
```
/parallel-explore "rate limiting implementation"
```
Result: Finds rate limiter in adapters, acquireRateLimit in tests, retry logic across layers.

### Example 3: Type Search
```
/parallel-explore "Calendar event types and handlers"
```
Result: Finds Event in domain, calendar adapter, calendar CLI commands, calendar handlers in air.

## Best Practices

1. **Be specific** - "email attachments" not "email stuff"
2. **Let agents run** - Don't interrupt parallel execution
3. **Review all layers** - The answer might be distributed
4. **Note cross-cutting concerns** - Some features span all layers

## Troubleshooting

| Issue | Solution |
|-------|----------|
| Agent finds nothing | Broaden search terms |
| Too many results | Narrow to specific layer |
| Agents conflict | They're read-only, conflicts impossible |
| Slow completion | Normal - each agent does thorough search |

## Related Commands

- `/review-pr` - Review code changes (uses parallel reviewers)
- `/analyze-coverage` - Analyze test coverage
- `/security-scan` - Security analysis
