## Setup Guides

### Setup: Ollama (Local, Privacy-First)

**Recommended for:** Healthcare, legal, finance, privacy-conscious users

#### 1. Install Ollama

**macOS:**
```bash
brew install ollama
```

**Linux:**
```bash
curl https://ollama.ai/install.sh | sh
```

**Windows:**
Download from [https://ollama.ai/download](https://ollama.ai/download)

#### 2. Start Ollama Service

```bash
ollama serve
```

#### 3. Download a Model

**Recommended models:**

```bash
# Mistral (7B) - Best balance of speed/quality
ollama pull mistral:latest

# Llama 3 (8B) - Good for general tasks
ollama pull llama3:latest

# CodeLlama (7B) - Better for technical scheduling
ollama pull codellama:latest
```

#### 4. Configure Nylas CLI

```bash
nylas ai config set default_provider ollama
nylas ai config set ollama.model mistral:latest
nylas ai config set ollama.host http://localhost:11434
```

#### 5. Test

```bash
nylas calendar ai schedule "30-min call tomorrow afternoon"
```

✅ **Your data never leaves your machine!**

#### Remote Ollama Hosts

If you're running Ollama on a different machine (e.g., a home server), configure the host:

```bash
# Configure remote Ollama host
nylas ai config set ollama.host http://192.168.1.100:11434

# Or using hostname
nylas ai config set ollama.host http://ollama-server.local:11434

# Test connection
curl http://192.168.1.100:11434/api/tags
```

**Benefits of Remote Ollama:**
- Run on more powerful hardware (GPU-enabled server)
- Share one Ollama instance across multiple machines
- Offload AI processing from laptop

**Security Note:** Ensure your Ollama host is on a trusted network (LAN) and not exposed to the internet.

---

### Setup: Claude (Cloud, Advanced)

**Recommended for:** Complex reasoning, long context analysis, multi-step workflows

#### 1. Get API Key

Visit [https://console.anthropic.com/](https://console.anthropic.com/) and create an API key.

#### 2. Set Environment Variable

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

Add to `~/.bashrc` or `~/.zshrc`:
```bash
echo 'export ANTHROPIC_API_KEY="sk-ant-..."' >> ~/.bashrc
source ~/.bashrc
```

#### 3. Configure Nylas CLI

```bash
nylas ai config set default_provider claude
nylas ai config set claude.model claude-3-5-sonnet-20241022
```

#### 4. Test

```bash
nylas calendar ai schedule "Find time for team planning next week"
```

⚠️ **Note:** Data is sent to Anthropic's API (subject to their privacy policy)

---

### Setup: OpenAI (Cloud, General)

**Recommended for:** General-purpose tasks, wide model selection

#### 1. Get API Key

Visit [https://platform.openai.com/api-keys](https://platform.openai.com/api-keys)

#### 2. Set Environment Variable

```bash
export OPENAI_API_KEY="sk-..."
echo 'export OPENAI_API_KEY="sk-..."' >> ~/.bashrc
source ~/.bashrc
```

#### 3. Configure Nylas CLI

```bash
nylas ai config set default_provider openai
nylas ai config set openai.model gpt-4-turbo
```

#### 4. Test

```bash
nylas calendar ai schedule "Schedule meeting with john@example.com"
```

---

### Setup: Groq (Cloud, Fast)

**Recommended for:** Real-time applications, low-latency requirements

#### 1. Get API Key

Visit [https://console.groq.com/](https://console.groq.com/)

#### 2. Set Environment Variable

```bash
export GROQ_API_KEY="gsk_..."
echo 'export GROQ_API_KEY="gsk_..."' >> ~/.bashrc
source ~/.bashrc
```

#### 3. Configure Nylas CLI

```bash
nylas ai config set default_provider groq
nylas ai config set groq.model mixtral-8x7b-32768
```

#### 4. Test

```bash
nylas calendar ai schedule "Quick 15-min sync tomorrow"
```

---

