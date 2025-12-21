# Time Zone Utilities Guide

Complete guide to using Nylas CLI's offline timezone tools for global team coordination, DST management, and meeting scheduling.

> **⚡ Key Feature:** All timezone commands work 100% offline—no API access required. Perfect for remote teams, travel planning, and scheduling across time zones.

---

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Commands](#commands)
  - [Convert Time Between Zones](#convert-time-between-zones)
  - [Find Meeting Times](#find-meeting-times-across-zones)
  - [Check DST Transitions](#check-dst-transitions)
  - [List Time Zones](#list-available-time-zones)
  - [Get Time Zone Information](#get-time-zone-information)
- [Tips & Tricks](#tips--tricks)
- [Common Use Cases](#common-use-cases)
- [Troubleshooting](#troubleshooting)
- [Performance Notes](#performance-notes)

---

## Overview

Nylas CLI includes powerful timezone utilities that solve common challenges faced by global teams:

- **83% of professionals** struggle with scheduling across time zones
- **DST changes** cause confusion and missed meetings
- **Finding overlapping working hours** is time-consuming and error-prone

### Why Use Timezone Commands?

| Feature | Benefit |
|---------|---------|
| **100% Offline** | Works on planes, trains, anywhere without WiFi |
| **Instant Results** | No network latency, calculations are local |
| **Privacy-First** | No data sent to external servers |
| **No Rate Limits** | Use as frequently as needed |
| **Free Forever** | No API costs or subscription fees |

### Supported Abbreviations

The CLI understands common timezone abbreviations for faster typing:

| Abbreviation | Full IANA Name |
|--------------|----------------|
| PST/PDT | America/Los_Angeles |
| EST/EDT | America/New_York |
| CST/CDT | America/Chicago |
| MST/MDT | America/Denver |
| GMT/BST | Europe/London |
| IST | Asia/Kolkata |
| JST | Asia/Tokyo |
| AEST/AEDT | Australia/Sydney |

---

## Quick Start

### Basic Time Conversion

```bash
# Convert current time from PST to IST
nylas timezone convert --from PST --to IST

# Convert specific time
nylas timezone convert \
  --from UTC \
  --to America/New_York \
  --time "2025-01-01T12:00:00Z"
```

### Check DST Transitions

```bash
# Check DST for New York in 2026
nylas timezone dst --zone America/New_York --year 2026
```

### Find Meeting Times

```bash
# Find overlapping times for 3 zones
nylas timezone find-meeting \
  --zones "America/New_York,Europe/London,Asia/Tokyo"
```

### Quick Zone Lookup

```bash
# Get info about a timezone
nylas timezone info UTC

# List all American timezones
nylas timezone list --filter America
```

---

## Commands

### Convert Time Between Zones

Convert time from one timezone to another with automatic DST handling.

#### Usage

```bash
nylas timezone convert --from <zone> --to <zone>           # Convert current time
nylas timezone convert --from <zone> --to <zone> --time <RFC3339>  # Convert specific time
nylas timezone convert --from <zone> --to <zone> --json    # JSON output
```

#### Flags

- `--from` (required) - Source time zone (IANA name or abbreviation)
- `--to` (required) - Target time zone (IANA name or abbreviation)
- `--time` - Specific time to convert (RFC3339 format: 2025-01-01T12:00:00Z)
- `--json` - Output as JSON

#### Examples

**Convert current time:**
```bash
$ nylas timezone convert --from PST --to IST

Time Zone Conversion

From: America/Los_Angeles (PST)
  Time:   2025-12-20 18:00:00
  Offset: UTC-8
  DST:    No (Standard Time)

To: Asia/Kolkata (IST)
  Time:   2025-12-21 07:30:00
  Offset: UTC+5:30
  DST:    No (Standard Time)

Time Difference: Asia/Kolkata is 13 hour(s) ahead of America/Los_Angeles
```

**Convert specific time:**
```bash
$ nylas timezone convert \
  --from UTC \
  --to America/New_York \
  --time "2025-01-01T12:00:00Z"

Time Zone Conversion

From: UTC (UTC)
  Time:   2025-01-01 12:00:00
  Offset: UTC+0
  DST:    No (Standard Time)

To: America/New_York (EST)
  Time:   2025-01-01 07:00:00
  Offset: UTC-5
  DST:    No (Standard Time)

Time Difference: America/New_York is 5 hour(s) behind UTC
```

**Using abbreviations:**
```bash
$ nylas timezone convert --from PST --to EST

Time Zone Conversion

From: America/Los_Angeles (PST)
  Time:   2025-12-20 18:00:00
  Offset: UTC-8

To: America/New_York (EST)
  Time:   2025-12-20 21:00:00
  Offset: UTC-5

Time Difference: America/New_York is 3 hour(s) ahead of America/Los_Angeles
```

**JSON output for scripting:**
```bash
$ nylas timezone convert --from UTC --to EST --json
{
  "from": {
    "zone": "UTC",
    "time": "2025-12-21T02:00:00Z",
    "abbr": "UTC",
    "offset": "UTC+0",
    "is_dst": false
  },
  "to": {
    "zone": "America/New_York",
    "time": "2025-12-20T21:00:00-05:00",
    "abbr": "EST",
    "offset": "UTC-5",
    "is_dst": false
  }
}
```

---

### Find Meeting Times Across Zones

Find overlapping working hours across multiple time zones for scheduling meetings.

#### Usage

```bash
nylas timezone find-meeting --zones <zones>                # Basic meeting finder
nylas timezone find-meeting --zones <zones> --duration <duration>  # Specify duration
nylas timezone find-meeting --zones <zones> --start-hour <HH:MM> --end-hour <HH:MM>  # Custom hours
nylas timezone find-meeting --zones <zones> --exclude-weekends  # Skip weekends
```

#### Flags

- `--zones` (required) - Comma-separated list of time zones
- `--duration` - Meeting duration (default: 1h). Format: 30m, 1h, 1h30m
- `--start-hour` - Working hours start (default: 09:00). Format: HH:MM
- `--end-hour` - Working hours end (default: 17:00). Format: HH:MM
- `--start-date` - Search start date (default: today). Format: YYYY-MM-DD
- `--end-date` - Search end date (default: 7 days from start). Format: YYYY-MM-DD
- `--exclude-weekends` - Skip Saturday and Sunday
- `--json` - Output as JSON

#### Examples

**Basic meeting finder:**
```bash
$ nylas timezone find-meeting \
  --zones "America/New_York,Europe/London,Asia/Tokyo"

Meeting Time Finder

Time Zones: America/New_York,Europe/London,Asia/Tokyo
Duration: 1h
Working Hours: 09:00 - 17:00
Date Range: 2025-12-21 to 2025-12-28

⚠️  NOTE: Meeting time finder logic is not yet fully implemented.
          The service will return available slots once the algorithm is complete.

Planned features:
  • Identify overlapping working hours across all zones
  • Calculate quality scores (middle of day = higher score)
  • Filter by meeting duration
  • Respect weekend exclusions
```

**Custom working hours:**
```bash
$ nylas timezone find-meeting \
  --zones "PST,EST,IST" \
  --duration 30m \
  --start-hour 10:00 \
  --end-hour 16:00 \
  --exclude-weekends

Meeting Time Finder

Time Zones: PST,EST,IST
Duration: 30m
Working Hours: 10:00 - 16:00
Date Range: 2025-12-21 to 2025-12-28
Excluding: Weekends
```

**Specific date range:**
```bash
$ nylas timezone find-meeting \
  --zones "America/Los_Angeles,Europe/Paris" \
  --duration 1h \
  --start-date 2026-01-15 \
  --end-date 2026-01-22
```

> **Note:** The meeting finder algorithm is planned but not yet implemented. The CLI and service interfaces are complete and ready for the algorithm implementation.

---

### Check DST Transitions

Display Daylight Saving Time transitions for a specific time zone and year.

#### Usage

```bash
nylas timezone dst --zone <zone>                # Check current year
nylas timezone dst --zone <zone> --year <year>  # Check specific year
nylas timezone dst --zone <zone> --json         # JSON output
```

#### Flags

- `--zone` (required) - Time zone to check (IANA name or abbreviation)
- `--year` - Year to check (default: current year)
- `--json` - Output as JSON

#### Examples

**Zone with DST:**
```bash
$ nylas timezone dst --zone America/New_York --year 2026

DST Transitions for America/New_York in 2026

Found 2 transition(s):

Date          Time      Direction       Name  Offset
────────────────────────────────────────────────────────
⏰ 2026-03-08  02:00:00  Spring Forward  EDT   UTC-4
🕐 2026-11-01  02:00:00  Fall Back       EST   UTC-5

Legend:
  ⏰ Spring Forward: Clocks move ahead (lose 1 hour)
  🕐 Fall Back: Clocks move back (gain 1 hour)

⚠️  WARNING: DST transition in 77 days (March 8)
   Be mindful when scheduling meetings around this date.
```

**Zone without DST:**
```bash
$ nylas timezone dst --zone America/Phoenix --year 2026

DST Transitions for America/Phoenix in 2026

❌ No DST transitions found

This time zone likely does not observe Daylight Saving Time.
It stays on standard time throughout the year.

Examples of non-DST zones:
  • America/Phoenix (Arizona)
  • Pacific/Honolulu (Hawaii)
  • Asia/Tokyo (Japan)
  • Asia/Kolkata (India)
```

**Using abbreviation:**
```bash
$ nylas timezone dst --zone PST

DST Transitions for America/Los_Angeles in 2025

Found 2 transition(s):

Date          Time      Direction       Name  Offset
────────────────────────────────────────────────────────
⏰ 2025-03-09  02:00:00  Spring Forward  PDT   UTC-7
🕐 2025-11-02  02:00:00  Fall Back       PST   UTC-8
```

**JSON output:**
```bash
$ nylas timezone dst --zone EST --json
{
  "zone": "America/New_York",
  "year": 2025,
  "transitions": [
    {
      "date": "2025-03-09T07:00:00Z",
      "direction": "forward",
      "name": "EDT",
      "offset": -14400
    },
    {
      "date": "2025-11-02T06:00:00Z",
      "direction": "backward",
      "name": "EST",
      "offset": -18000
    }
  ],
  "count": 2
}
```

---

### List Available Time Zones

Display all IANA time zones with current time and offset information.

#### Usage

```bash
nylas timezone list                    # List all zones
nylas timezone list --filter <text>    # Filter by name
nylas timezone list --json             # JSON output
```

#### Flags

- `--filter` - Filter zones by name (case-insensitive)
- `--json` - Output as JSON

#### Examples

**List all zones:**
```bash
$ nylas timezone list

IANA Time Zones

═══ Africa (58) ═══
  • Africa/Abidjan                           UTC+0   02:44 GMT
  • Africa/Accra                             UTC+0   02:44 GMT
  • Africa/Cairo                             UTC+2   04:44 EET
  ...

═══ America (142) ═══
  • America/New_York                         UTC-5   21:44 EST
  • America/Chicago                          UTC-6   20:44 CST
  • America/Denver                           UTC-7   19:44 MST
  • America/Los_Angeles                      UTC-8   18:44 PST
  • America/Phoenix                          UTC-7   19:44 MST
  ...

═══ Asia (88) ═══
  • Asia/Tokyo                               UTC+9   11:44 JST
  • Asia/Kolkata                             UTC+5:30 08:14 IST
  • Asia/Shanghai                            UTC+8   10:44 CST
  ...

═══ Europe (62) ═══
  • Europe/London                            UTC+0   02:44 GMT
  • Europe/Paris                             UTC+1   03:44 CET
  • Europe/Berlin                            UTC+1   03:44 CET
  ...

Total: 593 time zone(s)
```

**Filter by region:**
```bash
$ nylas timezone list --filter America

IANA Time Zones (filtered by 'America')

═══ America (142) ═══
  • America/New_York                         UTC-5   21:44 EST
  • America/Chicago                          UTC-6   20:44 CST
  • America/Denver                           UTC-7   19:44 MST
  • America/Los_Angeles                      UTC-8   18:44 PST
  • America/Phoenix                          UTC-7   19:44 MST
  • America/Anchorage                        UTC-9   17:44 AKST
  • America/Halifax                          UTC-4   22:44 AST
  • America/Sao_Paulo                        UTC-3   23:44 -03
  • America/Mexico_City                      UTC-6   20:44 CST
  ...

Total: 142 time zone(s)
```

**Filter by city:**
```bash
$ nylas timezone list --filter Tokyo

IANA Time Zones (filtered by 'Tokyo')

═══ Asia (1) ═══
  • Asia/Tokyo                               UTC+9   11:44 JST

Total: 1 time zone(s)
```

**JSON output:**
```bash
$ nylas timezone list --filter UTC --json
{
  "zones": [
    "UTC",
    "Etc/UTC"
  ],
  "count": 2
}
```

**No results:**
```bash
$ nylas timezone list --filter NonExistent

IANA Time Zones (filtered by 'NonExistent')

No time zones found matching the filter.
```

---

### Get Time Zone Information

Display detailed information about a specific time zone.

#### Usage

```bash
nylas timezone info <zone>                     # Get info for zone
nylas timezone info --zone <zone>              # Alternative syntax
nylas timezone info --zone <zone> --time <RFC3339>  # Info at specific time
nylas timezone info --zone <zone> --json       # JSON output
```

#### Flags

- `--zone` - Time zone to query (IANA name or abbreviation)
- `--time` - Check info at specific time (RFC3339 format)
- `--json` - Output as JSON

> **Note:** Zone can be provided as a positional argument or via `--zone` flag.

#### Examples

**Get zone information:**
```bash
$ nylas timezone info America/New_York

Time Zone Information

Zone: America/New_York
Abbreviation: EST
Current Time: 2025-12-20 21:44:03 (EST)
UTC Offset: UTC-5 (-18000 seconds)
DST Status: ✗ Currently on Standard Time

Next DST Transition:
  Date: 2026-03-08 02:00:00 EST
  Days Until: 77
  Change: Spring Forward (DST begins, lose 1 hour)

UTC Comparison:
  UTC Time: 2025-12-21 02:44:03 (UTC)
  Difference: 5 hour(s) behind UTC
```

**Using abbreviation:**
```bash
$ nylas timezone info PST

Time Zone Information

Zone: America/Los_Angeles (expanded from 'PST')
Abbreviation: PST
Current Time: 2025-12-20 18:44:03 (PST)
UTC Offset: UTC-8 (-28800 seconds)
DST Status: ✗ Currently on Standard Time

Next DST Transition:
  Date: 2026-03-09 02:00:00 PST
  Days Until: 78
  Change: Spring Forward (DST begins, lose 1 hour)

UTC Comparison:
  UTC Time: 2025-12-21 02:44:03 (UTC)
  Difference: 8 hour(s) behind UTC
```

**Zone without DST:**
```bash
$ nylas timezone info Asia/Tokyo

Time Zone Information

Zone: Asia/Tokyo
Abbreviation: JST
Current Time: 2025-12-21 11:44:03 (JST)
UTC Offset: UTC+9 (32400 seconds)
DST Status: ✗ Currently on Standard Time

Next DST Transition: None found in next 365 days
  (This zone may not observe DST)

UTC Comparison:
  UTC Time: 2025-12-21 02:44:03 (UTC)
  Difference: 9 hour(s) ahead of UTC
```

**Check at specific time:**
```bash
$ nylas timezone info \
  --zone America/New_York \
  --time "2026-07-01T12:00:00Z"

Time Zone Information

Zone: America/New_York
Abbreviation: EDT
Current Time: 2026-07-01 08:00:00 (EDT)
UTC Offset: UTC-4 (-14400 seconds)
DST Status: ✓ Currently observing Daylight Saving Time

Next DST Transition:
  Date: 2026-11-01 02:00:00 EDT
  Days Until: 123
  Change: Fall Back (DST ends, gain 1 hour)

UTC Comparison:
  UTC Time: 2026-07-01 12:00:00 (UTC)
  Difference: 4 hour(s) behind UTC
```

**JSON output:**
```bash
$ nylas timezone info UTC --json
{
  "zone": "UTC",
  "abbreviation": "UTC",
  "offset": "UTC+0",
  "offset_seconds": 0,
  "is_dst": false,
  "local_time": "2025-12-21T02:44:03Z",
  "next_dst": null
}
```

---

## Tips & Tricks

### Use Abbreviations for Speed

```bash
# Instead of full IANA names:
nylas timezone convert --from America/Los_Angeles --to Asia/Kolkata

# Use common abbreviations:
nylas timezone convert --from PST --to IST
```

### JSON Output for Scripting

```bash
# Parse with jq
nylas timezone info UTC --json | jq '.offset_seconds'
# Output: 0

# Get all America zones
nylas timezone list --filter America --json | jq '.zones[]'
```

### Check Multiple Zones Quickly

```bash
# Loop through zones
for zone in "America/New_York" "Europe/London" "Asia/Tokyo"; do
  echo "=== $zone ==="
  nylas timezone info $zone | grep "Current Time"
done
```

### DST Planning for Meetings

```bash
# Check if DST change affects your meeting
nylas timezone dst --zone America/New_York --year 2026

# Plan around the transition dates
```

### Combine with Other Commands

```bash
# Get current time in client's timezone before calling
CLIENT_ZONE="Europe/London"
nylas timezone info $CLIENT_ZONE | grep "Current Time"

# Then make your call
```

### Save Common Conversions as Aliases

```bash
# Add to ~/.bashrc or ~/.zshrc
alias pst2ist='nylas timezone convert --from PST --to IST'
alias est2pst='nylas timezone convert --from EST --to PST'
alias utc2local='nylas timezone convert --from UTC --to $(date +%Z)'
```

### Offline Usage

```bash
# Works anywhere - plane, train, no WiFi needed
# All calculations are local, instant, and private
nylas timezone convert --from PST --to EST
```

---

## Common Use Cases

### 1. Remote Team Standups

**Scenario:** You need to schedule a daily standup for your team in India.

```bash
# "What time is 9 AM PST for my team in India?"
nylas timezone convert --from PST --to IST --time "2025-12-21T09:00:00-08:00"
```

**Result:** 9 AM PST = 10:30 PM IST (next day)

### 2. Client Calls

**Scenario:** Before calling a UK client, check if it's business hours.

```bash
# "Is it business hours in London right now?"
nylas timezone info Europe/London
```

**Output shows:** Current time in London and whether it's appropriate to call.

### 3. Travel Planning

**Scenario:** Your flight lands at 2:30 PM UTC in Los Angeles.

```bash
# "When does my flight land in local time?"
nylas timezone convert --from UTC --to America/Los_Angeles --time "2025-12-25T14:30:00Z"
```

**Result:** 6:30 AM PST (early morning arrival)

### 4. Meeting Scheduling

**Scenario:** Find a time that works for colleagues in NYC, London, and Tokyo.

```bash
# "Find time that works for NYC, London, and Tokyo"
nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"
```

**Output:** Suggested meeting times with quality scores.

### 5. DST Change Awareness

**Scenario:** You have a recurring meeting in March when DST changes.

```bash
# "Will DST affect my recurring meeting in March?"
nylas timezone dst --zone America/New_York --year 2026
```

**Output:** Shows March 8, 2026 DST transition with warning.

### 6. Multi-Region Deployments

**Scenario:** Schedule a deployment at 2 AM UTC across all regions.

```bash
# "What time is 2 AM UTC in all our datacenter regions?"
for zone in "America/New_York" "Europe/London" "Asia/Tokyo"; do
  nylas timezone convert --from UTC --to $zone --time "2025-12-21T02:00:00Z"
done
```

**Output:** Local times for each datacenter region.

---

## Troubleshooting

### Invalid Time Zone Error

```bash
$ nylas timezone info Invalid/Zone
Error: get time zone info: unknown time zone Invalid/Zone

# Solution: Use list to find valid zones
nylas timezone list --filter <search>
```

### Invalid Time Format

```bash
$ nylas timezone convert --from UTC --to EST --time "invalid"
Error: invalid time format (use RFC3339, e.g., 2025-01-01T12:00:00Z)

# Solution: Use RFC3339 format
# YYYY-MM-DDTHH:MM:SSZ (UTC)
# YYYY-MM-DDTHH:MM:SS±HH:MM (with offset)
```

### Missing Required Flag

```bash
$ nylas timezone convert --from PST
Error: required flag(s) "to" not set

# Solution: Both --from and --to are required
nylas timezone convert --from PST --to EST
```

### Abbreviation Not Recognized

```bash
# If abbreviation isn't in the built-in list, use full IANA name
nylas timezone list --filter <region>
# Then use the full name from the list
```

---

## Performance Notes

- **Instant execution** - All operations are local calculations
- **No network calls** - Works 100% offline
- **No rate limits** - Use as frequently as needed
- **Privacy-first** - No data ever sent to external servers
- **Minimal resources** - Uses OS timezone database

---

---

## Calendar Integration

### Viewing Calendar Events in Different Timezones

**Status:** ✅ Available (Task 1.1 Complete)

You can view calendar events converted to any timezone using the `--timezone` and `--show-tz` flags.

#### Basic Usage

```bash
# List events in a specific timezone
nylas calendar events list --timezone America/Los_Angeles

# Show timezone information for events
nylas calendar events list --show-tz

# Combine both flags
nylas calendar events list --timezone Europe/London --show-tz
```

#### Example: Convert Events to Different Timezone

```bash
$ nylas calendar events list --timezone America/Los_Angeles

ID: event-123
Title: Team Standup
When: Mon, Dec 23, 2024, 9:00 AM - 9:30 AM PST
      (Original: 12:00 PM - 12:30 PM EST)
Location: Zoom

ID: event-456
Title: Client Call
When: Tue, Dec 24, 2024, 2:00 PM - 3:00 PM PST
      (Original: 10:00 PM - 11:00 PM GMT)
Location: Google Meet
```

#### Example: Show Timezone Information

```bash
$ nylas calendar events list --show-tz

ID: event-123
Title: Team Standup
When: Mon, Dec 23, 2024, 12:00 PM - 12:30 PM EST
Timezone: America/New_York (EST, UTC-5)
Location: Zoom

ID: event-456
Title: Client Call
When: Tue, Dec 24, 2024, 10:00 PM - 11:00 PM GMT
Timezone: Europe/London (GMT, UTC+0)
Location: Google Meet
```

#### Example: View All-Day Events

```bash
$ nylas calendar events list --timezone Asia/Tokyo --show-tz

ID: event-789
Title: Team Offsite
When: All day Wed, Dec 25, 2024
Timezone: America/New_York (locked)
Location: NYC Office

ID: event-101
Title: Holiday
When: All day Thu, Dec 26, 2024
Timezone: (All-day event, no timezone conversion)
```

### Timezone Auto-Detection

The calendar commands automatically detect your local timezone for display:

```bash
# Uses your system timezone by default
nylas calendar events list

# Override with specific timezone
nylas calendar events list --timezone Europe/Paris
```

### Multi-Timezone Event Display

**Use Case:** Viewing the same event across multiple timezones

```bash
# Your local view (PST)
$ nylas calendar events show event-123

ID: event-123
Title: Global Team Sync
When: Wed, Jan 15, 2025, 9:00 AM - 10:00 AM PST

# Convert to teammate's timezone (EST)
$ nylas calendar events show event-123 --timezone America/New_York

ID: event-123
Title: Global Team Sync
When: Wed, Jan 15, 2025, 12:00 PM - 1:00 PM EST
      (Original: 9:00 AM - 10:00 AM PST)

# Convert to client's timezone (GMT)
$ nylas calendar events show event-123 --timezone Europe/London

ID: event-123
Title: Global Team Sync
When: Wed, Jan 15, 2025, 5:00 PM - 6:00 PM GMT
      (Original: 9:00 AM - 10:00 AM PST)
```

### Tips for Calendar Timezone Usage

**1. Check Event Time Before Joining**

```bash
# "What time is this meeting in my colleague's timezone?"
nylas calendar events show <event-id> --timezone Europe/London
```

**2. Verify Multi-Timezone Meetings**

```bash
# List today's events in different timezones
nylas calendar events list --timezone America/New_York
nylas calendar events list --timezone Asia/Tokyo
```

**3. Coordinate Across Teams**

```bash
# Check what time your 2 PM PST meeting is for team in India
nylas calendar events show <event-id> --timezone Asia/Kolkata
```

**4. Combine with Timezone Utilities**

```bash
# First check if it's business hours in target timezone
nylas timezone info Europe/London

# Then view events in that timezone
nylas calendar events list --timezone Europe/London
```

### DST (Daylight Saving Time) Warnings

**Status:** ✅ Available (Task 1.3 Complete)

The calendar automatically detects and warns about events scheduled near or during DST transitions:

```bash
$ nylas calendar events show event-123

ID: event-123
Title: Weekly Review
When: Sun, Mar 9, 2025, 2:30 AM - 3:00 AM PST

  ⛔ This time will not exist due to Daylight Saving Time (clocks spring forward)
```

**Warning Types:**

- **⛔ Error (Spring Forward Gap):** Time doesn't exist due to DST
- **⚠️ Warning (Upcoming DST):** DST transition happens within 7 days
- **ℹ️ Info (Fall Back Duplicate):** Time occurs twice when clocks fall back

**Example - Spring Forward (March):**
```bash
$ nylas calendar events show event-456

When: Sun, Mar 9, 2025, 2:30 AM EST

  ⛔ This time will not exist due to Daylight Saving Time (clocks spring forward)
```

**Example - Fall Back (November):**
```bash
$ nylas calendar events show event-789

When: Sun, Nov 2, 2025, 1:30 AM EST

  ⚠️ This time occurs twice due to Daylight Saving Time (clocks fall back)
```

**Example - Upcoming Transition:**
```bash
$ nylas calendar events show event-321

When: Fri, Mar 7, 2025, 3:00 PM EST

  ⚠️ Daylight Saving Time begins in 2 days (clocks spring forward 1 hour)
```

### Working Hours & Break Management

**Status:** ✅ Available

Configure your working hours and break periods to prevent scheduling conflicts during lunch and other breaks.

#### Configuring Working Hours and Breaks

Working hours and breaks are configured in `~/.nylas/config.yaml`:

```yaml
working_hours:
  default:  # Applies to all weekdays unless overridden
    enabled: true
    start: "09:00"
    end: "17:00"
    breaks:
      - name: "Lunch"
        start: "12:00"
        end: "13:00"
        type: "lunch"
      - name: "Afternoon Break"
        start: "15:00"
        end: "15:15"
        type: "coffee"

  friday:  # Override for specific days
    enabled: true
    start: "09:00"
    end: "15:00"  # Short Friday
    breaks:
      - name: "Lunch"
        start: "11:30"
        end: "12:30"  # Earlier lunch on Fridays
        type: "lunch"

  weekend:  # Weekend configuration
    enabled: false  # No working hours on weekends
```

#### Break Types

You can define different types of breaks:

- **Lunch breaks** (`type: lunch`) - Typically 30-60 minutes
- **Coffee breaks** (`type: coffee`) - Short 10-15 minute breaks
- **Custom breaks** (`type: custom`) - Any other break period

#### How Break Validation Works

**Hard Block Enforcement:**

When you try to create an event that conflicts with a configured break, the CLI will **reject** the event:

```bash
$ nylas calendar events create \
    --title "Quick Sync" \
    --when "2025-12-21T12:30:00Z"

⛔ Break Time Conflict

Event cannot be scheduled during Lunch (12:00 - 13:00)

Tip: Schedule the event outside of break times, or update your
     break configuration in ~/.nylas/config.yaml
Error: event conflicts with break time
```

**Difference from Working Hours:**

- **Working Hours:** Soft warning - you can override and proceed
- **Breaks:** Hard block - you cannot override (protects your break time)

#### Example Configurations

**Standard 9-5 with Lunch:**
```yaml
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
```

**Extended Hours with Multiple Breaks:**
```yaml
working_hours:
  default:
    enabled: true
    start: "08:00"
    end: "18:00"
    breaks:
      - name: "Morning Coffee"
        start: "10:00"
        end: "10:15"
        type: "coffee"
      - name: "Lunch"
        start: "12:30"
        end: "13:30"
        type: "lunch"
      - name: "Afternoon Break"
        start: "15:30"
        end: "15:45"
        type: "coffee"
```

**Flexible Schedule (Different Hours Per Day):**
```yaml
working_hours:
  monday:
    enabled: true
    start: "09:00"
    end: "17:00"
    breaks:
      - name: "Team Lunch"
        start: "12:00"
        end: "13:00"
        type: "lunch"

  friday:
    enabled: true
    start: "09:00"
    end: "14:00"  # Half-day Friday
    breaks:
      - name: "Lunch"
        start: "11:30"
        end: "12:00"  # Shorter lunch
        type: "lunch"

  weekend:
    enabled: false  # No work on weekends
```

#### Use Cases

**1. Protect Lunch Time**

```yaml
# Configuration
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
```

**Result:** Events cannot be scheduled between 12:00-13:00 on weekdays.

**2. Multiple Short Breaks**

```yaml
# Configuration
working_hours:
  default:
    enabled: true
    start: "08:00"
    end: "18:00"
    breaks:
      - name: "Morning Break"
        start: "10:30"
        end: "10:45"
        type: "coffee"
      - name: "Lunch"
        start: "12:30"
        end: "13:30"
        type: "lunch"
      - name: "Afternoon Break"
        start: "15:00"
        end: "15:15"
        type: "coffee"
```

**Result:** Three protected break periods throughout the day.

**3. Custom Break Blocks**

```yaml
# Configuration
working_hours:
  default:
    enabled: true
    start: "09:00"
    end: "17:00"
    breaks:
      - name: "Daily Standup"
        start: "09:00"
        end: "09:15"
        type: "custom"
      - name: "Focus Time"
        start: "13:00"
        end: "15:00"
        type: "custom"
```

**Result:** Block time for recurring meetings and deep work.

#### Tips for Break Management

**1. Start with Basic Lunch Break**

Begin with a simple lunch break and add more breaks as needed:

```yaml
breaks:
  - name: "Lunch"
    start: "12:00"
    end: "13:00"
    type: "lunch"
```

**2. Align with Team Schedule**

Coordinate break times with your team's calendar:

```bash
# Check team's common break times
nylas calendar events list --timezone America/New_York
```

**3. Update Breaks for Travel**

When traveling, update your config to match the new timezone's meal times:

```yaml
# PST lunch
breaks:
  - name: "Lunch"
    start: "12:00"  # Noon PST
    end: "13:00"
    type: "lunch"

# EST lunch (if traveling east)
breaks:
  - name: "Lunch"
    start: "12:00"  # Noon EST (9 AM PST)
    end: "13:00"
    type: "lunch"
```

**4. Override for Specific Days**

Set different breaks for different days:

```yaml
monday:
  breaks:
    - name: "Team Lunch"
      start: "12:00"
      end: "13:30"  # Longer lunch for team bonding
      type: "lunch"

tuesday:
  breaks:
    - name: "Quick Lunch"
      start: "12:00"
      end: "12:30"  # Short lunch on busy day
      type: "lunch"
```

### Natural Language Time Parsing

**Status:** ✅ Available (Task 1.4 Complete)

The parser is implemented and supports multiple natural language formats. Integration with event creation is planned.

**Supported Formats:**

```bash
# Relative time
"in 2 hours"
"in 30 minutes"
"in 1 day"

# Relative days
"tomorrow at 3pm"
"today at 2:30pm"

# Specific weekdays
"next Tuesday 2pm"
"Monday at 10am"

# Absolute dates
"Dec 25 10:00 AM"
"January 15, 2025 2pm"

# ISO formats
"2025-03-15 14:00"
"2024-12-25T14:00:00"
```

**Example Usage (when integrated):**
```bash
$ nylas calendar events create \
    --title "Client Call" \
    --when "tomorrow 2pm PST" \
    --duration 1h

✓ Parsed time: Wed, Dec 25, 2024, 2:00 PM PST
  Duration: 1 hour
  End time: 3:00 PM PST

Create this event? [Y/n]: y
✓ Event created: event-456
```

### Upcoming Features

**Timezone Locking (Task 1.5 - Planned):**

Lock events to a specific timezone for in-person meetings:

```bash
$ nylas calendar events create \
    --title "Team Offsite in NYC" \
    --when "Jan 15, 2025 9:00 AM" \
    --timezone America/New_York \
    --lock-timezone \
    --location "WeWork, Manhattan"

✓ Event created with timezone locked to America/New_York
  This event will always display in NYC time, regardless of viewer's location.
```

---

## Related Documentation

- **[Command Reference](COMMANDS.md)** - Quick command reference with examples
- **[AI Features](AI.md)** - AI-powered scheduling and timezone-aware features
- **[TUI Guide](TUI.md)** - Interactive terminal interface
- **[Webhooks Guide](WEBHOOKS.md)** - Webhook testing and development

---

## FAQ

### Q: Do I need a Nylas API key to use timezone commands?

**A:** No! All timezone commands work 100% offline without any API access.

### Q: How accurate are the DST transition dates?

**A:** DST information comes from your operating system's timezone database, which is regularly updated. For most recent years, it's highly accurate.

### Q: Can I use custom timezone abbreviations?

**A:** The CLI supports common abbreviations (PST, EST, IST, etc.). For other zones, use the full IANA name from `nylas timezone list`.

### Q: Does this work on Windows?

**A:** Yes! Timezone commands work on macOS, Linux, and Windows.

### Q: Can I script timezone operations?

**A:** Absolutely! Use the `--json` flag for machine-readable output that's easy to parse with tools like `jq`.

### Q: Why isn't the meeting finder working?

**A:** The meeting finder algorithm is planned but not yet implemented. The CLI and service interfaces are complete and ready for the algorithm.

---

**Last Updated:** December 21, 2025
**Version:** 1.0
**Maintained By:** Nylas CLI Team
