# Scheduling and Timezone Coordination

Real-world scheduling workflows and multi-timezone meeting coordination.

---

## Table of Contents

- [Team Meeting Scheduling](#team-meeting-scheduling)
- [Multi-Timezone Coordination](#multi-timezone-coordination)
- [Calendar Management](#calendar-management)
- [Availability Checking](#availability-checking)
- [Recurring Events](#recurring-events)
- [Advanced Scheduling](#advanced-scheduling)

---

## Team Meeting Scheduling

### Schedule team meeting:

```bash
# Basic team meeting
nylas calendar events create \
  --title "Team Standup" \
  --start "2024-12-26 09:00" \
  --end "2024-12-26 09:30" \
  --participant "alice@team.com" \
  --participant "bob@team.com" \
  --participant "carol@team.com"
```

---

### Find optimal meeting time:

```bash
# Check team availability
nylas calendar availability check

# Find specific duration
nylas calendar find-time \
  --participants "alice@team.com,bob@team.com,carol@team.com" \
  --duration 60

# With preferences
nylas calendar find-time \
  --participants "alice@team.com,bob@team.com" \
  --duration 30 \
  --earliest 9 \
  --latest 17
```

---

### Schedule recurring team meetings:

```bash
#!/bin/bash
# schedule-recurring.sh

# Schedule daily standup for next month
for day in {1..30}; do
  date=$(date -v+${day}d +%Y-%m-%d)

  # Skip weekends
  day_of_week=$(date -v+${day}d +%u)
  if [ $day_of_week -eq 6 ] || [ $day_of_week -eq 7 ]; then
    continue
  fi

  echo "Scheduling standup for: $date"

  nylas calendar events create \
    --title "Daily Standup" \
    --start "$date 09:00" \
    --end "$date 09:15" \
    --participant "team@company.com" \
    --yes

  sleep 1  # Rate limiting
done
```

---

## Multi-Timezone Coordination

### Find meeting time across timezones:

```bash
# Find time for NY, London, and Tokyo
nylas timezone find-meeting \
  --zones "America/New_York,Europe/London,Asia/Tokyo"

# Output shows:
# Top-scored times that work for all zones
# Considers working hours (9 AM - 5 PM in each zone)
```

---

### View schedule in multiple timezones:

```bash
# Your events in different timezones
nylas calendar events list --timezone America/Los_Angeles
nylas calendar events list --timezone Europe/London
nylas calendar events list --timezone Asia/Tokyo

# Show timezone abbreviations
nylas calendar events list --show-tz
```

---

### Schedule meeting considering all timezones:

```bash
#!/bin/bash
# multi-timezone-meeting.sh

PARTICIPANTS="alice@us.company.com,bob@uk.company.com,carol@jp.company.com"
ZONES="America/New_York,Europe/London,Asia/Tokyo"

echo "Finding optimal meeting time for:"
echo "- Alice (New York)"
echo "- Bob (London)"
echo "- Carol (Tokyo)"
echo ""

# Find best time
echo "Suggested meeting times:"
nylas timezone find-meeting \
  --zones "$ZONES" \
  --duration 60 \
  --earliest 9 \
  --latest 17

# Let user pick a time, then create event
read -p "Enter start time (YYYY-MM-DD HH:MM): " start_time
read -p "Enter timezone: " tz

nylas calendar events create \
  --title "Global Team Sync" \
  --start "$start_time" \
  --timezone "$tz" \
  --duration 60 \
  --participant "$PARTICIPANTS"
```

---

### Convert meeting times for participants:

```bash
#!/bin/bash
# meeting-time-converter.sh

MEETING_TIME="2024-12-26 14:00"
MEETING_TZ="America/New_York"

echo "Meeting time conversion:"
echo "========================================"
echo "Original: $MEETING_TIME $MEETING_TZ"
echo ""

# Convert to other participant timezones
for tz in "Europe/London" "Asia/Tokyo" "Australia/Sydney"; do
  converted=$(nylas timezone convert \
    --from "$MEETING_TZ" \
    --to "$tz" \
    --time "$MEETING_TIME")

  echo "$tz: $converted"
done
```

---

### Send meeting invite with all timezones:

```bash
#!/bin/bash
# send-multi-tz-invite.sh

TITLE="Global Team Meeting"
START="2024-12-26 14:00"
TZ="America/New_York"

# Convert to all participant timezones
NY_TIME=$(nylas timezone convert --from "$TZ" --to "America/New_York" --time "$START")
UK_TIME=$(nylas timezone convert --from "$TZ" --to "Europe/London" --time "$START")
JP_TIME=$(nylas timezone convert --from "$TZ" --to "Asia/Tokyo" --time "$START")

BODY="Team Meeting

Time for each participant:
- New York: $NY_TIME EST
- London: $UK_TIME GMT
- Tokyo: $JP_TIME JST

Join link: https://meet.company.com/team-meeting"

nylas calendar events create \
  --title "$TITLE" \
  --start "$START" \
  --timezone "$TZ" \
  --duration 60 \
  --participant "team-us@company.com,team-uk@company.com,team-jp@company.com" \
  --description "$BODY"
```

---

## Calendar Management

### View upcoming events:

```bash
# Next 7 days (default)
nylas calendar events list

# Next 14 days
nylas calendar events list --days 14

# Specific calendar
nylas calendar events list --calendar <calendar-id>

# Include cancelled events
nylas calendar events list --show-cancelled
```

---

### Create different event types:

```bash
# Regular meeting
nylas calendar events create \
  --title "Project Review" \
  --start "2024-12-26 14:00" \
  --end "2024-12-26 15:00"

# All-day event
nylas calendar events create \
  --title "Company Holiday" \
  --start "2024-12-25" \
  --all-day

# Event with location
nylas calendar events create \
  --title "Client Meeting" \
  --start "2024-12-26 10:00" \
  --location "Conference Room A" \
  --participant "client@company.com"

# Event with description
nylas calendar events create \
  --title "Planning Session" \
  --start "2024-12-26 13:00" \
  --description "Agenda: Q1 planning, budget review, team assignments"
```

---

### Batch create events:

```bash
#!/bin/bash
# batch-create-events.sh

# Read events from CSV
# Format: title,start,duration,participants

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

**events.csv:**
```csv
Team Standup,2024-12-26 09:00,15,team@company.com
Project Review,2024-12-26 14:00,60,alice@company.com,bob@company.com
Client Call,2024-12-27 10:00,30,client@external.com
```

---

## Availability Checking

### Check your availability:

```bash
# Check availability for today
nylas calendar availability check

# Check specific date range
nylas calendar availability check --start "2024-12-26" --end "2024-12-27"
```

---

### Find free time slots:

```bash
#!/bin/bash
# find-free-slots.sh

# Get your schedule
nylas calendar events list --days 1 > schedule.txt

# Process and find gaps
# (This is simplified - actual implementation would parse times)

echo "Your schedule for today:"
cat schedule.txt

echo ""
echo "Available time slots:"
# Logic to find gaps between events
```

---

### Check team availability:

```bash
#!/bin/bash
# team-availability.sh

TEAM_MEMBERS=("alice@company.com" "bob@company.com" "carol@company.com")

echo "Team Availability Check"
echo "======================="

for member in "${TEAM_MEMBERS[@]}"; do
  echo ""
  echo "Checking: $member"

  # Use calendar find-time to check availability
  nylas calendar find-time \
    --participants "$member" \
    --duration 30

done
```

---

## Recurring Events

### Schedule weekly meetings:

```bash
#!/bin/bash
# weekly-meetings.sh

# Schedule weekly team meeting for next 12 weeks
for week in {0..11}; do
  date=$(date -v+${week}w -v1 +%Y-%m-%d)  # Next Monday

  echo "Scheduling for: $date"

  nylas calendar events create \
    --title "Weekly Team Meeting" \
    --start "$date 10:00" \
    --duration 60 \
    --participant "team@company.com" \
    --description "Weekly sync: updates, blockers, planning" \
    --yes

  sleep 1
done
```

---

### Monthly recurring events:

```bash
#!/bin/bash
# monthly-review.sh

# First Friday of each month for next 6 months
for month in {1..6}; do
  # Calculate first Friday
  first_day=$(date -v+${month}m -v1d +%Y-%m-01)
  first_friday=$(date -v+${month}m -v1d -v+fri +%Y-%m-%d)

  echo "Scheduling monthly review: $first_friday"

  nylas calendar events create \
    --title "Monthly Business Review" \
    --start "$first_friday 14:00" \
    --duration 120 \
    --participant "leadership@company.com" \
    --yes

  sleep 1
done
```

---

## Advanced Scheduling

### Smart scheduling based on patterns:

```bash
#!/bin/bash
# smart-schedule.sh

# Analyze calendar to find best meeting times
# (Simplified example)

# Find days with fewer meetings
echo "Analyzing your calendar..."

for day in {1..7}; do
  date=$(date -v+${day}d +%Y-%m-%d)
  count=$(nylas calendar events list --days 1 | grep -c "Title:")

  echo "$date: $count meetings"

  if [ $count -lt 3 ]; then
    echo "  ✅ Good day for scheduling"
  fi
done
```

---

### Schedule 1-on-1 meetings:

```bash
#!/bin/bash
# schedule-one-on-ones.sh

# Team members for 1-on-1s
TEAM=(
  "alice@company.com:Monday:10:00"
  "bob@company.com:Tuesday:10:00"
  "carol@company.com:Wednesday:10:00"
)

# Schedule for next 4 weeks
for week in {0..3}; do
  for member_config in "${TEAM[@]}"; do
    IFS=: read -r email day time <<< "$member_config"

    # Calculate next occurrence of that day
    next_date=$(date -v+${week}w -v"$day" +%Y-%m-%d)

    echo "Scheduling 1-on-1 with $email on $next_date"

    nylas calendar events create \
      --title "1-on-1 with $email" \
      --start "$next_date $time" \
      --duration 30 \
      --participant "$email" \
      --yes

    sleep 1
  done
done
```

---

### Interview scheduling automation:

```bash
#!/bin/bash
# interview-scheduler.sh

CANDIDATE_EMAIL="$1"
CANDIDATE_NAME="$2"
INTERVIEW_DATE="$3"

# Schedule interview panel
PANEL=(
  "hiring-manager@company.com:30:Technical Screen"
  "engineer1@company.com:60:Technical Interview"
  "engineer2@company.com:60:System Design"
  "hr@company.com:30:Culture Fit"
)

current_time="$INTERVIEW_DATE 09:00"

for interview in "${PANEL[@]}"; do
  IFS=: read -r interviewer duration title <<< "$interview"

  echo "Scheduling: $title with $interviewer"

  nylas calendar events create \
    --title "Interview: $CANDIDATE_NAME - $title" \
    --start "$current_time" \
    --duration "$duration" \
    --participant "$CANDIDATE_EMAIL,$interviewer" \
    --description "Interview with $CANDIDATE_NAME" \
    --yes

  # Calculate next start time (add duration + 15 min break)
  break_time=$((duration + 15))
  current_time=$(date -v+"${break_time}M" -j -f "%Y-%m-%d %H:%M" "$current_time" +"%Y-%m-%d %H:%M")

  sleep 1
done

echo "Interview day scheduled for $CANDIDATE_NAME"
```

**Usage:**
```bash
./interview-scheduler.sh \
  "candidate@email.com" \
  "Jane Smith" \
  "2024-12-30"
```

---

### Resource booking system:

```bash
#!/bin/bash
# book-conference-room.sh

ROOM_CALENDAR="conf-room-a@company.com"
TITLE="$1"
START="$2"
DURATION="${3:-60}"

# Check if room is available
echo "Checking availability of Conference Room A..."

available=$(nylas calendar find-time \
  --participants "$ROOM_CALENDAR" \
  --duration "$DURATION" | grep "$START")

if [ -n "$available" ]; then
  echo "Room is available. Booking..."

  nylas calendar events create \
    --title "$TITLE" \
    --start "$START" \
    --duration "$DURATION" \
    --participant "$ROOM_CALENDAR"

  echo "✅ Conference Room A booked for $START"
else
  echo "❌ Conference Room A not available at $START"
  echo "Finding alternative times..."

  nylas calendar find-time \
    --participants "$ROOM_CALENDAR" \
    --duration "$DURATION"
fi
```

---

## DST-Aware Scheduling

### Check for DST transitions:

```bash
#!/bin/bash
# dst-aware-schedule.sh

TIMEZONE="America/New_York"
EVENT_DATE="2025-03-09"  # Near DST transition

# Check if DST transition occurs
echo "Checking DST status for $TIMEZONE on $EVENT_DATE..."

nylas timezone dst --zone "$TIMEZONE" --year 2025

# Get transition dates
spring_forward=$(nylas timezone dst --zone "$TIMEZONE" --year 2025 | grep "Spring")

echo ""
echo "⚠️  Warning: Schedule carefully around DST transitions"
echo "Times like 2:30 AM may not exist (spring forward)"
echo "Times like 1:30 AM happen twice (fall back)"

# Schedule event with DST awareness
nylas calendar events create \
  --title "Important Meeting" \
  --start "$EVENT_DATE 10:00" \
  --timezone "$TIMEZONE" \
  --duration 60
```

---

### Multi-timezone DST coordination:

```bash
#!/bin/bash
# multi-tz-dst-check.sh

ZONES=("America/New_York" "Europe/London" "Asia/Tokyo")
YEAR="2025"

echo "DST Transition Calendar for $YEAR"
echo "====================================="

for zone in "${ZONES[@]}"; do
  echo ""
  echo "$zone:"
  nylas timezone dst --zone "$zone" --year "$YEAR"
done

echo ""
echo "⚠️  Avoid scheduling critical meetings during transition periods"
```

---

## Best Practices

### Rate limiting in batch operations:

```bash
# Add delay between calendar operations
for event in "${events[@]}"; do
  nylas calendar events create ...
  sleep 1  # Wait 1 second between requests
done
```

---

### Error handling:

```bash
# Check if event creation succeeded
if nylas calendar events create --title "Meeting" --start "..." --yes; then
  echo "✅ Event created successfully"
else
  echo "❌ Failed to create event" >&2
  # Handle error (retry, log, alert, etc.)
fi
```

---

### Timezone best practices:

1. **Always specify timezone explicitly** for multi-timezone teams
2. **Check DST transitions** when scheduling near March/November
3. **Use IANA timezone names** (not abbreviations like EST/PST)
4. **Show times in multiple timezones** when sending invites
5. **Use UTC** as common reference for global teams

---

## More Resources

- **Calendar Commands:** [Calendar Documentation](../commands/calendar.md)
- **Timezone Guide:** [Timezone Documentation](../TIMEZONE.md)
- **Troubleshooting:** [Timezone Troubleshooting](../troubleshooting/timezone.md)
- **API Reference:** https://developer.nylas.com/docs/api/v3/calendar/
