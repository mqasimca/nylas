## Calendar Management

View calendars, manage events, and check availability.

### List Calendars

```bash
nylas calendar list [grant-id]        # List all calendars
nylas cal list                        # Alias
```

**Example output:**
```bash
$ nylas calendar list

Found 3 calendar(s):

NAME                    ID                      PRIMARY   READ-ONLY
Personal                cal_primary_123         Yes
Work                    cal_work_456
Holidays                cal_holidays_789                  Yes
```

### Calendar Events

```bash
# List events
nylas calendar events list [grant-id]
nylas calendar events list --days 14        # Next 14 days
nylas calendar events list --limit 20       # Limit results
nylas calendar events list --calendar <id>  # Specific calendar
nylas calendar events list --show-cancelled # Include cancelled

# List events with timezone conversion (NEW)
nylas calendar events list --timezone America/Los_Angeles  # Convert to specific timezone
nylas calendar events list --show-tz                       # Show timezone abbreviations
nylas calendar events list --timezone Europe/London --show-tz  # Both

# Show event details
nylas calendar events show <event-id>
nylas calendar events show <event-id> --timezone Asia/Tokyo  # Show in specific timezone

# Create event
nylas calendar events create --title "Meeting" --start "2024-12-20 14:00" --end "2024-12-20 15:00"
nylas calendar events create --title "Vacation" --start "2024-12-25" --all-day
nylas calendar events create --title "Team Sync" --start "2024-12-20 10:00" \
  --participant "alice@example.com" --participant "bob@example.com"

# Create event with DST validation (automatically checks for conflicts)
nylas calendar events create --title "Early Meeting" --start "Mar 9, 2025 2:30 AM"

# Create event ignoring DST warnings
nylas calendar events create --title "Early Meeting" --start "Mar 9, 2025 2:30 AM" --ignore-dst-warning

# Delete event
nylas calendar events delete <event-id>
nylas calendar events delete <event-id> --force
```

**DST-Aware Event Creation (NEW):**

When creating events, the CLI automatically checks for Daylight Saving Time conflicts:
- **Spring Forward Gap**: Warns if time doesn't exist (e.g., 2:00-3:00 AM on DST start)
- **Fall Back Duplicate**: Warns if time occurs twice (e.g., 1:00-2:00 AM on DST end)
- Suggests alternative times
- Requires confirmation to proceed or use `--ignore-dst-warning` to skip

**Example DST Conflict Detection:**
```bash
$ nylas calendar events create --title "Early Meeting" --start "Mar 9, 2025 2:30 AM"

⚠️  DST Conflict Detected!

This time will not exist due to Daylight Saving Time (clocks spring forward)

Suggested alternatives:
  1. Schedule 1 hour earlier (before DST)
  2. Schedule at the requested time after DST
  3. Use a different date

Create anyway? [y/N]: n
Cancelled.
```

**Working Hours Validation (NEW):**

The CLI validates event times against configured working hours:
- **Default Hours**: 9:00 AM - 5:00 PM (if not configured)
- **Per-Day Configuration**: Different hours for different days
- **Weekend Support**: Separate weekend hours or disable weekends
- Warns when scheduling outside working hours
- Use `--ignore-working-hours` to skip validation

**Configuration Example:**
```yaml
# ~/.config/nylas/config.yaml
working_hours:
  default:
    enabled: true
    start: "09:00"
    end: "17:00"
  friday:
    enabled: true
    start: "09:00"
    end: "15:00"  # Short Fridays
  weekend:
    enabled: false  # No work on weekends
```

**Example Working Hours Warning:**
```bash
$ nylas calendar events create --title "Late Call" --start "2025-01-15 18:00" --end "2025-01-15 19:00"

⚠️  Working Hours Warning

This event is scheduled outside your working hours:
  • Your hours: 09:00 - 17:00
  • Event time: 6:00 PM Local
  • 1 hour(s) after end

Create anyway? [y/N]: n
Cancelled.

# Or skip validation:
$ nylas calendar events create --title "Late Call" --start "2025-01-15 18:00" --ignore-working-hours
✓ Event created successfully!
```

**Break Time Protection (NEW):**

Protect your lunch breaks and other break periods with hard-block enforcement:

- **Hard Block**: Cannot schedule events during breaks (unlike working hours which allow override)
- **Multiple Breaks**: Configure lunch, coffee breaks, and custom break periods
- **Per-Day Breaks**: Different break times for different days
- Use `--ignore-working-hours` to skip break validation

**Configuration Example:**
```yaml
# ~/.config/nylas/config.yaml
working_hours:
  default:
    enabled: true
    start: "09:00"
    end: "17:00"
    breaks:
      - name: "Lunch"
        start: "12:00"
        end: "13:00"
        type: "lunch"
      - name: "Afternoon Coffee"
        start: "15:00"
        end: "15:15"
        type: "coffee"
  friday:
    enabled: true
    start: "09:00"
    end: "15:00"
    breaks:
      - name: "Lunch"
        start: "11:30"
        end: "12:30"  # Earlier lunch on Fridays
        type: "lunch"
```

**Example Break Conflict:**
```bash
$ nylas calendar events create --title "Quick Sync" --start "2025-01-15 12:30" --end "2025-01-15 13:00"

⛔ Break Time Conflict

Event cannot be scheduled during Lunch (12:00 - 13:00)

Tip: Schedule the event outside of break times, or update your
     break configuration in ~/.nylas/config.yaml
Error: event conflicts with break time

# Break blocks are enforced - you must reschedule:
$ nylas calendar events create --title "Quick Sync" --start "2025-01-15 13:00" --end "2025-01-15 13:30"
✓ Event created successfully!
```

**Timezone Locking (NEW):**

Lock events to a specific timezone to prevent automatic conversion when viewing from different locations. Perfect for in-person events, conferences, or meetings in specific locations:

- **Lock on Creation**: Use `--lock-timezone` when creating events
- **Locked Display**: Shows 🔒 indicator next to time
- **No Auto-Convert**: Time always displays in locked timezone
- **Lock/Unlock**: Use `--lock-timezone` or `--unlock-timezone` in update command

**Example Timezone Locking:**
```bash
# Create event locked to NYC timezone (for in-person meeting)
$ nylas calendar events create \
    --title "NYC Office All-Hands" \
    --start "2025-01-15 09:00" \
    --location "New York Office" \
    --lock-timezone

✓ Event created successfully!

Title: NYC Office All-Hands
When: Wed, Jan 15, 2025, 9:00 AM - 10:00 AM
🔒 Timezone locked: America/New_York
     This event will always display in this timezone, regardless of viewer's location.
ID: event-123

# View locked event (shows lock indicator)
$ nylas calendar events show event-123

NYC Office All-Hands

When
  Wed, Jan 15, 2025, 9:00 AM - 10:00 AM EST 🔒
  (Your local: 6:00 AM PST)

Location
  New York Office

# Unlock timezone
$ nylas calendar events update event-123 --unlock-timezone

✓ Event updated successfully!
🔓 Timezone lock removed
```

**Example output (list events):**
```bash
$ nylas calendar events list --days 7

Found 4 event(s):

Team Standup
  When: Mon, Dec 16, 2024, 9:00 AM - 9:30 AM
  Location: Conference Room A
  Status: confirmed
  Guests: 5 participant(s)
  ID: event_abc123

Project Review
  When: Tue, Dec 17, 2024, 2:00 PM - 3:00 PM
  Status: confirmed
  Guests: 3 participant(s)
  ID: event_def456

Holiday Party
  When: Fri, Dec 20, 2024 (all day)
  Location: Main Office
  Status: confirmed
  ID: event_ghi789
```

**Example output (with timezone conversion):**
```bash
$ nylas calendar events list --timezone America/Los_Angeles

Found 3 event(s):

Team Standup
  When: Mon, Dec 16, 2024, 6:00 AM - 6:30 AM PST
        (Original: 9:00 AM - 9:30 AM EST)
  Location: Zoom
  Status: confirmed
  ID: event_abc123

Client Call
  When: Tue, Dec 17, 2024, 11:00 AM - 12:00 PM PST
        (Original: 7:00 PM - 8:00 PM GMT)
  Location: Google Meet
  Status: confirmed
  ID: event_def456
```

**Example output (show timezone info):**
```bash
$ nylas calendar events list --show-tz

Team Standup
  When: Mon, Dec 16, 2024, 9:00 AM - 9:30 AM EST
  Timezone: America/New_York (EST, UTC-5)
  Location: Conference Room A
  Status: confirmed
  ID: event_abc123
```

**Example output (show event):**
```bash
$ nylas calendar events show event_abc123

Team Standup

When
  Mon, Dec 16, 2024, 9:00 AM - 9:30 AM

Location
  Conference Room A

Description
  Daily team standup meeting to discuss progress and blockers.

Organizer
  John Smith <john@company.com>

Participants
  Alice Johnson <alice@company.com> ✓ accepted
  Bob Wilson <bob@company.com> ✓ accepted
  Carol Davis <carol@company.com> ? tentative

Video Conference
  Provider: zoom
  URL: https://zoom.us/j/123456789

Details
  Status: confirmed
  Busy: true
  ID: event_abc123
  Calendar: cal_primary_123
```

**Example output (list with timezone conversion):**
```bash
$ nylas calendar events list --timezone America/Los_Angeles --show-tz

Found 3 event(s):

Global Team Sync
  When: Mon, Dec 23, 2024, 6:00 AM - 7:00 AM PST
        (Original: Mon, Dec 23, 2024, 9:00 AM - 10:00 AM EST)
  Location: Zoom
  Status: confirmed
  Guests: 12 participant(s)
  ID: event_xyz123

Client Meeting
  When: Tue, Dec 24, 2024, 11:00 AM - 12:00 PM PST
        (Original: Tue, Dec 24, 2024, 2:00 PM - 3:00 PM EST)
  Status: confirmed
  Guests: 3 participant(s)
  ID: event_abc456

Holiday Party
  When: Fri, Dec 27, 2024 (all day)
  Location: Main Office
  Status: confirmed
  ID: event_def789
```

**Example output (show with timezone conversion):**
```bash
$ nylas calendar events show event_xyz123 --timezone Europe/London --show-tz

Global Team Sync

When
  Mon, Dec 23, 2024, 2:00 PM - 3:00 PM GMT
  (Original: Mon, Dec 23, 2024, 9:00 AM - 10:00 AM EST)

Location
  Zoom

Description
  Quarterly planning session with global team members.

Participants
  Alice (New York) <alice@company.com> ✓ accepted
  Bob (London) <bob@company.com> ✓ accepted
  Carol (Tokyo) <carol@company.com> ✓ accepted
  David (Sydney) <david@company.com> ? tentative

Video Conference
  Provider: zoom
  URL: https://zoom.us/j/987654321

Details
  Status: confirmed
  Busy: true
  ID: event_xyz123
  Calendar: cal_primary_123
```

### AI-Powered Scheduling

**NEW:** Schedule meetings using natural language with AI assistance. Supports multiple LLM providers including local privacy-first options.

```bash
# Basic AI scheduling
nylas calendar schedule ai "30-minute meeting with john@example.com next Tuesday afternoon"

# Use specific AI provider
nylas calendar schedule ai --provider claude "team meeting tomorrow morning"
nylas calendar schedule ai --provider openai "quarterly planning next week"
nylas calendar schedule ai --provider groq "quick sync with alice"

# Privacy mode (local LLM)
nylas calendar schedule ai --privacy "sensitive meeting about project X"

# Auto-confirm first option
nylas calendar schedule ai --yes "lunch with team next Friday"

# Specify your timezone
nylas calendar schedule ai --timezone America/Los_Angeles "call with UK team"

# Limit number of suggestions
nylas calendar schedule ai --max-options 5 "1-hour review meeting"
```

**Example Output:**
```bash
$ nylas calendar schedule ai "30-minute meeting with john@example.com next Tuesday afternoon"

🤖 AI Scheduling Assistant
Provider: Claude (Anthropic)

Processing your request: "30-minute meeting with john@example.com next Tuesday afternoon"

Top 3 AI-Suggested Times:

1. 🟢 Tuesday, Jan 21, 2:00 PM PST (Score: 94/100)
   you@example.com: 2:00 PM - 2:30 PM PST
   john@example.com: 5:00 PM - 5:30 PM EST

   Why this is good:
   • Both in working hours
   • No conflicts detected
   • Your calendar shows high productivity at 2 PM historically

2. 🟡 Tuesday, Jan 21, 1:00 PM PST (Score: 82/100)
   you@example.com: 1:00 PM - 1:30 PM PST
   john@example.com: 4:00 PM - 4:30 PM EST

   Why this is good:
   • Post-lunch slot, moderate energy
   • Late afternoon for John (still acceptable)

3. 🟢 Tuesday, Jan 21, 3:00 PM PST (Score: 90/100)
   you@example.com: 3:00 PM - 3:30 PM PST
   john@example.com: 6:00 PM - 6:30 PM EST

   ⚠️  Warnings:
   • Near end of working hours for John

Create meeting with option #1? [y/N/2/3]: y

Creating event...
✓ Event created
  Title: Meeting with john
  When: Tuesday, Jan 21, 2025, 2:00 PM - 2:30 PM PST
  Participants: john@example.com

💰 Estimated cost: ~$0.0150 (1500 tokens)
```

**Privacy Mode (Ollama - Local LLM):**
```bash
$ nylas calendar schedule ai --privacy "team standup tomorrow 9am"

🤖 AI Scheduling Assistant (Privacy Mode)
Provider: Ollama (Local LLM)

Processing locally... ✓

[... AI suggestions ...]

🔒 Privacy: All processing done locally, no data sent to cloud.
```

**Supported AI Providers:**
- `ollama` - Local LLM (privacy-first, free, no API key needed)
- `claude` - Anthropic Claude (best for complex scheduling)
- `openai` - OpenAI GPT-4 (well-balanced)
- `groq` - Groq (very fast, cheap)

**Configuration:**
Add AI configuration to `~/.nylas/config.yaml`:

```yaml
ai:
  default_provider: ollama  # Default provider

  fallback:
    enabled: true
    providers: [ollama, claude]  # Try in order

  ollama:
    host: http://localhost:11434
    model: mistral:latest

  claude:
    api_key: ${ANTHROPIC_API_KEY}
    model: claude-3-5-sonnet-20241022

  openai:
    api_key: ${OPENAI_API_KEY}
    model: gpt-4-turbo

  groq:
    api_key: ${GROQ_API_KEY}
    model: mixtral-8x7b-32768
```

**AI Features:**
- Natural language parsing
- Multi-timezone analysis
- Working hours validation
- DST transition detection
- Participant availability checking
- Meeting time scoring (0-100)
- Detailed reasoning for each option
- Function calling for calendar operations

### Predictive Calendar Analytics

**NEW:** Analyze your meeting history to learn patterns and get AI-powered recommendations for optimizing your calendar.

```bash
# Analyze last 90 days of meetings
nylas calendar analyze

# Analyze custom time period
nylas calendar analyze --days 60

# Score a specific meeting time
nylas calendar analyze --score-time "2025-01-15T14:00:00Z" \
  --participants john@example.com \
  --duration 30

# Show recommendations
nylas calendar analyze --apply
```

**Example Output:**
```bash
$ nylas calendar analyze

🔍 Analyzing 90 days of meeting history...

📊 Analysis Period: 2024-09-22 to 2024-12-21
📅 Total Meetings Analyzed: 156

✅ Meeting Acceptance Patterns
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Overall Acceptance Rate: 84.6%

By Day of Week:
    Monday: 78.3% ████████████████
   Tuesday: 92.1% ██████████████████
 Wednesday: 88.7% ██████████████████
  Thursday: 86.4% █████████████████
    Friday: 64.2% █████████████

By Time of Day (working hours):
  09:00: 72.4% ███████████████
  10:00: 88.9% ██████████████████
  11:00: 91.2% ██████████████████
  14:00: 85.6% █████████████████
  15:00: 79.3% ████████████████

⏱️  Meeting Duration Patterns
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Average Scheduled: 34 minutes
Average Actual: 38 minutes
Overrun Rate: 41.7%

🌍 Timezone Distribution
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  America/Los_Angeles: 89 meetings
  America/New_York: 42 meetings
  Europe/London: 25 meetings

🎯 Productivity Insights
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Peak Focus Times (recommended for deep work):
  1. Tuesday 10:00-12:00 (score: 92/100)
  2. Thursday 10:00-12:00 (score: 88/100)
  3. Wednesday 14:00-16:00 (score: 85/100)
  4. Tuesday 14:00-16:00 (score: 82/100)
  5. Thursday 14:00-16:00 (score: 79/100)

Meeting Density by Day:
    Monday: 3.2 meetings/day
   Tuesday: 2.8 meetings/day
 Wednesday: 3.1 meetings/day
  Thursday: 2.9 meetings/day
    Friday: 1.4 meetings/day

💡 AI Recommendations
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. 🔴 Block Tuesday 10:00-12:00 for focus time [focus_time]
   Historical data shows you have few meetings during this time and accept 92% of meetings outside this block
   📌 Action: Create recurring focus time block
   📈 Impact: Increase productivity by 20-30%
   🎯 Confidence: 92%

2. 🟡 Adjust default meeting duration to 40 minutes [duration_adjustment]
   Your meetings typically run 4 minutes over the scheduled 30 minutes
   📌 Action: Update meeting templates
   📈 Impact: Reduce schedule overruns by 40%
   🎯 Confidence: 78%

3. 🟡 Prefer Tuesday/Wednesday afternoons for team meetings [scheduling_preference]
   Acceptance rate is 88% for Tuesday/Wednesday vs 71% for Monday/Friday
   📌 Action: Suggest Tuesday/Wednesday in meeting invites
   📈 Impact: Reduce declined meetings by 15%
   🎯 Confidence: 85%

📝 Key Insights
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. You accept 92% of meetings on Tuesdays but only 64% on Fridays
2. Your meetings run 12% longer than scheduled on average
3. You have the most focus time on Tuesdays and Thursdays between 10-12 AM
4. Most of your meetings (57%) are with participants in Pacific timezone
```

**Scoring a Specific Meeting Time:**
```bash
$ nylas calendar analyze --score-time "2025-01-21T14:00:00Z" \
  --participants john@example.com \
  --duration 30

🔍 Analyzing historical patterns...

🎯 Meeting Score for Tuesday, Jan 21, 2025 at 2:00 PM PST
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🟢 Overall Score: 86/100
   █████████████████

🎯 Confidence: 85%
📊 Historical Success Rate: 88%

Contributing Factors:
  ➕ Day Preference: +12
     88.7% acceptance rate on Tuesdays
  ➕ Time Preference: +10
     85.6% acceptance rate at 14:00
  ⚪ Productivity: +5
     Moderate productivity time
  ➕ Participant Match: +7
     Based on historical meetings with these participants
  ⚪ Timezone: +0
     Time works well for all timezones

💡 Good time - aligns well with your preferences
```

**Privacy & Local Storage:**
- All pattern learning happens locally
- No meeting data sent to cloud servers
- Patterns stored in `~/.nylas/patterns.json`
- GDPR/HIPAA compliant

**What Gets Analyzed:**
- Meeting acceptance/decline patterns by day and time
- Actual vs scheduled meeting durations
- Timezone distribution of participants
- Productivity windows (times with fewer meetings)
- Per-participant scheduling preferences

**How It Works:**
1. Fetches last 90 days of calendar events
2. Analyzes patterns using local ML algorithms
3. Generates personalized recommendations
4. All processing done locally (privacy-first)

### Conflict Detection & Smart Rescheduling

**NEW:** AI-powered conflict detection and intelligent meeting rescheduling with alternative time suggestions.

**Check for Conflicts:**
```bash
# Check conflicts for a proposed meeting
nylas calendar conflicts check \
  --title "Product Review" \
  --start "2025-01-22T14:00:00Z" \
  --duration 60 \
  --participants team@company.com

# Check and auto-select best alternative
nylas calendar conflicts check \
  --title "Team Sync" \
  --start "2025-01-23T10:00:00Z" \
  --duration 30 \
  --auto-resolve
```

**Example Output:**
```bash
$ nylas calendar conflicts check --title "Weekly Standup" \
  --start "2025-01-22T10:00:00Z" --duration 30

🔍 Analyzing your calendar patterns...
✓ Analyzed 156 meetings from last 90 days

⚙️  Detecting conflicts...

📊 Conflict Analysis
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔴 Hard Conflicts (1)

1. Overlaps with 'Executive Review'
   Event: Executive Review
   Time: Wed, Jan 22 at 10:15 AM PST
   Status: confirmed
   Impact: Cannot attend both meetings simultaneously
   Suggestion: Reschedule to avoid overlap

🟡 Soft Conflicts (2)

1. ⏱️ Back-to-back with 'Team Planning'
   Severity: medium
   Impact: No buffer time between meetings
   ✓ Can auto-resolve

2. 🎯 Interrupts focus time block
   Severity: high
   Impact: Conflicts with Tuesday 10:00-12:00 focus block
   ✓ Can auto-resolve

💡 Recommendations
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Add 15-minute buffer before/after meetings
  Consider Tuesday afternoon instead (92% acceptance rate)

🔄 Suggested Alternative Times
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. 🟢 Wed, Jan 22, 2025 at 2:00 PM PST (Score: 88/100)

   Pros:
   ✓ High acceptance rate on Wednesdays (88.7%)
   ✓ Preferred time slot (85.6% acceptance)
   ✓ No conflicts detected

   💡 This time aligns well with team availability patterns

2. 🟢 Thu, Jan 23, 2025 at 10:00 AM PST (Score: 85/100)

   Pros:
   ✓ High acceptance rate on Thursdays (86.4%)
   ✓ Good time for collaborative work

   Cons:
   ⚠️  Close to another meeting (11 min gap)

3. 🟡 Wed, Jan 22, 2025 at 3:00 PM PST (Score: 74/100)

   Pros:
   ✓ Same day as original
   ✓ No hard conflicts

   Cons:
   ⚠️  Lower acceptance rate for afternoon slots

🤖 AI Recommendation
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Consider rescheduling to Wednesday 2:00 PM. This time has:
- 88% historical acceptance rate
- No scheduling conflicts
- Good match for team availability patterns
- Optimal for collaborative work based on past meetings

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
❌ Status: Cannot proceed (hard conflicts)
```

**AI-Powered Rescheduling:**
```bash
# Get AI suggestions for rescheduling an event
nylas calendar reschedule ai event_abc123 \
  --reason "Conflict with client meeting"

# Reschedule with constraints
nylas calendar reschedule ai event_abc123 \
  --max-delay-days 7 \
  --avoid-days Friday \
  --must-include john@company.com

# Auto-select best time and notify participants
nylas calendar reschedule ai event_abc123 \
  --reason "Calendar conflict" \
  --auto-select \
  --notify
```

**Example Reschedule Output:**
```bash
$ nylas calendar reschedule ai event_abc123

📅 Fetching event event_abc123...
✓ Found: Weekly Team Sync
  Current time: Wed, Jan 22, 2025 at 10:00 AM PST

🔍 Analyzing your calendar patterns...
✓ Analyzed 156 meetings from last 90 days

⚙️  Finding optimal alternative times...

📊 Reschedule Analysis
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Reason: Conflict with client meeting

🔄 Found 5 Alternative Time(s)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

1. 🟢 Wed, Jan 22, 2025 at 2:00 PM PST (Score: 92/100)

   Pros:
   ✓ High acceptance rate on Wednesdays (88.7%)
   ✓ Preferred time slot (88% acceptance)
   ✓ Good match for team schedules

2. 🟢 Thu, Jan 23, 2025 at 10:00 AM PST (Score: 89/100)

   Pros:
   ✓ Same time, next day
   ✓ High acceptance rate on Thursdays

   ⚠️  1 soft conflict(s)

3. 🟡 Wed, Jan 22, 2025 at 11:00 AM PST (Score: 78/100)

   Pros:
   ✓ Same day as original
   ✓ One hour later

   Cons:
   ⚠️  Back-to-back with another meeting

💡 To apply a suggestion, use:
   nylas calendar events update event_abc123 --start 2025-01-22T14:00:00Z
```

**Available Flags:**

Conflict Check:
- `--title` - Meeting title (required)
- `--start` - Start time in RFC3339 format (required)
- `--end` - End time (optional, uses --duration if not set)
- `--duration` - Duration in minutes (default: 60)
- `--participants` - Participant email addresses
- `--auto-resolve` - Automatically select best alternative

AI Reschedule:
- `--reason` - Reason for rescheduling
- `--preferred-times` - Preferred alternative times (RFC3339 format)
- `--max-delay-days` - Maximum days to delay (default: 14)
- `--notify` - Send notification to participants
- `--auto-select` - Automatically apply best alternative
- `--must-include` - Emails that must be available
- `--avoid-days` - Days to avoid (e.g., Friday, Monday)

**Conflict Types Detected:**

Hard Conflicts (blocking):
- Overlapping meetings - Cannot attend both simultaneously

Soft Conflicts (warnings):
- Back-to-back meetings - No buffer time between events
- Focus time interruption - Conflicts with productive work blocks
- Meeting overload - Too many meetings in one day (6+)
- Close proximity - Less than 15 minutes between meetings

**How Conflict Detection Works:**
1. Analyzes proposed meeting time
2. Scans all calendars for conflicts
3. Uses learned patterns to detect soft conflicts
4. Scores alternative times using ML algorithm
5. Suggests top 3-5 alternative times with reasoning

**Reschedule Scoring Algorithm:**
The AI considers multiple factors when scoring alternatives:
- Historical acceptance patterns (day/time preferences)
- Participant availability and preferences
- Meeting density and calendar balance
- Focus time protection
- Timezone fairness for distributed teams

Score ranges:
- 🟢 85-100: Excellent match
- 🟡 70-84: Good option
- 🔴 0-69: Suboptimal (consider other times)

### AI Focus Time Protection

Automatically protect deep work time by analyzing productivity patterns and blocking focus time.

```bash
# Analyze productivity patterns and enable focus time protection
nylas calendar ai focus-time --enable

# Analyze patterns without enabling protection
nylas calendar ai focus-time --analyze

# Create recommended focus blocks
nylas calendar ai focus-time --create

# Customize target focus hours per week
nylas calendar ai focus-time --enable --target-hours 12

# Enable with auto-decline for meeting requests
nylas calendar ai focus-time --enable --auto-decline

# Allow urgent meeting overrides
nylas calendar ai focus-time --enable --allow-override
```

**Example output:**
```bash
$ nylas calendar ai focus-time --enable

🧠 AI Focus Time Protection

Analyzing your productivity patterns...

✨ Discovered Focus Patterns:

  • Peak productivity:
    - Tuesday: 10:00--12:00 (95% focus score) ⭐ Top
    - Thursday: 10:00--12:00 (92% focus score)
    - Wednesday: 09:00--11:00 (85% focus score)

  • Deep work sessions: Average 2.5 hours
  • Most productive day: Wednesday (fewest interruptions)

📅 AI-Recommended Focus Time Blocks:

Weekly Schedule:
  Monday:    ░░░░░░░░░░░░████████░░░░░░░░ 2.0 hrs
  Tuesday:   ████████████░░░░░░░░░░░░░░░░ 2.0 hrs ⭐ Peak
  Wednesday: ████████████████████░░░░░░░░ 4.0 hrs 🎯
  Thursday:  ████████████░░░░░░░░░░░░░░░░ 2.0 hrs ⭐ Peak
  Friday:    ░░░░░░░░░░░░░░░░░░░░████████ 2.0 hrs

Total: 14.0 hours/week protected for focus time

🛡️  Protection Rules:
  1. Auto-decline meeting requests during focus blocks
  2. Suggest alternative times when requests come in
  3. Allow override for "urgent" meetings (you approve)
  4. Dynamically adjust if deadline pressure increases

💡 AI Insights:

  • Your peak productivity is Tuesday at 10:00--12:00 (95% focus score)
  • High meeting density on [Monday Friday] - consider protecting more focus time on these days
  • AI recommends 14.0 hours/week of protected focus time across 5 blocks

📊 Confidence: 100%
   Based on 90 days of calendar history

✅ Focus time protection is enabled!

To create these focus blocks in your calendar, run:
  nylas calendar ai focus-time --create
```

**Creating Focus Blocks:**
```bash
$ nylas calendar ai focus-time --create

🔨 Creating Focus Time Blocks...

✅ Created 5 focus time blocks:

1. Peak productivity time (95% score)
   📅 Tuesday, 10:00 AM--12:00 PM (120 min)
   🔒 Protected with auto-decline: true
   📆 Calendar Event ID: evt_abc123

2. Peak productivity time (92% score)
   📅 Thursday, 10:00 AM--12:00 PM (120 min)
   🔒 Protected with auto-decline: true
   📆 Calendar Event ID: evt_def456

...

✨ Focus time blocks are now protected in your calendar!

To view adaptive schedule recommendations, run:
  nylas calendar ai adapt
```

### Adaptive Schedule Optimization

Real-time adaptive schedule optimization based on changing priorities and workload.

```bash
# Detect and suggest adaptive changes
nylas calendar ai adapt

# Adapt for specific triggers
nylas calendar ai adapt --trigger overload      # Meeting overload
nylas calendar ai adapt --trigger deadline      # Deadline change
nylas calendar ai adapt --trigger focus-risk    # Focus time at risk

# Automatically apply recommended changes
nylas calendar ai adapt --auto-apply
```

**Example output:**
```bash
$ nylas calendar ai adapt

🔄 AI Adaptive Scheduling

Analyzing schedule changes and workload...

📊 Detected Changes:

  • Trigger: Meeting overload detected
  • Affected events: 3
  • Confidence: 85%

📈 Predicted Impact:

  • Focus time gained: 2.0 hours
  • Meetings to reschedule: 2
  • Time saved: 30 minutes
  • Conflicts resolved: 1

  Predicted benefit: Improved focus time availability

🤖 AI Adaptive Actions:

1. Move low-priority meeting to reduce meeting overload
   Event ID: evt_123

2. Move low-priority meeting to reduce meeting overload
   Event ID: evt_456

3. Add additional focus blocks due to deadline pressure

⏸️  Changes require approval (use --auto-apply to apply automatically)

To approve these changes, run:
  nylas calendar ai adapt --auto-apply
```

**How Adaptive Scheduling Works:**
1. Monitors schedule changes and workload patterns
2. Detects triggers (deadline changes, meeting overload, focus time erosion)
3. Analyzes impact of proposed changes
4. Suggests optimizations to protect focus time and reduce overload
5. Learns from historical patterns to improve recommendations

**Adaptive Triggers:**
- **Meeting Overload**: Too many meetings scheduled (18+ hours/week)
- **Deadline Change**: Project deadline moved up, need more focus time
- **Focus Time At Risk**: Protected focus blocks being eroded by meetings
- **Priority Shift**: Task priorities changed, schedule needs adjustment

### Calendar Availability

```bash
# Check free/busy status
nylas calendar availability check [grant-id]
nylas calendar availability check --emails alice@example.com,bob@example.com
nylas calendar availability check --start "tomorrow 9am" --end "tomorrow 5pm"
nylas calendar availability check --duration 7d
nylas calendar availability check --format json

# Find available meeting times
nylas calendar availability find --participants alice@example.com,bob@example.com
nylas calendar availability find --participants team@example.com --duration 60
nylas calendar availability find --participants alice@example.com \
  --start "tomorrow 9am" --end "tomorrow 5pm" --interval 15
```

**Example output (check):**
```bash
$ nylas calendar availability check --emails alice@example.com,bob@example.com

Free/Busy Status: Mon Dec 16 2:30 PM - Tue Dec 17 2:30 PM
────────────────────────────────────────────────────────────

📧 alice@example.com
   Busy times:
   ● Mon Dec 16 3:00 PM - 4:00 PM
   ● Tue Dec 17 9:00 AM - 10:00 AM

📧 bob@example.com
   ✓ Free during this period
```

**Example output (find):**
```bash
$ nylas calendar availability find --participants alice@example.com,bob@example.com --duration 30

Available 30-minute Meeting Slots
────────────────────────────────────────

📅 Mon, Dec 16
   1. 9:00 AM - 9:30 AM
   2. 9:30 AM - 10:00 AM
   3. 11:00 AM - 11:30 AM
   4. 2:00 PM - 2:30 PM

📅 Tue, Dec 17
   5. 10:30 AM - 11:00 AM
   6. 1:00 PM - 1:30 PM
   7. 3:00 PM - 3:30 PM

Found 7 available slots
```

### Smart Meeting Finder (Multi-Timezone)

**NEW:** Find optimal meeting times across multiple timezones with intelligent scoring.

The smart meeting finder analyzes participant timezones and suggests meeting times using a 100-point scoring algorithm that considers:
- **Working Hours (40 pts)**: All participants within working hours
- **Time Quality (25 pts)**: Quality of time for participants (morning/afternoon preference)
- **Cultural Considerations (15 pts)**: Respects cultural norms (no Friday PM, no lunch hour, no Monday early AM)
- **Weekday Preference (10 pts)**: Prefers mid-week meetings (Tuesday/Wednesday best)
- **Holiday Check (10 pts)**: Avoids holidays

```bash
# Find optimal meeting time for multiple participants
nylas calendar find-time --participants alice@example.com,bob@example.com --duration 1h

# Custom working hours and date range
nylas calendar find-time \
  --participants alice@example.com,bob@example.com,carol@example.com \
  --duration 1h \
  --working-start 09:00 \
  --working-end 17:00 \
  --days 7

# 30-minute meeting with weekend availability
nylas calendar find-time \
  --participants team@example.com \
  --duration 30m \
  --exclude-weekends=false
```

**Example output:**
```bash
$ nylas calendar find-time --participants alice@example.com,bob@example.com --duration 1h

🌍 Multi-Timezone Meeting Finder

Participants:
  • alice@example.com: America/New_York
  • bob@example.com: Europe/London

Top 3 Suggested Times:

1. 🟢 Tuesday, Jan 7, 10:00 AM PST (Score: 94/100)
   alice: 1:00 PM - 2:00 PM America/New_York (Good)
   bob: 6:00 PM - 7:00 PM Europe/London (Poor ⚠️)

   Score Breakdown:
   • Working Hours: 40/40 (✓)
   • Time Quality: 22/25
   • Cultural: 15/15
   • Weekday: 10/10
   • Holidays: 7/10

2. 🟢 Wednesday, Jan 8, 11:00 AM PST (Score: 92/100)
   alice: 2:00 PM - 3:00 PM America/New_York (Good)
   bob: 7:00 PM - 8:00 PM Europe/London (Bad 🔴)

   Score Breakdown:
   • Working Hours: 40/40 (✓)
   • Time Quality: 20/25
   • Cultural: 15/15
   • Weekday: 10/10
   • Holidays: 7/10

3. 🟡 Thursday, Jan 9, 9:00 AM PST (Score: 75/100)
   alice: 12:00 PM - 1:00 PM America/New_York (Good)
   bob: 5:00 PM - 6:00 PM Europe/London (Poor ⚠️)

   Score Breakdown:
   • Working Hours: 40/40 (✓)
   • Time Quality: 18/25
   • Cultural: 12/15
   • Weekday: 8/10
   • Holidays: 7/10

💡 Recommendation: Book option #1 for best overall experience
```

**Scoring Legend:**
- 🟢 Excellent (85-100): Great time for all participants
- 🟡 Good (70-84): Acceptable with minor compromises
- 🔴 Poor (<70): Significant compromises, consider alternatives

**Time Quality Labels:**
- ✨ Excellent: 9-11 AM
- Good: 11 AM - 2 PM
- Fair: 2-5 PM
- ⚠️ Poor: 8-9 AM or 5-6 PM
- 🔴 Bad: Outside working hours

### Virtual Calendars

Virtual calendars allow scheduling without connecting to a third-party provider. They're perfect for conference rooms, equipment, or external contractors.

**Features:**
- No OAuth required
- Never expire
- Support calendar and event operations only (no email/contacts)
- Maximum 10 calendars per virtual account

```bash
# List all virtual calendar grants
nylas calendar virtual list
nylas calendar virtual list --json

# Create a virtual calendar grant
nylas calendar virtual create --email conference-room-a@company.com
nylas calendar virtual create --email projector-1@company.com

# Show virtual calendar grant details
nylas calendar virtual show <grant-id>
nylas calendar virtual show <grant-id> --json

# Delete a virtual calendar grant
nylas calendar virtual delete <grant-id>
nylas calendar virtual delete <grant-id> -y  # Skip confirmation
```

**Example workflow:**
```bash
# 1. Create a virtual calendar grant for a conference room
$ nylas calendar virtual create --email conference-room-a@company.com
✓ Created virtual calendar grant
  ID:     vcal-grant-123abc
  Email:  conference-room-a@company.com
  Status: valid

# 2. Create a calendar for this virtual grant
$ nylas calendar create vcal-grant-123abc --name "Conference Room A"
✓ Created calendar
  ID:   primary
  Name: Conference Room A

# 3. Create events on the virtual calendar
$ nylas calendar events create vcal-grant-123abc primary \
  --title "Board Meeting" \
  --start "2024-01-15T14:00:00" \
  --end "2024-01-15T16:00:00"
✓ Created event
```

### Recurring Events

Manage recurring calendar events, including viewing all instances and updating or deleting specific occurrences.

**Supported recurrence patterns:**
- Daily: `RRULE:FREQ=DAILY;COUNT=5`
- Weekly: `RRULE:FREQ=WEEKLY;BYDAY=MO,WE,FR;COUNT=10`
- Monthly: `RRULE:FREQ=MONTHLY;COUNT=12`
- Yearly: `RRULE:FREQ=YEARLY;COUNT=3`

```bash
# List all instances of a recurring event
nylas calendar recurring list <master-event-id> --calendar <calendar-id>
nylas calendar recurring list event-123 --calendar cal-456 --limit 100
nylas calendar recurring list event-123 --calendar cal-456 \
  --start 1704067200 --end 1706745600

# Update a single instance
nylas calendar recurring update <instance-id> --calendar <calendar-id> \
  --title "Updated Meeting Title"
nylas calendar recurring update instance-789 --calendar cal-456 \
  --start "2024-01-15T14:00:00" --end "2024-01-15T15:30:00" \
  --location "Conference Room B"

# Delete a single instance (creates an exception)
nylas calendar recurring delete <instance-id> --calendar <calendar-id>
nylas calendar recurring delete instance-789 --calendar cal-456 -y
```

**Example output (list):**
```bash
$ nylas calendar recurring list event-master-123 --calendar primary

INSTANCE ID        TITLE                   START TIME        STATUS
event-inst-1       Weekly Team Meeting     2024-01-08 10:00  confirmed
event-inst-2       Weekly Team Meeting     2024-01-15 10:00  confirmed
event-inst-3       Weekly Team Meeting     2024-01-22 10:00  confirmed
event-inst-4       Weekly Team Meeting     2024-01-29 10:00  confirmed

Total instances: 4
Master Event ID: event-master-123
```

**Understanding recurring events:**
- **Master Event ID**: The parent event that defines the recurrence pattern
- **Instance**: A single occurrence of the recurring series
- **Exception**: An instance that has been modified or deleted
- When you update an instance, it becomes an exception with custom properties
- When you delete an instance, it adds an EXDATE to the recurrence rule

---

## Scheduling Workflows

### Multi-Timezone Meeting Coordination

```bash
#!/bin/bash
# multi-timezone-meeting.sh

ZONES="America/New_York,Europe/London,Asia/Tokyo"

echo "Finding optimal meeting time for global team..."

# Find best time across all zones
nylas timezone find-meeting \
  --zones "$ZONES" \
  --duration 60 \
  --earliest 9 \
  --latest 17

# View schedule in multiple timezones
nylas calendar events list --timezone America/Los_Angeles
nylas calendar events list --timezone Europe/London --show-tz
```

---

### Batch Create Events

```bash
#!/bin/bash
# batch-create-events.sh

# Read from CSV: title,start,duration,participants
while IFS=, read -r title start duration participants; do
  echo "Creating: $title"

  nylas calendar events create \
    --title "$title" \
    --start "$start" \
    --duration "$duration" \
    --participant "$participants" \
    --yes

  sleep 1  # Rate limiting
done < events.csv
```

---

### Interview Scheduling Automation

```bash
#!/bin/bash
# interview-scheduler.sh

CANDIDATE_EMAIL="$1"
CANDIDATE_NAME="$2"
INTERVIEW_DATE="$3"

# Schedule interview panel
PANEL=(
  "hiring-manager@company.com:30:Technical Screen"
  "engineer@company.com:60:Technical Interview"
  "hr@company.com:30:Culture Fit"
)

current_time="$INTERVIEW_DATE 09:00"

for interview in "${PANEL[@]}"; do
  IFS=: read -r interviewer duration title <<< "$interview"

  nylas calendar events create \
    --title "Interview: $CANDIDATE_NAME - $title" \
    --start "$current_time" \
    --duration "$duration" \
    --participant "$CANDIDATE_EMAIL,$interviewer" \
    --yes

  sleep 1
done
```

---

### DST-Aware Scheduling

```bash
#!/bin/bash
# dst-aware-schedule.sh

# Check for DST transitions before scheduling
nylas timezone dst --zone "America/New_York" --year 2025

# Schedule with explicit timezone to avoid DST issues
nylas calendar events create \
  --title "Important Meeting" \
  --start "2025-03-10 10:00" \
  --timezone "America/New_York" \
  --duration 60
```

---

### Best Practices

**Rate limiting:**
```bash
for event in "${events[@]}"; do
  nylas calendar events create ...
  sleep 1  # Wait between API calls
done
```

**Timezone best practices:**
1. Always specify timezone explicitly for multi-timezone teams
2. Check DST transitions when scheduling near March/November
3. Use IANA timezone names (not abbreviations like EST/PST)
4. Use UTC as common reference for global teams

---

