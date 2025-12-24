## Best Practices

### 1. Start with Ollama (Privacy-First)

Begin with local AI to understand the workflow without sending data to cloud.

```bash
# Initial setup
ollama serve
ollama pull mistral:latest
nylas ai config set default_provider ollama
```

### 2. Use Cloud AI for Complex Tasks Only

Reserve cloud AI (Claude, OpenAI) for tasks that truly need advanced reasoning.

```bash
# Simple scheduling: Use Ollama (local)
nylas calendar ai schedule "30-min call tomorrow"

# Complex multi-timezone: Use Claude (cloud)
nylas calendar ai schedule --provider claude \
  "Find time for team across 5 timezones"
```

### 3. Enable Fallback for Reliability

Configure fallback chain for high availability.

```yaml
ai:
  fallback:
    enabled: true
    providers: [ollama, claude, openai]  # Try in order
```

### 4. Monitor Token Usage

Track cloud API costs to avoid surprises.

```bash
# Check token usage (planned feature)
nylas ai usage --month current

# Set budget alerts (planned feature)
nylas ai set-budget --monthly 50
```

### 5. Clear Learned Patterns Periodically

Refresh AI patterns to avoid outdated preferences.

```bash
# Clear all learned patterns
nylas ai clear-data

# Re-analyze last 90 days
nylas calendar ai analyze --learn-patterns
```

### 6. Use Specific Language

More specific = better AI results.

```bash
# ❌ Vague
nylas calendar ai schedule "meeting with team"

# ✅ Specific
nylas calendar ai schedule \
  "60-minute Q1 planning with engineering team next Wednesday morning"
```

---

