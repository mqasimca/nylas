# Webhooks

Local webhook development with Nylas CLI.

> **Quick Links:** [README](../README.md) | [Commands](COMMANDS.md) | [Development](DEVELOPMENT.md)

---

## Built-in Webhook Server

```bash
# Start server (local only)
nylas webhooks server

# Start with public tunnel (cloudflared required)
nylas webhooks server --tunnel cloudflared

# Custom port
nylas webhooks server --port 8080 --tunnel cloudflared
```

**Install cloudflared:**
```bash
brew install cloudflared                    # macOS
# Or download from: github.com/cloudflare/cloudflared
```

---

## Webhook Management

```bash
nylas webhooks create --url URL --triggers "event.created"
nylas webhooks list
nylas webhooks show <webhook-id>
nylas webhooks delete <webhook-id>
nylas webhooks test <webhook-id>
```

---

## Quick Test

```bash
# Start server with tunnel
nylas webhooks server --tunnel cloudflared

# Copy the public URL, then create webhook
nylas webhooks create \
  --url https://your-tunnel.trycloudflare.com/webhook \
  --triggers "message.created,message.updated"

# Send test event
nylas webhooks test <webhook-id>
```

---

## TUI Webhook Server

```bash
nylas tui                    # Launch TUI
# Then type: :ws or :server

# Controls:
# s - Start/stop server
# t - Toggle tunnel
# c - Clear events
```

---

**Detailed guide:** See `docs/webhooks/` for cloudflared setup, ngrok configuration, signature verification, and troubleshooting.
