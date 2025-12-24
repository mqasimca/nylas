# Advanced Automation Examples

Complex automation workflows and integration patterns.

---

## Table of Contents

- [OTP Extraction](#otp-extraction)
- [Email Automation Pipelines](#email-automation-pipelines)
- [Calendar Automation](#calendar-automation)
- [Cross-Platform Integration](#cross-platform-integration)
- [Monitoring and Alerts](#monitoring-and-alerts)

---

## OTP Extraction

### Extract OTP codes from emails:

```bash
# Get latest OTP
nylas otp get

# Get OTP from specific service
nylas otp get --from "service@company.com"

# Watch for OTP (poll every 5 seconds)
while true; do
  otp=$(nylas otp get)
  if [ -n "$otp" ]; then
    echo "OTP received: $otp"
    # Copy to clipboard (macOS)
    echo "$otp" | pbcopy
    break
  fi
  sleep 5
done
```

---

### Automated login with OTP:

```bash
#!/bin/bash
# auto-login-with-otp.sh

SERVICE_URL="https://app.example.com/login"
USERNAME="user@example.com"

echo "Starting automated login..."

# Step 1: Request OTP
curl -X POST "$SERVICE_URL/request-otp" \
  -d "email=$USERNAME"

echo "OTP requested. Waiting for email..."

# Step 2: Wait for OTP email
sleep 10

# Step 3: Extract OTP
OTP=$(nylas otp get --from "noreply@example.com")

if [ -z "$OTP" ]; then
  echo "Failed to receive OTP"
  exit 1
fi

echo "OTP received: $OTP"

# Step 4: Login with OTP
curl -X POST "$SERVICE_URL/login" \
  -d "email=$USERNAME&otp=$OTP"

echo "Login successful!"
```

---

### OTP monitoring service:

```python
#!/usr/bin/env python3
# otp-monitor.py

import subprocess
import time
import re

def get_latest_otp():
    """Get OTP using Nylas CLI"""
    result = subprocess.run(
        ['nylas', 'otp', 'get'],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        # Extract OTP code from output
        match = re.search(r'\b\d{6}\b', result.stdout)
        if match:
            return match.group(0)

    return None

def monitor_otp(callback, timeout=60):
    """Monitor for OTP and call callback when received"""
    start_time = time.time()

    while time.time() - start_time < timeout:
        otp = get_latest_otp()

        if otp:
            callback(otp)
            return True

        time.sleep(5)  # Check every 5 seconds

    return False

def handle_otp(code):
    """Handle received OTP"""
    print(f"OTP received: {code}")

    # Auto-fill in browser, send to API, etc.
    # Example: Send to automation service
    # requests.post('http://localhost:8080/otp', json={'code': code})

if __name__ == '__main__':
    print("Monitoring for OTP...")
    success = monitor_otp(handle_otp, timeout=120)

    if not success:
        print("Timeout waiting for OTP")
```

---

## Email Automation Pipelines

### Email-to-task automation:

```python
#!/usr/bin/env python3
# email-to-task.py

import subprocess
import json
import requests

TODOIST_API_KEY = "your-api-key"

def get_unread_emails():
    """Get unread emails using Nylas CLI"""
    result = subprocess.run(
        ['nylas', 'email', 'list', '--unread', '--limit', '10'],
        capture_output=True,
        text=True
    )

    return result.stdout

def create_task(title, due_date=None):
    """Create task in Todoist"""
    url = "https://api.todoist.com/rest/v2/tasks"

    headers = {
        'Authorization': f'Bearer {TODOIST_API_KEY}',
        'Content-Type': 'application/json'
    }

    task = {'content': title}
    if due_date:
        task['due_string'] = due_date

    response = requests.post(url, json=task, headers=headers)
    return response.status_code == 200

def process_emails():
    """Convert emails with [TASK] to Todoist tasks"""
    emails = get_unread_emails()

    for line in emails.split('\n'):
        if '[TASK]' in line and 'Subject:' in line:
            # Extract subject
            subject = line.split('Subject:')[1].strip()
            task_title = subject.replace('[TASK]', '').strip()

            if create_task(task_title):
                print(f"✅ Created task: {task_title}")
            else:
                print(f"❌ Failed to create task: {task_title}")

if __name__ == '__main__':
    process_emails()
```

**Usage with cron:**
```bash
# Run every hour
0 * * * * /path/to/email-to-task.py
```

---

### Automated email triage:

```bash
#!/bin/bash
# email-triage.sh

# Categorize and process emails automatically

# Create folders if they don't exist
mkdir -p ~/email-archives/{urgent,tasks,newsletters,receipts}

# Process unread emails
nylas email list --unread --limit 50 | \
while read -r line; do
  # Extract subject and from
  if [[ $line =~ Subject:\ (.+) ]]; then
    subject="${BASH_REMATCH[1]}"
  fi

  if [[ $line =~ From:\ (.+) ]]; then
    from="${BASH_REMATCH[1]}"
  fi

  # Categorize
  if [[ $subject =~ URGENT|ASAP|IMPORTANT ]]; then
    echo "🔴 Urgent: $subject"
    # Send notification
    osascript -e "display notification \"$subject\" with title \"Urgent Email\""

  elif [[ $subject =~ TODO|TASK|\[Action\ Required\] ]]; then
    echo "📋 Task: $subject"
    # Create task (see above)

  elif [[ $from =~ newsletter|noreply|no-reply ]]; then
    echo "📰 Newsletter: $subject"
    # Archive for later reading

  elif [[ $subject =~ receipt|invoice|order|payment ]]; then
    echo "🧾 Receipt: $subject"
    # Archive for records
  fi
done
```

---

### Smart email forwarding:

```bash
#!/bin/bash
# smart-forwarder.sh

# Forward emails based on rules

forward_email() {
  local msg_id="$1"
  local to="$2"
  local note="$3"

  # Read original email
  original=$(nylas email read "$msg_id")

  # Forward with note
  nylas email send \
    --to "$to" \
    --subject "Fwd: $(echo "$original" | grep Subject: | cut -d: -f2-)" \
    --body "$note

------- Forwarded Message -------
$original" \
    --yes
}

# Process unread emails
nylas email list --unread | \
while read -r line; do
  if [[ $line =~ ID:\ ([a-z0-9_]+) ]]; then
    msg_id="${BASH_REMATCH[1]}"

    # Get email details
    email_data=$(nylas email read "$msg_id")

    # Forward based on content
    if echo "$email_data" | grep -qi "sales inquiry"; then
      forward_email "$msg_id" "sales@company.com" "New sales inquiry"

    elif echo "$email_data" | grep -qi "support request"; then
      forward_email "$msg_id" "support@company.com" "Support ticket"

    elif echo "$email_data" | grep -qi "partnership"; then
      forward_email "$msg_id" "partnerships@company.com" "Partnership opportunity"
    fi
  fi
done
```

---

## Calendar Automation

### Meeting prep automation:

```bash
#!/bin/bash
# meeting-prep.sh

# Prepare for upcoming meetings

# Get today's meetings
nylas calendar events list --days 1 | \
grep -A5 "^Title:" | \
while read -r line; do
  if [[ $line =~ ^Title:\ (.+) ]]; then
    meeting_title="${BASH_REMATCH[1]}"

    # Create prep checklist
    cat > "meeting-prep-${meeting_title// /_}.md" <<EOF
# Meeting Preparation: $meeting_title

## Pre-meeting checklist
- [ ] Review agenda
- [ ] Prepare materials
- [ ] Test video link
- [ ] Review participant backgrounds

## Notes


## Action items

EOF

    echo "✅ Created prep document for: $meeting_title"
  fi
done
```

---

### Automated meeting notes:

```bash
#!/bin/bash
# meeting-notes.sh

# Create meeting notes template

create_notes() {
  local title="$1"
  local start="$2"
  local participants="$3"

  cat > "notes-$(date +%Y%m%d)-${title// /_}.md" <<EOF
# Meeting Notes: $title

**Date:** $(date +%Y-%m-%d)
**Time:** $start
**Participants:** $participants

## Agenda


## Discussion


## Decisions


## Action Items
- [ ]
- [ ]
- [ ]

## Next Steps

EOF

  echo "Created notes template: $title"
}

# Get today's meetings and create notes
nylas calendar events list --days 1
# Parse and create notes for each meeting
```

---

### Calendar sync between services:

```python
#!/usr/bin/env python3
# calendar-sync.py

import subprocess
import requests
import json

GOOGLE_CALENDAR_API = "https://www.googleapis.com/calendar/v3"
GOOGLE_API_KEY = "your-api-key"

def get_nylas_events():
    """Get events from Nylas"""
    result = subprocess.run(
        ['nylas', 'calendar', 'events', 'list', '--days', '7'],
        capture_output=True,
        text=True
    )
    return result.stdout

def sync_to_secondary_calendar(event_data):
    """Sync event to secondary calendar service"""
    # Parse event data
    # Call secondary calendar API
    # Create/update events
    pass

def bidirectional_sync():
    """Two-way sync between calendars"""
    # Get events from both calendars
    # Compare and sync differences
    # Handle conflicts
    pass

if __name__ == '__main__':
    # Implement sync logic
    pass
```

---

## Cross-Platform Integration

### Integrate with project management:

```python
#!/usr/bin/env python3
# pm-integration.py

import subprocess
import requests
import os

JIRA_URL = "https://company.atlassian.net"
JIRA_API_TOKEN = os.environ['JIRA_API_TOKEN']
JIRA_EMAIL = os.environ['JIRA_EMAIL']

def create_jira_ticket(summary, description):
    """Create Jira ticket from email"""

    url = f"{JIRA_URL}/rest/api/3/issue"

    auth = (JIRA_EMAIL, JIRA_API_TOKEN)

    data = {
        "fields": {
            "project": {"key": "PROJ"},
            "summary": summary,
            "description": {
                "type": "doc",
                "version": 1,
                "content": [{
                    "type": "paragraph",
                    "content": [{"type": "text", "text": description}]
                }]
            },
            "issuetype": {"name": "Task"}
        }
    }

    response = requests.post(url, json=data, auth=auth)

    if response.status_code == 201:
        ticket = response.json()
        print(f"Created ticket: {ticket['key']}")
        return ticket['key']

    return None

def process_bug_reports():
    """Convert bug report emails to Jira tickets"""

    # Get emails with [BUG] tag
    result = subprocess.run(
        ['nylas', 'email', 'list', '--subject', 'BUG', '--unread'],
        capture_output=True,
        text=True
    )

    # Parse and create tickets
    # ...

if __name__ == '__main__':
    process_bug_reports()
```

---

### Integrate with CRM:

```python
#!/usr/bin/env python3
# crm-sync.py

import subprocess
import requests

CRM_API = "https://api.crm.com/v1"
CRM_API_KEY = "your-key"

def sync_contacts_to_crm():
    """Sync Nylas contacts to CRM"""

    # Get contacts from Nylas
    result = subprocess.run(
        ['nylas', 'contacts', 'list', '--limit', '100'],
        capture_output=True,
        text=True
    )

    # Parse contacts
    # Upload to CRM
    # Handle duplicates

    pass

def sync_emails_to_crm():
    """Log email interactions in CRM"""

    # Get recent emails
    # Match to CRM contacts
    # Log as activities

    pass

if __name__ == '__main__':
    sync_contacts_to_crm()
    sync_emails_to_crm()
```

---

## Monitoring and Alerts

### Email SLA monitoring:

```bash
#!/bin/bash
# sla-monitor.sh

# Monitor response times for support emails

SLA_HOURS=24

# Get support emails from last 48 hours
nylas email list --from "support@company.com" --limit 100 | \
while read -r line; do
  if [[ $line =~ Date:\ (.+) ]]; then
    email_date="${BASH_REMATCH[1]}"

    # Calculate age
    age_hours=$(( ($(date +%s) - $(date -d "$email_date" +%s)) / 3600 ))

    if [ $age_hours -gt $SLA_HOURS ]; then
      echo "⚠️  SLA breach: Email older than $SLA_HOURS hours"
      # Send alert
      curl -X POST https://alerts.company.com/sla-breach \
        -d "email_age=$age_hours"
    fi
  fi
done
```

---

### Calendar availability monitor:

```bash
#!/bin/bash
# availability-monitor.sh

# Monitor calendar and alert if overbooked

MAX_MEETINGS_PER_DAY=8

meeting_count=$(nylas calendar events list --days 1 | grep -c "^Title:")

if [ $meeting_count -gt $MAX_MEETINGS_PER_DAY ]; then
  echo "⚠️  Warning: $meeting_count meetings today (max: $MAX_MEETINGS_PER_DAY)"

  # Send Slack notification
  curl -X POST "$SLACK_WEBHOOK" \
    -d "{\"text\": \"📅 You have $meeting_count meetings today!\"}"
fi
```

---

### System health dashboard:

```python
#!/usr/bin/env python3
# health-dashboard.py

import subprocess
import time
from datetime import datetime

def check_email_connectivity():
    """Test email API"""
    result = subprocess.run(
        ['nylas', 'email', 'list', '--limit', '1'],
        capture_output=True,
        timeout=10
    )
    return result.returncode == 0

def check_calendar_connectivity():
    """Test calendar API"""
    result = subprocess.run(
        ['nylas', 'calendar', 'list'],
        capture_output=True,
        timeout=10
    )
    return result.returncode == 0

def generate_report():
    """Generate health status report"""
    report = {
        'timestamp': datetime.now().isoformat(),
        'email': check_email_connectivity(),
        'calendar': check_calendar_connectivity(),
    }

    print(f"Health Status: {report}")

    # Send to monitoring service
    # Log to database
    # Generate alerts if needed

    return report

if __name__ == '__main__':
    while True:
        generate_report()
        time.sleep(300)  # Check every 5 minutes
```

---

## Complete Automation System

```python
#!/usr/bin/env python3
# complete-automation-system.py

"""
Complete email and calendar automation system
Combines multiple automation patterns
"""

import subprocess
import requests
import sqlite3
import schedule
import time
from datetime import datetime

class AutomationSystem:
    def __init__(self):
        self.db = self.setup_database()

    def setup_database(self):
        """Setup SQLite database for tracking"""
        conn = sqlite3.connect('automation.db')
        cursor = conn.cursor()

        cursor.execute('''
            CREATE TABLE IF NOT EXISTS processed_emails (
                id TEXT PRIMARY KEY,
                processed_at TEXT,
                action TEXT
            )
        ''')

        conn.commit()
        return conn

    def run_cli_command(self, args):
        """Run Nylas CLI command"""
        result = subprocess.run(
            ['nylas'] + args,
            capture_output=True,
            text=True
        )
        return result.stdout

    def process_new_emails(self):
        """Process new unread emails"""
        print(f"[{datetime.now()}] Processing emails...")

        emails = self.run_cli_command(['email', 'list', '--unread'])

        # Parse and process emails
        # Apply rules and automations
        # Log processed emails

    def sync_calendar(self):
        """Sync calendar with external services"""
        print(f"[{datetime.now()}] Syncing calendar...")

        events = self.run_cli_command(['calendar', 'events', 'list'])

        # Sync with other calendar services
        # Update project management tools
        # Create meeting prep documents

    def generate_daily_digest(self):
        """Generate daily summary"""
        print(f"[{datetime.now()}] Generating daily digest...")

        # Compile email statistics
        # Summarize calendar for the day
        # Send digest email

    def cleanup_old_data(self):
        """Cleanup old processed records"""
        cursor = self.db.cursor()
        cursor.execute('''
            DELETE FROM processed_emails
            WHERE processed_at < datetime('now', '-30 days')
        ''')
        self.db.commit()

    def run(self):
        """Run automation system"""
        # Schedule tasks
        schedule.every(5).minutes.do(self.process_new_emails)
        schedule.every(15).minutes.do(self.sync_calendar)
        schedule.every().day.at("08:00").do(self.generate_daily_digest)
        schedule.every().day.at("00:00").do(self.cleanup_old_data)

        print("Automation system started...")

        while True:
            schedule.run_pending()
            time.sleep(60)

if __name__ == '__main__':
    system = AutomationSystem()
    system.run()
```

---

## Best Practices

### Error handling:

```python
import subprocess
from tenacity import retry, stop_after_attempt, wait_exponential

@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=4, max=10)
)
def reliable_cli_call(args):
    """Call CLI with automatic retries"""
    result = subprocess.run(
        ['nylas'] + args,
        capture_output=True,
        text=True,
        timeout=30
    )

    if result.returncode != 0:
        raise Exception(f"CLI error: {result.stderr}")

    return result.stdout
```

### Rate limiting:

```python
import time
from functools import wraps

def rate_limit(calls_per_minute=60):
    """Decorator for rate limiting"""
    min_interval = 60.0 / calls_per_minute
    last_called = [0.0]

    def decorator(func):
        @wraps(func)
        def wrapper(*args, **kwargs):
            elapsed = time.time() - last_called[0]
            left_to_wait = min_interval - elapsed

            if left_to_wait > 0:
                time.sleep(left_to_wait)

            ret = func(*args, **kwargs)
            last_called[0] = time.time()
            return ret

        return wrapper
    return decorator

@rate_limit(calls_per_minute=30)
def call_api():
    # Your API call
    pass
```

---

## More Resources

- **Email Examples:** [Email Workflows](email-workflows.md)
- **Scheduling Examples:** [Scheduling Guide](scheduling.md)
- **Webhook Examples:** [Webhook Integration](webhooks.md)
- **API Reference:** https://developer.nylas.com/docs/api/v3/
