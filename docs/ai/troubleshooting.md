## Troubleshooting

### No Default AI Provider Configured

```bash
Error: AI is not configured. Set ai.default_provider in ~/.config/nylas/config.yaml

Solution:
1. Create or edit config file:
   mkdir -p ~/.config/nylas
   nano ~/.config/nylas/config.yaml

2. Add AI configuration:
   ai:
     default_provider: ollama
     ollama:
       host: http://localhost:11434
       model: llama3.2

3. Verify configuration:
   nylas ai config show
```

### Ollama Not Running

```bash
Error: Failed to connect to Ollama at http://localhost:11434

Solution:
1. Start Ollama service:
   ollama serve

2. Verify Ollama is running:
   curl http://localhost:11434/api/tags

3. Check Ollama status:
   ps aux | grep ollama
```

### Model Not Downloaded

```bash
Error: Model 'mistral:latest' not found

Solution:
1. List available models:
   ollama list

2. Download model:
   ollama pull mistral:latest

3. Verify download:
   ollama list
```

### Cloud API Key Missing

```bash
Error: ANTHROPIC_API_KEY not set

Solution:
1. Get API key from https://console.anthropic.com/
2. Set environment variable:
   export ANTHROPIC_API_KEY="sk-ant-..."
3. Add to shell profile:
   echo 'export ANTHROPIC_API_KEY="sk-ant-..."' >> ~/.bashrc
   source ~/.bashrc
```

### Rate Limit Exceeded

```bash
Error: Rate limit exceeded for OpenAI API

Solution:
1. Switch to Ollama (no rate limits):
   nylas ai config set default_provider ollama

2. Wait for rate limit reset (usually 1 minute)

3. Enable fallback to Ollama:
   nylas ai config set fallback.enabled true
   nylas ai config set fallback.providers ollama,claude
```

### Slow Performance

```bash
Issue: AI scheduling taking 30+ seconds

Solutions:
1. Use Groq for faster inference:
   nylas ai config set default_provider groq

2. Use smaller Ollama model:
   ollama pull mistral:7b-instruct-v0.2-q4_0  # Quantized, faster

3. Increase timeout (if supported):
   nylas ai config set timeout 60
```

---

