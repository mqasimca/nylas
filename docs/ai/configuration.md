## Configuration

### CLI Configuration Commands

**Quick Configuration with CLI:**

```bash
# View current AI configuration
nylas ai config show

# List all AI settings
nylas ai config list

# Get a specific value
nylas ai config get default_provider
nylas ai config get ollama.model

# Set configuration values
nylas ai config set default_provider ollama
nylas ai config set ollama.host http://localhost:11434
nylas ai config set ollama.model mistral:latest
nylas ai config set claude.model claude-3-5-sonnet-20241022
nylas ai config set fallback.enabled true
nylas ai config set fallback.providers ollama,claude
```

**Available Configuration Keys:**

| Key | Description | Example Value |
|-----|-------------|---------------|
| `default_provider` | Default AI provider to use | `ollama`, `claude`, `openai`, `groq`, `openrouter` |
| `ollama.host` | Ollama server URL | `http://localhost:11434` |
| `ollama.model` | Ollama model name | `mistral:latest`, `llama3:latest` |
| `claude.api_key` | Claude API key | `${ANTHROPIC_API_KEY}` |
| `claude.model` | Claude model name | `claude-3-5-sonnet-20241022` |
| `openai.api_key` | OpenAI API key | `${OPENAI_API_KEY}` |
| `openai.model` | OpenAI model name | `gpt-4-turbo`, `gpt-4o` |
| `groq.api_key` | Groq API key | `${GROQ_API_KEY}` |
| `groq.model` | Groq model name | `mixtral-8x7b-32768` |
| `openrouter.api_key` | OpenRouter API key | `${OPENROUTER_API_KEY}` |
| `openrouter.model` | OpenRouter model name | `anthropic/claude-3.5-sonnet` |
| `fallback.enabled` | Enable fallback providers | `true`, `false` |
| `fallback.providers` | Comma-separated fallback chain | `ollama,claude,openai` |

### Configuration File

Location: `~/.config/nylas/config.yaml`

**Default Configuration (Privacy-First):**
```yaml
ai:
  # Default provider (privacy-first)
  default_provider: ollama

  # Fallback strategy (optional)
  fallback:
    enabled: true
    providers: [ollama, claude]  # Try in order

  # Ollama (local, privacy-first)
  ollama:
    host: http://localhost:11434
    model: mistral:latest
    enabled: true

  # Claude (cloud, advanced reasoning)
  claude:
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-3-5-sonnet-20241022
    enabled: false

  # OpenAI (cloud, general-purpose)
  openai:
    api_key: ${OPENAI_API_KEY}
    model: gpt-4-turbo
    enabled: false

  # Groq (cloud, fast inference)
  groq:
    api_key: ${GROQ_API_KEY}
    model: mixtral-8x7b-32768
    enabled: false

  # Privacy settings
  privacy:
    allow_cloud_ai: false        # Require explicit opt-in
    data_retention: 90           # Days to keep patterns
    local_storage_only: true     # Local storage only

  # Feature toggles
  features:
    natural_language_scheduling: true
    predictive_scheduling: true
    focus_time_protection: true
    conflict_resolution: true
    email_context_analysis: false  # Requires email access
```

### Environment Variables

```bash
# AI Provider API Keys
export ANTHROPIC_API_KEY="sk-ant-..."
export OPENAI_API_KEY="sk-..."
export GROQ_API_KEY="gsk_..."

# Privacy Settings
export NYLAS_AI_ALLOW_CLOUD=false
export NYLAS_AI_LOCAL_ONLY=true

# Ollama Settings
export OLLAMA_HOST=http://localhost:11434
```

---

