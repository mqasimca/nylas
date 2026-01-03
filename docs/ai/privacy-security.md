# AI Privacy & Security

Comprehensive guide to data privacy and security when using AI features in the Nylas CLI.

> **Quick Links:** [AI Overview](../commands/ai.md) | [Configuration](configuration.md) | [Providers](providers.md)

---

## Privacy-First Design

The Nylas CLI AI features are designed with privacy as the default:

| Principle | Implementation |
|-----------|----------------|
| **Local by default** | Ollama runs entirely on your machine |
| **Opt-in cloud** | Cloud AI requires explicit configuration |
| **No data retention** | Transient processing only |
| **Minimal data** | Only necessary context sent to AI |

---

## Provider Comparison

### Privacy & Data Handling

| Provider | Data Location | Data Retention | Encryption | Best For |
|----------|--------------|----------------|------------|----------|
| **Ollama** | Your machine | None (local) | N/A | Maximum privacy |
| **Claude** | Anthropic servers | Not trained on | TLS + at-rest | Balance of privacy/capability |
| **OpenAI** | OpenAI servers | See policy | TLS + at-rest | Fast processing |
| **Groq** | Groq servers | Transient | TLS | Low latency |

### Compliance Considerations

| Provider | GDPR | HIPAA | SOC 2 | Notes |
|----------|------|-------|-------|-------|
| **Ollama** | N/A | N/A | N/A | No data leaves your machine |
| **Claude** | Yes | BAA available | Yes | Enterprise agreements available |
| **OpenAI** | Yes | BAA available | Yes | Enterprise plans available |
| **Groq** | Yes | Contact sales | In progress | Growing enterprise support |

---

## Local AI (Ollama) - Recommended

### Why Local AI?

1. **Complete privacy**: No data leaves your machine
2. **Zero cost**: Free to run
3. **No network required**: Works offline
4. **No rate limits**: Unlimited usage
5. **Full control**: Your data, your rules

### Setup

```bash
# Install Ollama (macOS)
brew install ollama

# Start the service
ollama serve

# Pull a model
ollama pull llama3.2

# Configure Nylas CLI
nylas ai config
# Select: Ollama
# Model: llama3.2
```

### Data Flow (Local)

```
┌─────────────────────────────────────────────┐
│                Your Machine                  │
├─────────────────────────────────────────────┤
│                                             │
│  Nylas CLI ──────────> Ollama               │
│      │                    │                 │
│      │  Email/Calendar    │  AI Processing  │
│      │  Context           │  (100% local)   │
│      │                    │                 │
│      ◄────────────────────┘                 │
│          AI Response                        │
│                                             │
└─────────────────────────────────────────────┘
            Nothing leaves your machine
```

---

## Cloud AI - Opt-In

### When to Consider Cloud AI

- Need advanced reasoning capabilities
- Require faster processing than local hardware provides
- Working with complex multi-step tasks
- Have appropriate data handling agreements in place

### Configuration

```yaml
# ~/.config/nylas/config.yaml
ai:
  default_provider: claude  # or openai, groq
  privacy:
    allow_cloud_ai: true    # Must explicitly enable
  claude:
    api_key: ${ANTHROPIC_API_KEY}  # Use environment variable
```

### Data Flow (Cloud)

```
┌─────────────────┐     TLS 1.2+    ┌─────────────────┐
│   Your Machine  │ ────────────────> │  Cloud Provider │
│                 │                   │                 │
│   Nylas CLI     │                   │   AI Service    │
│   (minimal      │ <──────────────── │   (transient    │
│    context)     │     Response      │    processing)  │
└─────────────────┘                   └─────────────────┘
```

---

## Data Minimization

### What Data is Sent

When using AI features, only necessary context is provided:

| Feature | Data Sent | Not Sent |
|---------|-----------|----------|
| Email analysis | Subject, snippet, sender | Full body, attachments |
| Scheduling | Event times, participants | Event details, notes |
| Meeting finder | Availability windows | Calendar names, locations |

### Example: Email Analysis

```bash
nylas email ai analyze --limit 5
```

**Data sent to AI:**
```json
{
  "emails": [
    {
      "subject": "Project Update",
      "from": "alice@example.com",
      "snippet": "Quick update on the project timeline...",
      "unread": true
    }
  ]
}
```

**Data NOT sent:**
- Full email body
- Attachments
- Email headers
- Thread history
- Account credentials

---

## Privacy Controls

### Configuration Options

```yaml
# ~/.config/nylas/config.yaml
ai:
  privacy:
    allow_cloud_ai: false        # Block all cloud AI
    local_storage_only: true     # Never cache AI responses
    data_retention: 0            # Don't retain patterns
    anonymize_patterns: true     # Remove PII from learned patterns
    max_context_emails: 10       # Limit context size
```

### Command-Line Overrides

```bash
# Force local AI for a command
nylas email ai analyze --provider ollama

# Verify no cloud AI is used
nylas ai status --show-data-flow
```

---

## Security Best Practices

### API Key Management

```bash
# Store API keys in environment (not in config files)
export ANTHROPIC_API_KEY="sk-..."
export OPENAI_API_KEY="sk-..."

# Or use system keyring
nylas ai config  # Stores in keyring when prompted
```

### Audit Trail

```bash
# View AI usage log (if enabled)
nylas ai log --last 10

# Check what providers were used
nylas ai status
```

### Network Security

All cloud AI communications use:
- TLS 1.2+ encryption
- Certificate validation
- Request timeouts (30s default)
- No plaintext transmission

---

## Compliance Recommendations

### For GDPR Compliance

1. Use Ollama (local) for EU data subjects
2. If using cloud AI, ensure provider has EU data processing agreement
3. Document AI usage in your privacy policy
4. Implement data subject access requests

### For HIPAA Compliance

1. Use Ollama (local) for PHI
2. If cloud AI needed, require BAA from provider
3. Ensure minimum necessary standard
4. Log all AI interactions for audit

### For Enterprise Compliance

```yaml
# Recommended enterprise configuration
ai:
  default_provider: ollama       # Local first
  privacy:
    allow_cloud_ai: false        # Require explicit approval
    local_storage_only: true     # No caching
    audit_log: true              # Full audit trail
```

---

## Disabling AI Features

To completely disable AI features:

```yaml
# ~/.config/nylas/config.yaml
ai:
  enabled: false
```

Or remove AI configuration entirely. AI commands will show appropriate error messages.

---

## FAQ

### Q: Does Ollama send any data to the internet?

No. Ollama runs entirely on your machine. It may download models initially, but all inference is local.

### Q: Are my emails stored by cloud AI providers?

No. Cloud AI providers process data transiently. Nylas CLI doesn't enable any data retention features.

### Q: Can I use AI features without any network access?

Yes, with Ollama. Once models are downloaded, all features work offline.

### Q: How do I verify no data is leaving my machine?

Use network monitoring tools or configure a firewall to block the AI provider domains.

### Q: What happens if I accidentally use cloud AI with sensitive data?

Contact the provider about their data deletion policies. Consider rotating any exposed credentials.

---

## Provider Privacy Policies

- **Ollama**: No data leaves your machine
- **Anthropic (Claude)**: [Privacy Policy](https://www.anthropic.com/privacy)
- **OpenAI**: [Privacy Policy](https://openai.com/privacy)
- **Groq**: [Privacy Policy](https://groq.com/privacy)

---

**See also:**
- [AI Overview](../commands/ai.md)
- [AI Configuration](configuration.md)
- [AI Providers](providers.md)
- [Security Practices](../security/practices.md)
