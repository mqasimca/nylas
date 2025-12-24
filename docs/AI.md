# AI-Powered Features

**Privacy-first AI assistant with multi-provider support**

> **Quick Links:** [Configuration](ai/configuration.md) | [Providers](ai/providers.md) | [Privacy](ai/privacy-security.md)

---

## Overview

AI assistant features:
- **Natural language scheduling**: "Schedule 30-min call with John next Tuesday"
- **Find optimal meeting times**: Analyze timezones and working hours
- **Conflict resolution**: Auto-suggest rescheduling
- **Focus time protection**: AI-adaptive deep work blocks
- **Meeting analytics**: Pattern analysis and insights
- **Email analysis**: Inbox summary, categorization, and action items

**Time saved:** 7.6 hours/week average | **Errors reduced:** 80%

---

## Privacy-First Design

### Local AI (Ollama) - Default
- ✅ All data stays on your machine
- ✅ No external API calls
- ✅ GDPR/HIPAA compliant
- ✅ Works offline
- ✅ **Zero cost** (completely free)

### Cloud AI - Opt-In
- ⚠️ Requires explicit consent
- ⚠️ Data sent to third-party API
- ✅ Advanced reasoning
- ✅ Faster processing

---

## Quick Start

### 1. Install Local AI (Recommended)
```bash
# Install Ollama (free, private, offline)
brew install ollama
ollama serve
ollama pull llama3.2
```

### 2. Configure
```bash
nylas ai config
# Select: Ollama (default)
# Model: llama3.2
```

### 3. Use AI Features
```bash
# Natural language scheduling
nylas calendar schedule ai "coffee with Alice next Monday 2pm"

# AI analytics
nylas calendar analyze

# Find best meeting time
nylas calendar find-time --participants alice@example.com,bob@example.com --duration 1h
```

---

## Providers

| Provider | Privacy | Cost | Best For |
|----------|---------|------|----------|
| **Ollama** | 🟢 Local | Free | Privacy, offline, zero cost |
| **Claude** | 🟡 Cloud | $$ | Advanced reasoning |
| **OpenAI** | 🟡 Cloud | $$ | Fast processing |
| **Groq** | 🟡 Cloud | $ | Low latency |

**Default:** Ollama (local, private, free)

---

## Features

### Natural Language Scheduling
```bash
nylas calendar schedule ai "team standup every Monday at 10am"
nylas calendar schedule ai "lunch with Bob next week"
nylas calendar schedule ai "30-min 1:1 with Sarah on Thursday"
```

### AI Analytics
```bash
nylas calendar analyze               # Meeting patterns
nylas calendar analyze focus-time    # Deep work analysis
nylas calendar analyze productivity  # Productivity insights
```

### Smart Conflict Resolution
- Auto-detects scheduling conflicts
- Suggests optimal rescheduling times
- Respects working hours and breaks
- Timezone-aware recommendations

### Focus Time Protection
- Learns your meeting patterns
- Blocks focus time automatically
- Adapts to your calendar changes
- Protects against meeting overload

---

## Email AI Features

### Inbox Analysis
```bash
nylas email ai analyze                    # Analyze last 10 emails
nylas email ai analyze --limit 25         # Analyze more emails
nylas email ai analyze --unread           # Only unread emails
nylas email ai analyze --folder SENT      # Analyze specific folder
nylas email ai analyze --provider claude  # Use specific AI provider
```

### What You Get
- **Summary**: Brief overview of your inbox
- **Categories**: Emails grouped by type (Work, Personal, Newsletters, etc.)
- **Action Items**: Emails needing response with urgency levels (high/medium/low)
- **Highlights**: Key points extracted from emails

### Example Output
```
📧 Email Analysis (10 emails)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

📋 Summary
Your inbox contains mostly work-related emails. 3 emails require
immediate attention, and you have 2 newsletter updates.

📁 Categories
  Work (5)
    • Project Update - from alice@company.com
    • Meeting Request - from bob@company.com
  Newsletters (3)
    • Weekly Digest - from news@tech.com

⚡ Action Items
  🔴 HIGH: "Urgent: Contract Review" from legal@company.com
     → Needs response: Contract deadline approaching
  🟡 MEDIUM: "Meeting Request" from bob@company.com
     → Needs response: Awaiting confirmation

✨ Highlights
  • Project deadline moved to January 15
  • Team meeting scheduled for Friday

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Provider: ollama | Tokens: 450
```

---

## Configuration

**Basic config (`~/.config/nylas/config.yaml`):**
```yaml
ai:
  default_provider: ollama  # Required for AI commands
  ollama:
    host: http://localhost:11434
    model: llama3.2
  privacy:
    allow_cloud_ai: false
    local_storage_only: true
```

**For detailed configuration:** `docs/ai/configuration.md`

---

## Privacy Controls

```yaml
ai:
  privacy:
    allow_cloud_ai: false        # Require explicit opt-in
    data_retention: 90           # Days to keep patterns
    local_storage_only: true     # Never send data externally
    anonymize_patterns: true     # Remove PII from learned patterns
```

**Privacy details:** `docs/ai/privacy-security.md`

---

## Detailed Documentation

- **Configuration:** `docs/ai/configuration.md`
- **Provider Setup:** `docs/ai/providers.md`
- **Features Reference:** `docs/ai/features.md`
- **Privacy & Security:** `docs/ai/privacy-security.md`
- **Architecture:** `docs/ai/architecture.md`
- **Troubleshooting:** `docs/ai/troubleshooting.md`
- **Best Practices:** `docs/ai/best-practices.md`
- **FAQ:** `docs/ai/faq.md`

---

**Get started:** `nylas ai config` (choose Ollama for privacy and zero cost)
