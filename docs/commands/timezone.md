## Time Zone Utilities

Offline time zone conversion and meeting scheduling tools. Works 100% offline—no API access required.

> **📚 For comprehensive timezone documentation, see the [Timezone Guide](TIMEZONE.md).**

> **💡 Pro Tip:** All timezone commands work instantly without network calls. Perfect for remote teams, travel planning, and global coordination.

### Quick Examples

```bash
# Convert current time between zones
nylas timezone convert --from PST --to IST

# Check DST transitions for planning
nylas timezone dst --zone America/New_York --year 2026

# Find meeting times across multiple zones
nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"

# List all available time zones
nylas timezone list --filter America

# Get detailed zone information
nylas timezone info UTC
```

---

### Convert Time Between Zones

Convert time from one timezone to another with automatic DST handling.

```bash
nylas timezone convert --from <zone> --to <zone>           # Convert current time
nylas timezone convert --from <zone> --to <zone> --time <RFC3339>  # Convert specific time
nylas timezone convert --from <zone> --to <zone> --json    # JSON output
```

**Flags:**
- `--from` (required) - Source time zone (IANA name or abbreviation)
- `--to` (required) - Target time zone (IANA name or abbreviation)
- `--time` - Specific time to convert (RFC3339 format: 2025-01-01T12:00:00Z)
- `--json` - Output as JSON

**Supported Abbreviations:**
- PST/PDT → America/Los_Angeles
- EST/EDT → America/New_York
- CST/CDT → America/Chicago
- MST/MDT → America/Denver
- GMT/BST → Europe/London
- IST → Asia/Kolkata
- JST → Asia/Tokyo
- AEST/AEDT → Australia/Sydney

**Example: Convert current time**
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

**Example: Convert specific time**
```bash
$ nylas timezone convert --from UTC --to America/New_York --time "2025-01-01T12:00:00Z"

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

**Example: Using abbreviations**
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

**Example: JSON output for scripting**
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

```bash
nylas timezone find-meeting --zones <zones>                # Basic meeting finder
nylas timezone find-meeting --zones <zones> --duration <duration>  # Specify duration
nylas timezone find-meeting --zones <zones> --start-hour <HH:MM> --end-hour <HH:MM>  # Custom hours
nylas timezone find-meeting --zones <zones> --exclude-weekends  # Skip weekends
```

**Flags:**
- `--zones` (required) - Comma-separated list of time zones
- `--duration` - Meeting duration (default: 1h). Format: 30m, 1h, 1h30m
- `--start-hour` - Working hours start (default: 09:00). Format: HH:MM
- `--end-hour` - Working hours end (default: 17:00). Format: HH:MM
- `--start-date` - Search start date (default: today). Format: YYYY-MM-DD
- `--end-date` - Search end date (default: 7 days from start). Format: YYYY-MM-DD
- `--exclude-weekends` - Skip Saturday and Sunday
- `--json` - Output as JSON

**Example: Basic meeting finder**
```bash
$ nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"

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

**Example: Custom working hours**
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

**Example: Specific date range**
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

```bash
nylas timezone dst --zone <zone>                # Check current year
nylas timezone dst --zone <zone> --year <year>  # Check specific year
nylas timezone dst --zone <zone> --json         # JSON output
```

**Flags:**
- `--zone` (required) - Time zone to check (IANA name or abbreviation)
- `--year` - Year to check (default: current year)
- `--json` - Output as JSON

**Example: Zone with DST**
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

**Example: Zone without DST**
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

**Example: Using abbreviation**
```bash
$ nylas timezone dst --zone PST

DST Transitions for America/Los_Angeles in 2025

Found 2 transition(s):

Date          Time      Direction       Name  Offset
────────────────────────────────────────────────────────
⏰ 2025-03-09  02:00:00  Spring Forward  PDT   UTC-7
🕐 2025-11-02  02:00:00  Fall Back       PST   UTC-8
```

**Example: JSON output**
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

```bash
nylas timezone list                    # List all zones
nylas timezone list --filter <text>    # Filter by name
nylas timezone list --json             # JSON output
```

**Flags:**
- `--filter` - Filter zones by name (case-insensitive)
- `--json` - Output as JSON

**Example: List all zones**
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

**Example: Filter by region**
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

**Example: Filter by city**
```bash
$ nylas timezone list --filter Tokyo

IANA Time Zones (filtered by 'Tokyo')

═══ Asia (1) ═══
  • Asia/Tokyo                               UTC+9   11:44 JST

Total: 1 time zone(s)
```

**Example: JSON output**
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

**Example: No results**
```bash
$ nylas timezone list --filter NonExistent

IANA Time Zones (filtered by 'NonExistent')

No time zones found matching the filter.
```

---

### Get Time Zone Information

Display detailed information about a specific time zone.

```bash
nylas timezone info <zone>                     # Get info for zone
nylas timezone info --zone <zone>              # Alternative syntax
nylas timezone info --zone <zone> --time <RFC3339>  # Info at specific time
nylas timezone info --zone <zone> --json       # JSON output
```

**Flags:**
- `--zone` - Time zone to query (IANA name or abbreviation)
- `--time` - Check info at specific time (RFC3339 format)
- `--json` - Output as JSON

> **Note:** Zone can be provided as a positional argument or via `--zone` flag.

**Example: Get zone information**
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

**Example: Using abbreviation**
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

**Example: Zone without DST**
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

**Example: Check at specific time**
```bash
$ nylas timezone info --zone America/New_York --time "2026-07-01T12:00:00Z"

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

**Example: JSON output**
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

**Example: Using flag instead of positional arg**
```bash
$ nylas timezone info --zone Europe/London

Time Zone Information

Zone: Europe/London
Abbreviation: GMT
Current Time: 2025-12-21 02:44:03 (GMT)
UTC Offset: UTC+0 (0 seconds)
DST Status: ✗ Currently on Standard Time

Next DST Transition:
  Date: 2026-03-29 01:00:00 GMT
  Days Until: 97
  Change: Spring Forward (DST begins, lose 1 hour)

UTC Comparison:
  UTC Time: 2025-12-21 02:44:03 (UTC)
  Difference: Same as UTC
```

---

### Tips & Tricks

**Use Abbreviations for Speed**
```bash
# Instead of full IANA names:
nylas timezone convert --from America/Los_Angeles --to Asia/Kolkata

# Use common abbreviations:
nylas timezone convert --from PST --to IST
```

**JSON Output for Scripting**
```bash
# Parse with jq
nylas timezone info UTC --json | jq '.offset_seconds'
# Output: 0

# Get all America zones
nylas timezone list --filter America --json | jq '.zones[]'
```

**Check Multiple Zones Quickly**
```bash
# Loop through zones
for zone in "America/New_York" "Europe/London" "Asia/Tokyo"; do
  echo "=== $zone ==="
  nylas timezone info $zone | grep "Current Time"
done
```

**DST Planning for Meetings**
```bash
# Check if DST change affects your meeting
nylas timezone dst --zone America/New_York --year 2026

# Plan around the transition dates
```

**Combine with Other Commands**
```bash
# Get current time in client's timezone before calling
CLIENT_ZONE="Europe/London"
nylas timezone info $CLIENT_ZONE | grep "Current Time"

# Then make your call
```

**Save Common Conversions as Aliases**
```bash
# Add to ~/.bashrc or ~/.zshrc
alias pst2ist='nylas timezone convert --from PST --to IST'
alias est2pst='nylas timezone convert --from EST --to PST'
alias utc2local='nylas timezone convert --from UTC --to $(date +%Z)'
```

**Offline Usage**
```bash
# Works anywhere - plane, train, no WiFi needed
# All calculations are local, instant, and private
nylas timezone convert --from PST --to EST
```

---

### Common Use Cases

**1. Remote Team Standups**
```bash
# "What time is 9 AM PST for my team in India?"
nylas timezone convert --from PST --to IST --time "2025-12-21T09:00:00-08:00"
```

**2. Client Calls**
```bash
# "Is it business hours in London right now?"
nylas timezone info Europe/London
```

**3. Travel Planning**
```bash
# "When does my flight land in local time?"
nylas timezone convert --from UTC --to America/Los_Angeles --time "2025-12-25T14:30:00Z"
```

**4. Meeting Scheduling**
```bash
# "Find time that works for NYC, London, and Tokyo"
nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"
```

**5. DST Change Awareness**
```bash
# "Will DST affect my recurring meeting in March?"
nylas timezone dst --zone America/New_York --year 2026
```

**6. Multi-Region Deployments**
```bash
# "What time is 2 AM UTC in all our datacenter regions?"
for zone in "America/New_York" "Europe/London" "Asia/Tokyo"; do
  nylas timezone convert --from UTC --to $zone --time "2025-12-21T02:00:00Z"
done
```

---

### Troubleshooting

**Invalid Time Zone Error**
```bash
$ nylas timezone info Invalid/Zone
Error: get time zone info: unknown time zone Invalid/Zone

# Use list to find valid zones:
nylas timezone list --filter <search>
```

**Invalid Time Format**
```bash
$ nylas timezone convert --from UTC --to EST --time "invalid"
Error: invalid time format (use RFC3339, e.g., 2025-01-01T12:00:00Z)

# Use RFC3339 format:
# YYYY-MM-DDTHH:MM:SSZ (UTC)
# YYYY-MM-DDTHH:MM:SS±HH:MM (with offset)
```

**Missing Required Flag**
```bash
$ nylas timezone convert --from PST
Error: required flag(s) "to" not set

# Both --from and --to are required
nylas timezone convert --from PST --to EST
```

**Abbreviation Not Recognized**
```bash
# If abbreviation isn't in the built-in list, use full IANA name
nylas timezone list --filter <region>
# Then use the full name from the list
```

---

### Performance Notes

- **Instant execution** - All operations are local calculations
- **No network calls** - Works 100% offline
- **No rate limits** - Use as frequently as needed
- **Privacy-first** - No data ever sent to external servers
- **Minimal resources** - Uses OS timezone database

---

### Related Commands

- `nylas auth detect` - Detect email provider timezone
- `nylas calendar list` - View events (which may have timezone info)
- `nylas tui` - Interactive terminal interface

---

