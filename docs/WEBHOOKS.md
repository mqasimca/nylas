# Webhook Development Guide

Guide for setting up local webhook development with Nylas using tunneling services.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [Development](DEVELOPMENT.md) | [Architecture](ARCHITECTURE.md)

---

## Overview

When developing with Nylas webhooks locally, your machine is typically behind a router/NAT and not directly accessible from the internet. Tunneling services solve this by creating a public URL that forwards requests to your local server.

```
┌─────────────┐      ┌─────────────────────┐      ┌──────────────┐
│   Nylas     │ ──── │  Tunneling Service  │ ──── │ Your Laptop  │
│   Server    │      │  (Public URL)       │      │ (localhost)  │
└─────────────┘      └─────────────────────┘      └──────────────┘
```

This guide covers:
- **Nylas CLI Webhook Server** - Built-in server with cloudflared integration (recommended)
- **Cloudflare Tunnel (cloudflared)** - Free, no account required for quick tunnels
- **ngrok** - Simple setup, widely used

---

## Option 1: Nylas CLI Webhook Server (Recommended)

The Nylas CLI includes a built-in webhook server with optional cloudflared tunnel integration.

### CLI Usage

```bash
# Start server on default port 3000
nylas webhooks server

# Start server with cloudflared tunnel (automatic public URL)
nylas webhooks server --tunnel cloudflared

# Custom port with tunnel
nylas webhooks server --port 8080 --tunnel cloudflared

# With webhook signature verification
nylas webhooks server --tunnel cloudflared --secret your-webhook-secret

# JSON output for scripting
nylas webhooks server --tunnel cloudflared --json

# Quiet mode (only show events)
nylas webhooks server --tunnel cloudflared --quiet
```

### TUI Usage

The TUI also includes a webhook server view:

```bash
# Launch TUI and navigate to webhook server
nylas tui

# Then type :ws or :server to open the webhook server view
```

In the TUI webhook server view:
- Press `s` to start/stop the server
- Press `t` to toggle cloudflared tunnel
- Press `p` to change the port
- Press `c` to clear received events

### Features

- **Built-in HTTP Server:** No need to write a server yourself
- **Cloudflared Integration:** Automatically starts a tunnel for public URL
- **Real-time Event Display:** See webhook events as they arrive
- **Signature Verification:** Optionally verify webhook signatures
- **TUI Integration:** Visual interface for monitoring webhooks

### Example Output

```
╔══════════════════════════════════════════════════════════════╗
║              Nylas Webhook Server                           ║
╚══════════════════════════════════════════════════════════════╝

✓ Server started successfully

  Local URL:    http://localhost:3000/webhook
  Public URL:   https://random-words.trycloudflare.com/webhook

  Tunnel:       cloudflared (connected)

Register this URL with Nylas:
  nylas webhooks create --url https://random-words.trycloudflare.com/webhook --triggers message.created

Press Ctrl+C to stop

─────────────────────────────────────────────────────────────────
Incoming Webhooks:

[14:32:15] message.created ✓
  ID: event-abc123
  Grant: grant-xyz
  Subject: New email received
```

---

## Option 2: Cloudflare Tunnel (cloudflared)

### How It Works

Cloudflared creates an outbound-only connection from your machine to Cloudflare's edge servers. No inbound ports need to be opened on your router.

### Installation

**macOS:**
```bash
brew install cloudflared
```

**Linux:**
```bash
# Debian/Ubuntu
curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb -o cloudflared.deb
sudo dpkg -i cloudflared.deb

# Or download binary directly
curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64 -o cloudflared
chmod +x cloudflared
sudo mv cloudflared /usr/local/bin/
```

**Windows:**
```powershell
# Using winget
winget install Cloudflare.cloudflared

# Or download from GitHub releases
```

### Quick Start (No Account Required)

1. **Start your local webhook server** (e.g., on port 3000):
   ```bash
   # Example: Simple Python server
   python3 -c "
   from http.server import HTTPServer, BaseHTTPRequestHandler
   import json

   class Handler(BaseHTTPRequestHandler):
       def do_POST(self):
           length = int(self.headers.get('Content-Length', 0))
           body = self.rfile.read(length)
           print(f'Webhook received: {body.decode()}')
           self.send_response(200)
           self.end_headers()
           self.wfile.write(b'OK')

   print('Server running on port 3000...')
   HTTPServer(('', 3000), Handler).serve_forever()
   "
   ```

2. **Start the tunnel:**
   ```bash
   cloudflared tunnel --url http://localhost:3000
   ```

3. **Copy the generated URL** from the output:
   ```
   Your quick Tunnel has been created! Visit it at:
   https://random-words-here.trycloudflare.com
   ```

4. **Register the webhook with Nylas:**
   ```bash
   nylas webhooks create \
     --url https://random-words-here.trycloudflare.com/webhook \
     --triggers message.created,message.updated
   ```

### Persistent Setup (With Cloudflare Account)

For a stable URL that doesn't change:

1. **Login to Cloudflare:**
   ```bash
   cloudflared tunnel login
   ```

2. **Create a named tunnel:**
   ```bash
   cloudflared tunnel create nylas-dev
   ```

3. **Configure DNS** (if using your own domain):
   ```bash
   cloudflared tunnel route dns nylas-dev webhooks.yourdomain.com
   ```

4. **Create config file** (`~/.cloudflared/config.yml`):
   ```yaml
   tunnel: nylas-dev
   credentials-file: ~/.cloudflared/<tunnel-id>.json

   ingress:
     - hostname: webhooks.yourdomain.com
       service: http://localhost:3000
     - service: http_status:404
   ```

5. **Run the tunnel:**
   ```bash
   cloudflared tunnel run nylas-dev
   ```

### Cloudflared Pros & Cons

| Pros | Cons |
|------|------|
| Free unlimited bandwidth | Random URL without account |
| No account required for quick tunnels | Slightly more complex persistent setup |
| DDoS protection included | Requires Cloudflare account for custom domains |
| Origin IP hidden | |

---

## Option 2: ngrok

### How It Works

ngrok creates a secure tunnel from a public URL to your local machine. It's known for its simplicity and developer-friendly interface.

### Installation

**macOS:**
```bash
brew install ngrok
```

**Linux:**
```bash
curl -s https://ngrok-agent.s3.amazonaws.com/ngrok.asc | \
  sudo tee /etc/apt/trusted.gpg.d/ngrok.asc >/dev/null && \
  echo "deb https://ngrok-agent.s3.amazonaws.com buster main" | \
  sudo tee /etc/apt/sources.list.d/ngrok.list && \
  sudo apt update && sudo apt install ngrok
```

**Windows:**
```powershell
choco install ngrok
```

**Or download directly:**
Visit https://ngrok.com/download

### Quick Start

1. **Sign up for a free account** at https://ngrok.com

2. **Add your authtoken:**
   ```bash
   ngrok config add-authtoken <your-auth-token>
   ```

3. **Start your local webhook server** (e.g., on port 3000)

4. **Start the tunnel:**
   ```bash
   ngrok http 3000
   ```

5. **Copy the forwarding URL** from the output:
   ```
   Forwarding    https://abc123.ngrok-free.app -> http://localhost:3000
   ```

6. **Register the webhook with Nylas:**
   ```bash
   nylas webhooks create \
     --url https://abc123.ngrok-free.app/webhook \
     --triggers message.created,message.updated
   ```

### ngrok Web Interface

ngrok provides a local web interface for inspecting requests:

```
http://127.0.0.1:4040
```

Features:
- View all incoming requests
- Inspect headers and body
- Replay requests for debugging

### ngrok Configuration File

Create `~/.ngrok2/ngrok.yml` for persistent settings:

```yaml
version: "2"
authtoken: <your-auth-token>
tunnels:
  nylas-webhook:
    proto: http
    addr: 3000
    inspect: true
```

Start with:
```bash
ngrok start nylas-webhook
```

### ngrok Pros & Cons

| Pros | Cons |
|------|------|
| Very simple setup | Free tier has limitations |
| Built-in request inspector | URL changes on restart (free tier) |
| Easy replay for debugging | Requires account |
| Good documentation | Paid for custom domains |

---

## Testing Your Webhook

### Manual Test with curl

```bash
# Test your tunnel endpoint
curl -X POST https://your-tunnel-url.com/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "specversion": "1.0",
    "type": "message.created",
    "source": "nylas",
    "id": "test-event-123",
    "time": "2024-01-15T10:30:00Z",
    "data": {
      "object": {
        "grant_id": "grant-abc",
        "id": "message-xyz",
        "subject": "Test Email"
      }
    }
  }'
```

### Using Nylas CLI

```bash
# Send a test webhook event
nylas webhooks test <webhook-id> --trigger message.created
```

### Verify Webhook Signature

Always verify webhook signatures in production:

```go
import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/hex"
)

func verifySignature(payload []byte, signature, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expected))
}
```

---

## Comparison Table

| Feature | Nylas CLI Server | Cloudflare Tunnel | ngrok |
|---------|------------------|-------------------|-------|
| Setup | One command | Simple | Very simple |
| Built-in server | Yes | No (need your own) | No (need your own) |
| Free tier | Unlimited | Unlimited | Limited |
| Account required | No | No (quick tunnel) | Yes |
| Tunnel included | Optional (cloudflared) | Built-in | Built-in |
| TUI support | Yes | No | No |
| Request display | Real-time in terminal | No | Yes (web UI) |
| Signature verification | Built-in | Manual | Manual |
| Custom domains | Via cloudflared | With account | Paid |
| Stable URL (free) | No | No | No |

---

## Troubleshooting

### 502 Bad Gateway

**Cause:** Your local server isn't running or isn't listening on the correct port.

**Solution:**
```bash
# Verify your server is running
curl http://localhost:3000

# Check if port is in use
lsof -i :3000
```

### Connection Refused

**Cause:** Tunnel can't reach your local server.

**Solution:**
- Ensure your server binds to `localhost` or `0.0.0.0`
- Check firewall settings
- Verify the port number matches

### Webhook Not Received

**Cause:** URL might have changed or webhook isn't registered.

**Solution:**
```bash
# List your webhooks
nylas webhooks list

# Verify the URL matches your current tunnel
nylas webhooks get <webhook-id>

# Update if needed
nylas webhooks update <webhook-id> --url <new-tunnel-url>
```

### URL Changes on Restart

**Cause:** Free tiers generate random URLs.

**Solution:**
- Use a persistent/named tunnel (requires account)
- Update webhook URL after each restart
- Consider a paid plan for stable URLs

---

## Best Practices

1. **Development Only:** Don't use tunnels in production; deploy a proper webhook receiver.

2. **Keep Tunnels Running:** Start your tunnel before registering webhooks with Nylas.

3. **Use HTTPS:** Both cloudflared and ngrok provide HTTPS by default.

4. **Respond Quickly:** Return `200 OK` within 30 seconds to acknowledge receipt.

5. **Idempotent Processing:** Webhooks may be delivered more than once; handle duplicates gracefully.

6. **Verify Signatures:** Always validate the `X-Nylas-Signature` header in production.

7. **Log Everything:** During development, log all incoming webhook payloads for debugging.

---

## Related Commands

```bash
# Webhook server (local development)
nylas webhooks server                      # Start local webhook server
nylas webhooks server --tunnel cloudflared # With public URL via cloudflared

# Webhook management
nylas webhooks list              # List all webhooks
nylas webhooks create            # Create a new webhook
nylas webhooks get <id>          # Get webhook details
nylas webhooks update <id>       # Update webhook URL or triggers
nylas webhooks delete <id>       # Delete a webhook
nylas webhooks test <id>         # Send a test event
nylas webhooks triggers          # List available trigger types

# TUI webhook server
nylas tui                        # Launch TUI, then :ws for webhook server
```

See [Commands Documentation](COMMANDS.md) for full webhook command reference.
