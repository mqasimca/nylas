# Email Workflows and Automation

Real-world email workflows and automation examples.

---

## Table of Contents

- [Daily Email Check](#daily-email-check)
- [Email Filtering](#email-filtering)
- [Bulk Email Operations](#bulk-email-operations)
- [Scheduled Sending](#scheduled-sending)
- [Email Templates](#email-templates)
- [Automation Scripts](#automation-scripts)

---

## Daily Email Check

### Morning email routine:

```bash
#!/bin/bash
# morning-email-check.sh

echo "=== Morning Email Summary ==="

# Count unread emails
unread=$(nylas email list --unread | grep -c "From:")
echo "Unread emails: $unread"

# Check for urgent/important
urgent=$(nylas email list --unread | grep -ci "urgent\|important\|asap")
echo "Urgent emails: $urgent"

# List recent unread
echo ""
echo "Recent unread emails:"
nylas email list --unread --limit 10
```

**Usage:**
```bash
chmod +x morning-email-check.sh
./morning-email-check.sh
```

---

### Check for specific sender:

```bash
#!/bin/bash
# check-from-boss.sh

BOSS_EMAIL="boss@company.com"

echo "Checking emails from $BOSS_EMAIL..."

nylas email list --from "$BOSS_EMAIL" --limit 5

# Count unread from boss
unread=$(nylas email list --from "$BOSS_EMAIL" --unread | grep -c "From:")

if [ $unread -gt 0 ]; then
  echo "⚠️  You have $unread unread emails from your boss!"
fi
```

---

## Email Filtering

### Filter by date range:

```bash
# Get emails from today
nylas email list --limit 50 | grep "$(date +%Y-%m-%d)"

# Get emails from this week
nylas email list --limit 100 | grep "$(date +%Y-%m)"
```

---

### Filter by subject keywords:

```bash
# Find meeting emails
nylas email list --subject "meeting"

# Find invoice emails
nylas email list --subject "invoice"

# Case-insensitive search
nylas email list --subject "URGENT"  # Works for "urgent", "Urgent", "URGENT"
```

---

### Complex filtering with grep:

```bash
# Find emails with attachments (if output includes attachment info)
nylas email list --limit 50 | grep -i "attachment"

# Find emails from this month with specific subject
nylas email list --limit 100 | \
  grep "$(date +%Y-%m)" | \
  grep -i "report"

# Find unread emails from specific domain
nylas email list --unread | grep "@company.com"
```

---

## Bulk Email Operations

### Send to multiple recipients:

```bash
# Send to multiple people (comma-separated)
nylas email send \
  --to "person1@example.com,person2@example.com,person3@example.com" \
  --subject "Team Update" \
  --body "Hi team, here's the weekly update..."

# With CC
nylas email send \
  --to "team@company.com" \
  --cc "manager@company.com,director@company.com" \
  --subject "Project Status" \
  --body "Current status..."
```

---

### Send emails from a list:

```bash
#!/bin/bash
# bulk-send.sh

# Read emails from file (one per line)
while IFS= read -r email; do
  echo "Sending to: $email"

  nylas email send \
    --to "$email" \
    --subject "Newsletter - $(date +%B\ %Y)" \
    --body "Dear subscriber, ..." \
    --yes  # Skip confirmation

  # Rate limiting - wait 2 seconds between sends
  sleep 2

done < email-list.txt

echo "Bulk send complete!"
```

**email-list.txt:**
```
user1@example.com
user2@example.com
user3@example.com
```

---

### Send personalized emails:

```bash
#!/bin/bash
# personalized-send.sh

# Format: email,name,company
while IFS=, read -r email name company; do
  echo "Sending to: $name at $company"

  body="Hi $name,

Thank you for your interest in our product. As a company in the $company industry, you might be interested in...

Best regards,
Sales Team"

  nylas email send \
    --to "$email" \
    --subject "Custom Solution for $company" \
    --body "$body" \
    --yes

  sleep 2

done < contacts.csv
```

**contacts.csv:**
```csv
john@acme.com,John Doe,ACME Corp
jane@techco.com,Jane Smith,TechCo
```

---

## Scheduled Sending

### Schedule emails for later:

```bash
# Send in 2 hours
nylas email send \
  --to "team@company.com" \
  --subject "Meeting Reminder" \
  --body "Reminder: Team meeting in 1 hour" \
  --schedule 2h

# Send tomorrow morning
nylas email send \
  --to "team@company.com" \
  --subject "Daily Standup" \
  --body "Good morning! Today's standup at 9:30 AM" \
  --schedule "tomorrow 9am"

# Send next Monday
nylas email send \
  --to "team@company.com" \
  --subject "Week Kickoff" \
  --body "Week of $(date -v+mon +%B\ %d)" \
  --schedule "next monday 8am"
```

---

### Schedule recurring emails:

```bash
#!/bin/bash
# weekly-report.sh

# Run this script with cron every Monday at 8am
# crontab entry: 0 8 * * 1 /path/to/weekly-report.sh

WEEK=$(date +%Y-W%V)

nylas email send \
  --to "management@company.com" \
  --subject "Weekly Report - $WEEK" \
  --body "$(cat weekly-report.txt)" \
  --attach "weekly-report.pdf" \
  --yes
```

**Cron setup:**
```bash
crontab -e

# Add line:
0 8 * * 1 /path/to/weekly-report.sh
```

---

### Schedule different emails for different times:

```bash
#!/bin/bash
# schedule-emails.sh

# Morning email (9 AM tomorrow)
nylas email send \
  --to "team@company.com" \
  --subject "Morning Briefing" \
  --body "Today's priorities..." \
  --schedule "tomorrow 9am" \
  --yes

# Lunch reminder (12 PM tomorrow)
nylas email send \
  --to "team@company.com" \
  --subject "Lunch & Learn" \
  --body "Join us for lunch..." \
  --schedule "tomorrow 12pm" \
  --yes

# EOD reminder (5 PM tomorrow)
nylas email send \
  --to "team@company.com" \
  --subject "EOD Reports Due" \
  --body "Please submit your reports..." \
  --schedule "tomorrow 5pm" \
  --yes
```

---

## Email Templates

### Create reusable templates:

```bash
# templates/meeting-request.txt
Subject: Meeting Request: {TOPIC}
To: {RECIPIENT}

Hi {NAME},

I would like to schedule a meeting to discuss {TOPIC}.

Proposed times:
- {TIME1}
- {TIME2}
- {TIME3}

Please let me know which works best for you.

Best regards,
{YOUR_NAME}
```

---

### Use templates with substitution:

```bash
#!/bin/bash
# send-from-template.sh

RECIPIENT="$1"
NAME="$2"
TOPIC="$3"

TEMPLATE=$(cat templates/meeting-request.txt)

# Simple string replacement
BODY="${TEMPLATE/\{NAME\}/$NAME}"
BODY="${BODY/\{TOPIC\}/$TOPIC}"
BODY="${BODY/\{TIME1\}/$(date -v+1d +%A\ %B\ %d\ at\ 2PM)}"
BODY="${BODY/\{TIME2\}/$(date -v+2d +%A\ %B\ %d\ at\ 2PM)}"
BODY="${BODY/\{TIME3\}/$(date -v+3d +%A\ %B\ %d\ at\ 2PM)}"
BODY="${BODY/\{YOUR_NAME\}/Your Name}"

nylas email send \
  --to "$RECIPIENT" \
  --subject "Meeting Request: $TOPIC" \
  --body "$BODY"
```

**Usage:**
```bash
./send-from-template.sh \
  "colleague@company.com" \
  "John" \
  "Q4 Planning"
```

---

## Automation Scripts

### Auto-reply to certain emails:

```bash
#!/bin/bash
# auto-reply.sh

# Get unread emails from specific sender
nylas email list --from "support@service.com" --unread | \
while read -r line; do
  if [[ $line =~ ID:\ ([a-z0-9_]+) ]]; then
    msg_id="${BASH_REMATCH[1]}"

    echo "Auto-replying to: $msg_id"

    nylas email send \
      --to "support@service.com" \
      --subject "Re: Support Request" \
      --body "Thank you for your email. I will review and respond within 24 hours." \
      --yes
  fi
done
```

---

### Archive old emails:

```bash
#!/bin/bash
# archive-old-emails.sh

# Get emails older than 30 days
THIRTY_DAYS_AGO=$(date -v-30d +%Y-%m-%d)

echo "Archiving emails older than $THIRTY_DAYS_AGO"

nylas email list --limit 1000 | \
  grep "Date:" | \
  while read -r line; do
    # Extract and process date
    # Archive logic here
    echo "Processing: $line"
  done
```

---

### Email notification system:

```bash
#!/bin/bash
# email-monitor.sh

# Monitor for high-priority emails

while true; do
  # Check for urgent emails every 5 minutes
  urgent=$(nylas email list --unread | grep -ci "urgent\|asap\|important")

  if [ $urgent -gt 0 ]; then
    # Send notification (macOS)
    osascript -e "display notification \"You have $urgent urgent emails\" with title \"Email Alert\""

    # Or send SMS/Slack/etc
    # curl -X POST https://slack.com/api/chat.postMessage ...
  fi

  sleep 300  # Wait 5 minutes
done
```

---

### Daily digest email:

```bash
#!/bin/bash
# daily-digest.sh

# Create email summary for the day

DIGEST_FILE="/tmp/daily-digest.txt"

echo "Daily Email Digest - $(date +%B\ %d,\ %Y)" > "$DIGEST_FILE"
echo "===========================================" >> "$DIGEST_FILE"
echo "" >> "$DIGEST_FILE"

# Unread count
unread=$(nylas email list --unread | grep -c "From:")
echo "Unread emails: $unread" >> "$DIGEST_FILE"
echo "" >> "$DIGEST_FILE"

# Top senders
echo "Top Senders:" >> "$DIGEST_FILE"
nylas email list --limit 100 | \
  grep "From:" | \
  sort | \
  uniq -c | \
  sort -rn | \
  head -5 >> "$DIGEST_FILE"

echo "" >> "$DIGEST_FILE"

# Recent emails
echo "Recent emails:" >> "$DIGEST_FILE"
nylas email list --limit 10 >> "$DIGEST_FILE"

# Send digest
nylas email send \
  --to "myself@company.com" \
  --subject "Daily Email Digest - $(date +%Y-%m-%d)" \
  --body "$(cat $DIGEST_FILE)" \
  --yes
```

---

### Integration with other tools:

```bash
#!/bin/bash
# email-to-slack.sh

# Forward urgent emails to Slack

SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

nylas email list --unread | \
  grep -B5 -i "urgent" | \
  while read -r line; do
    if [[ $line =~ Subject:\ (.+) ]]; then
      subject="${BASH_REMATCH[1]}"

      # Send to Slack
      curl -X POST "$SLACK_WEBHOOK" \
        -H "Content-Type: application/json" \
        -d "{\"text\": \"Urgent email: $subject\"}"
    fi
  done
```

---

## Best Practices

### Rate limiting:

```bash
# Add delays between bulk operations
for email in "${emails[@]}"; do
  nylas email send --to "$email" --subject "..." --body "..."
  sleep 2  # Wait 2 seconds between sends
done
```

---

### Error handling:

```bash
# Check if command succeeded
if nylas email send --to "user@example.com" --subject "Test" --body "Test"; then
  echo "Email sent successfully"
else
  echo "Failed to send email" >&2
  exit 1
fi
```

---

### Logging:

```bash
#!/bin/bash
# with-logging.sh

LOG_FILE="/var/log/email-automation.log"

log() {
  echo "[$(date +%Y-%m-%d\ %H:%M:%S)] $1" | tee -a "$LOG_FILE"
}

log "Starting email automation..."

if nylas email send --to "user@example.com" --subject "Test" --body "Test" --yes; then
  log "✅ Email sent to user@example.com"
else
  log "❌ Failed to send email to user@example.com"
fi

log "Email automation complete"
```

---

## Complete Examples

### Customer onboarding automation:

```bash
#!/bin/bash
# customer-onboarding.sh

CUSTOMER_EMAIL="$1"
CUSTOMER_NAME="$2"

# Day 1: Welcome email
nylas email send \
  --to "$CUSTOMER_EMAIL" \
  --subject "Welcome to Our Service!" \
  --body "Hi $CUSTOMER_NAME, welcome aboard! Here's what to expect..." \
  --yes

# Day 3: Getting started tips (scheduled)
nylas email send \
  --to "$CUSTOMER_EMAIL" \
  --subject "Getting Started Tips" \
  --body "Hi $CUSTOMER_NAME, here are some tips to get the most out of our service..." \
  --schedule "3 days" \
  --yes

# Day 7: Check-in (scheduled)
nylas email send \
  --to "$CUSTOMER_EMAIL" \
  --subject "How's It Going?" \
  --body "Hi $CUSTOMER_NAME, we hope you're enjoying our service. Any questions?" \
  --schedule "7 days" \
  --yes
```

---

### Support ticket system:

```bash
#!/bin/bash
# support-tickets.sh

# Monitor for emails with [TICKET] in subject
nylas email list --unread --limit 50 | \
  grep "\[TICKET\]" | \
  while read -r line; do
    # Extract ticket info and process
    # Create ticket in tracking system
    # Send acknowledgment
    echo "Processing ticket: $line"
  done
```

---

## More Resources

- **Command Reference:** [Email Commands](../commands/email.md)
- **Troubleshooting:** [Email Troubleshooting](../troubleshooting/email.md)
- **API Docs:** https://developer.nylas.com/docs/api/v3/
