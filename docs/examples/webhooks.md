# Webhook Integration Examples

Real-world webhook integration and event handling patterns.

---

## Table of Contents

- [Basic Webhook Setup](#basic-webhook-setup)
- [Local Development](#local-development)
- [Event Handlers](#event-handlers)
- [Integration Patterns](#integration-patterns)
- [Production Setup](#production-setup)

---

## Basic Webhook Setup

### Create a simple webhook:

```bash
# Create webhook for email events
nylas webhook create \
  --url "https://myapp.com/webhook" \
  --triggers "message.created,message.updated"

# Create webhook for calendar events
nylas webhook create \
  --url "https://myapp.com/webhook" \
  --triggers "calendar.created,calendar.updated,calendar.deleted"

# Create webhook for all events
nylas webhook create \
  --url "https://myapp.com/webhook" \
  --triggers "message.created,calendar.created,contact.created"
```

---

### List and manage webhooks:

```bash
# List all webhooks
nylas webhook list

# Show specific webhook
nylas webhook show <webhook-id>

# Update webhook
nylas webhook update <webhook-id> \
  --url "https://newurl.com/webhook"

# Delete webhook
nylas webhook delete <webhook-id>

# Test webhook
nylas webhook test <webhook-id>
```

---

## Local Development

### Using ngrok for local testing:

```bash
# 1. Install ngrok
brew install ngrok
# or download from https://ngrok.com

# 2. Start local webhook server (see examples below)
python webhook-server.py &

# 3. Create ngrok tunnel
ngrok http 8080

# 4. Copy ngrok URL (e.g., https://abc123.ngrok.io)
# 5. Create webhook with ngrok URL
nylas webhook create \
  --url "https://abc123.ngrok.io/webhook" \
  --triggers "message.created"

# 6. Test webhook
nylas webhook test <webhook-id>
```

---

### Simple webhook server (Python):

```python
#!/usr/bin/env python3
# webhook-server.py

from http.server import BaseHTTPRequestHandler, HTTPServer
import json
import logging

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

class WebhookHandler(BaseHTTPRequestHandler):
    def do_POST(self):
        # Read request body
        content_length = int(self.headers['Content-Length'])
        body = self.rfile.read(content_length)

        # Parse JSON
        try:
            data = json.loads(body)
            logger.info(f"Received webhook: {json.dumps(data, indent=2)}")

            # Process webhook data
            if 'trigger' in data:
                trigger = data['trigger']
                logger.info(f"Webhook trigger: {trigger}")

                # Handle different event types
                if trigger == 'message.created':
                    self.handle_new_message(data)
                elif trigger == 'calendar.created':
                    self.handle_new_event(data)

            # Respond with 200 OK
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            self.wfile.write(json.dumps({"status": "ok"}).encode())

        except json.JSONDecodeError:
            logger.error("Invalid JSON received")
            self.send_response(400)
            self.end_headers()

    def handle_new_message(self, data):
        logger.info("New message received!")
        # Add your custom logic here
        # Example: Send notification, update database, etc.

    def handle_new_event(self, data):
        logger.info("New calendar event created!")
        # Add your custom logic here

def run_server(port=8080):
    server_address = ('', port)
    httpd = HTTPServer(server_address, WebhookHandler)
    logger.info(f"Webhook server running on port {port}")
    httpd.serve_forever()

if __name__ == '__main__':
    run_server()
```

**Run it:**
```bash
chmod +x webhook-server.py
./webhook-server.py
```

---

### Simple webhook server (Node.js):

```javascript
#!/usr/bin/env node
// webhook-server.js

const express = require('express');
const app = express();
const PORT = 8080;

app.use(express.json());

app.post('/webhook', (req, res) => {
  console.log('Received webhook:', JSON.stringify(req.body, null, 2));

  const { trigger, data } = req.body;

  // Handle different event types
  switch (trigger) {
    case 'message.created':
      handleNewMessage(data);
      break;
    case 'calendar.created':
      handleNewEvent(data);
      break;
    default:
      console.log(`Unhandled trigger: ${trigger}`);
  }

  res.json({ status: 'ok' });
});

function handleNewMessage(data) {
  console.log('New message received!');
  // Add your custom logic
}

function handleNewEvent(data) {
  console.log('New calendar event created!');
  // Add your custom logic
}

app.listen(PORT, () => {
  console.log(`Webhook server listening on port ${PORT}`);
});
```

**Run it:**
```bash
npm install express
node webhook-server.js
```

---

## Event Handlers

### Email notification on new message:

```python
#!/usr/bin/env python3
# email-notifier.py

import smtplib
from email.message import EmailMessage

def handle_new_message(webhook_data):
    """Send notification when new email arrives"""

    # Extract message details
    message_data = webhook_data.get('data', {})
    sender = message_data.get('from', [{}])[0].get('email', 'Unknown')
    subject = message_data.get('subject', 'No subject')

    # Create notification email
    msg = EmailMessage()
    msg['Subject'] = f'New Email: {subject}'
    msg['From'] = 'notifications@myapp.com'
    msg['To'] = 'me@myapp.com'
    msg.set_content(f'You have a new email from {sender}\nSubject: {subject}')

    # Send notification
    with smtplib.SMTP('localhost') as s:
        s.send_message(msg)

    print(f'Notification sent for email from {sender}')
```

---

### Slack notification on webhook:

```bash
#!/bin/bash
# slack-notifier.sh

SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# Function to send Slack message
send_to_slack() {
  local message="$1"

  curl -X POST "$SLACK_WEBHOOK" \
    -H 'Content-Type: application/json' \
    -d "{\"text\": \"$message\"}"
}

# Process webhook data
# (This would be called from your webhook handler)

# Example: New email notification
send_to_slack "📧 New email from: john@example.com\nSubject: Important update"

# Example: Calendar event notification
send_to_slack "📅 New calendar event: Team Meeting\nTime: 2:00 PM"
```

---

### Database logging of webhook events:

```python
#!/usr/bin/env python3
# webhook-logger.py

import sqlite3
import json
from datetime import datetime

def log_webhook(webhook_data):
    """Log webhook events to SQLite database"""

    # Connect to database
    conn = sqlite3.connect('webhooks.db')
    cursor = conn.cursor()

    # Create table if not exists
    cursor.execute('''
        CREATE TABLE IF NOT EXISTS webhook_events (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            timestamp TEXT,
            trigger TEXT,
            data TEXT
        )
    ''')

    # Insert webhook data
    cursor.execute('''
        INSERT INTO webhook_events (timestamp, trigger, data)
        VALUES (?, ?, ?)
    ''', (
        datetime.now().isoformat(),
        webhook_data.get('trigger'),
        json.dumps(webhook_data.get('data'))
    ))

    conn.commit()
    conn.close()

    print(f"Logged webhook: {webhook_data.get('trigger')}")
```

---

## Integration Patterns

### CRM integration (new contact webhook):

```python
#!/usr/bin/env python3
# crm-integration.py

import requests

CRM_API_URL = "https://api.crm.com/contacts"
CRM_API_KEY = "your-api-key"

def handle_new_contact(webhook_data):
    """Add new Nylas contact to CRM"""

    contact_data = webhook_data.get('data', {})

    # Extract contact information
    contact = {
        'email': contact_data.get('emails', [{}])[0].get('email'),
        'name': contact_data.get('given_name', ''),
        'company': contact_data.get('company_name', ''),
        'phone': contact_data.get('phone_numbers', [{}])[0].get('number'),
        'source': 'nylas_webhook'
    }

    # Send to CRM
    headers = {
        'Authorization': f'Bearer {CRM_API_KEY}',
        'Content-Type': 'application/json'
    }

    response = requests.post(CRM_API_URL, json=contact, headers=headers)

    if response.status_code == 201:
        print(f"Contact added to CRM: {contact['email']}")
    else:
        print(f"Failed to add contact: {response.status_code}")
```

---

### Auto-responder using webhooks:

```python
#!/usr/bin/env python3
# auto-responder.py

import requests
import os

NYLAS_API_KEY = os.environ['NYLAS_API_KEY']
GRANT_ID = os.environ['NYLAS_GRANT_ID']

def handle_new_message(webhook_data):
    """Auto-respond to certain emails"""

    message_data = webhook_data.get('data', {})
    sender = message_data.get('from', [{}])[0].get('email', '')
    subject = message_data.get('subject', '')

    # Auto-respond to support emails
    if 'support@' in sender.lower() or 'help@' in sender.lower():
        send_auto_reply(sender, subject)

def send_auto_reply(to_email, original_subject):
    """Send automated response"""

    url = f"https://api.nylas.com/v3/grants/{GRANT_ID}/messages/send"

    headers = {
        'Authorization': f'Bearer {NYLAS_API_KEY}',
        'Content-Type': 'application/json'
    }

    body = {
        'to': [{'email': to_email}],
        'subject': f'Re: {original_subject}',
        'body': 'Thank you for your email. We will respond within 24 hours.'
    }

    response = requests.post(url, json=body, headers=headers)

    if response.status_code == 200:
        print(f"Auto-reply sent to: {to_email}")
    else:
        print(f"Failed to send auto-reply: {response.status_code}")
```

---

### Task creation from calendar events:

```python
#!/usr/bin/env python3
# task-creator.py

import requests

TODOIST_API_KEY = "your-todoist-api-key"

def handle_calendar_event(webhook_data):
    """Create Todoist task from calendar event"""

    event_data = webhook_data.get('data', {})

    title = event_data.get('title', '')
    start_time = event_data.get('when', {}).get('start_time', '')

    # Create task in Todoist
    url = "https://api.todoist.com/rest/v2/tasks"

    headers = {
        'Authorization': f'Bearer {TODOIST_API_KEY}',
        'Content-Type': 'application/json'
    }

    task = {
        'content': f'Prepare for: {title}',
        'due_string': start_time,
        'priority': 3
    }

    response = requests.post(url, json=task, headers=headers)

    if response.status_code == 200:
        print(f"Task created: {title}")
    else:
        print(f"Failed to create task: {response.status_code}")
```

---

## Production Setup

### Production webhook server (Python + Flask):

```python
#!/usr/bin/env python3
# production-webhook.py

from flask import Flask, request, jsonify
import logging
import hmac
import hashlib
import os

app = Flask(__name__)

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Webhook secret for verification
WEBHOOK_SECRET = os.environ.get('WEBHOOK_SECRET', '')

def verify_webhook(request):
    """Verify webhook signature"""
    if not WEBHOOK_SECRET:
        return True  # Skip verification if no secret

    signature = request.headers.get('X-Nylas-Signature', '')
    body = request.get_data()

    expected_signature = hmac.new(
        WEBHOOK_SECRET.encode(),
        body,
        hashlib.sha256
    ).hexdigest()

    return hmac.compare_digest(signature, expected_signature)

@app.route('/webhook', methods=['POST'])
def webhook():
    """Handle incoming webhooks"""

    # Verify webhook authenticity
    if not verify_webhook(request):
        logger.warning("Webhook verification failed")
        return jsonify({"error": "Invalid signature"}), 401

    data = request.get_json()
    trigger = data.get('trigger')

    logger.info(f"Received webhook: {trigger}")

    # Route to appropriate handler
    handlers = {
        'message.created': handle_message_created,
        'message.updated': handle_message_updated,
        'calendar.created': handle_calendar_created,
        'calendar.updated': handle_calendar_updated,
    }

    handler = handlers.get(trigger)
    if handler:
        try:
            handler(data)
        except Exception as e:
            logger.error(f"Error handling webhook: {e}")
            return jsonify({"error": "Internal error"}), 500
    else:
        logger.warning(f"Unhandled trigger: {trigger}")

    return jsonify({"status": "ok"}), 200

def handle_message_created(data):
    logger.info("New message created")
    # Your logic here

def handle_message_updated(data):
    logger.info("Message updated")
    # Your logic here

def handle_calendar_created(data):
    logger.info("Calendar event created")
    # Your logic here

def handle_calendar_updated(data):
    logger.info("Calendar event updated")
    # Your logic here

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8080))
    app.run(host='0.0.0.0', port=port)
```

---

### Production deployment (Docker):

```dockerfile
# Dockerfile

FROM python:3.11-slim

WORKDIR /app

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY production-webhook.py .

EXPOSE 8080

CMD ["gunicorn", "--bind", "0.0.0.0:8080", "--workers", "4", "production-webhook:app"]
```

**requirements.txt:**
```
flask==3.0.0
gunicorn==21.2.0
requests==2.31.0
```

**Run with Docker:**
```bash
docker build -t webhook-server .
docker run -p 8080:8080 \
  -e WEBHOOK_SECRET="your-secret" \
  webhook-server
```

---

### Health check endpoint:

```python
@app.route('/health', methods=['GET'])
def health():
    """Health check endpoint"""
    return jsonify({
        "status": "healthy",
        "service": "webhook-handler",
        "version": "1.0.0"
    }), 200
```

---

### Error handling and retries:

```python
from tenacity import retry, stop_after_attempt, wait_exponential

@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=4, max=10)
)
def process_webhook_with_retry(data):
    """Process webhook with automatic retries"""
    # Your processing logic
    # Raises exception on failure, triggering retry
    pass
```

---

## Best Practices

### Webhook security:

1. **Verify signatures** - Always validate webhook signatures
2. **Use HTTPS** - Never use HTTP for webhooks in production
3. **IP whitelisting** - Restrict access to Nylas IP ranges
4. **Rate limiting** - Implement rate limiting on webhook endpoint
5. **Idempotency** - Handle duplicate webhooks gracefully

---

### Performance:

1. **Async processing** - Queue webhooks for background processing
2. **Quick response** - Respond with 200 OK quickly, process later
3. **Logging** - Log all webhook events for debugging
4. **Monitoring** - Monitor webhook failures and latency
5. **Scaling** - Use load balancers for high traffic

---

### Debugging:

```bash
# Test webhook locally
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "trigger": "message.created",
    "data": {
      "subject": "Test",
      "from": [{"email": "test@example.com"}]
    }
  }'

# Check webhook logs
tail -f webhook-server.log

# Test with Nylas CLI
nylas webhook test <webhook-id>
```

---

## Complete Example: Email Notification System

```python
#!/usr/bin/env python3
# complete-notification-system.py

from flask import Flask, request, jsonify
import requests
import logging
import os
from datetime import datetime

app = Flask(__name__)
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Configuration
SLACK_WEBHOOK = os.environ.get('SLACK_WEBHOOK_URL')
NOTIFICATION_EMAIL = os.environ.get('NOTIFICATION_EMAIL')

class NotificationHandler:
    def __init__(self):
        self.handlers = {
            'message.created': self.handle_new_email,
            'calendar.created': self.handle_new_event,
        }

    def process(self, webhook_data):
        trigger = webhook_data.get('trigger')
        handler = self.handlers.get(trigger)

        if handler:
            handler(webhook_data.get('data', {}))
        else:
            logger.warning(f"No handler for: {trigger}")

    def handle_new_email(self, data):
        sender = data.get('from', [{}])[0].get('email', 'Unknown')
        subject = data.get('subject', 'No subject')

        # Check if high priority
        if self.is_high_priority(sender, subject):
            self.send_slack_notification(
                f"🔴 High Priority Email\nFrom: {sender}\nSubject: {subject}"
            )

        logger.info(f"Processed new email from: {sender}")

    def handle_new_event(self, data):
        title = data.get('title', 'Untitled')
        start = data.get('when', {}).get('start_time', '')

        self.send_slack_notification(
            f"📅 New Calendar Event\nTitle: {title}\nTime: {start}"
        )

        logger.info(f"Processed new event: {title}")

    def is_high_priority(self, sender, subject):
        priority_keywords = ['urgent', 'asap', 'important', 'critical']
        priority_senders = ['boss@company.com', 'ceo@company.com']

        return (
            sender.lower() in priority_senders or
            any(kw in subject.lower() for kw in priority_keywords)
        )

    def send_slack_notification(self, message):
        if not SLACK_WEBHOOK:
            return

        requests.post(SLACK_WEBHOOK, json={'text': message})
        logger.info("Slack notification sent")

handler = NotificationHandler()

@app.route('/webhook', methods=['POST'])
def webhook():
    data = request.get_json()
    handler.process(data)
    return jsonify({"status": "ok"}), 200

@app.route('/health', methods=['GET'])
def health():
    return jsonify({"status": "healthy"}), 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)
```

---

## More Resources

- **Webhook Documentation:** [WEBHOOKS.md](../WEBHOOKS.md)
- **Troubleshooting:** [Troubleshooting Guide](../TROUBLESHOOTING.md)
- **API Reference:** https://developer.nylas.com/docs/api/v3/webhooks/
