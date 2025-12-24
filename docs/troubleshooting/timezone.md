# Timezone Troubleshooting

Comprehensive guide for resolving timezone-related issues.

---

## Table of Contents

- [Quick Diagnostics](#quick-diagnostics)
- [Common Issues](#common-issues)
- [DST Issues](#dst-issues)
- [Timezone Conversion](#timezone-conversion)
- [Calendar Timezone Display](#calendar-timezone-display)
- [Meeting Time Finder](#meeting-time-finder)

---

## Quick Diagnostics

```bash
# List available timezones
nylas timezone list

# Get timezone info
nylas timezone info America/New_York

# Check DST transitions
nylas timezone dst --zone America/New_York --year 2025

# Test conversion
nylas timezone convert --from UTC --to America/Los_Angeles
```

---

## Common Issues

### Issue: "Timezone not found" or "Invalid timezone"

**Symptoms:**
```
Error: invalid timezone: New York
Error: timezone not found
```

**Causes:**
- Incorrect timezone format
- Abbreviated timezone instead of full name
- Typo in timezone name

**Solutions:**

1. **Use full IANA timezone names:**
```bash
# ✅ Correct - Full IANA name
nylas timezone info America/New_York
nylas timezone info Europe/London
nylas timezone info Asia/Tokyo

# ❌ Wrong - Abbreviations
nylas timezone info EST  # Use America/New_York
nylas timezone info PST  # Use America/Los_Angeles
nylas timezone info GMT  # Use UTC or Europe/London

# ❌ Wrong - City names without continent
nylas timezone info New_York  # Should be America/New_York
```

2. **Search for timezone:**
```bash
# Find timezone by keyword
nylas timezone list --filter America
nylas timezone list --filter Europe
nylas timezone list --filter New_York

# Get exact name from results
nylas timezone list --filter Tokyo
# Result: Asia/Tokyo
```

3. **Common timezone mappings:**

| Abbreviation | IANA Timezone |
|-------------|---------------|
| EST/EDT | America/New_York |
| CST/CDT | America/Chicago |
| MST/MDT | America/Denver |
| PST/PDT | America/Los_Angeles |
| GMT | UTC or Europe/London |
| BST | Europe/London |
| IST (India) | Asia/Kolkata |
| JST | Asia/Tokyo |
| AEST | Australia/Sydney |

---

### Issue: Incorrect time conversion

**Symptoms:**
- Converted time is off by 1 hour
- Conversion doesn't account for DST
- Wrong offset applied

**Solutions:**

1. **Verify both timezones:**
```bash
# Check source timezone
nylas timezone info America/New_York

# Check target timezone
nylas timezone info Europe/London

# Verify conversion
nylas timezone convert \
  --from America/New_York \
  --to Europe/London \
  --time "2025-03-15 14:00"
```

2. **DST awareness:**
```bash
# Check if DST is active
nylas timezone dst --zone America/New_York --year 2025

# DST affects offset:
# Standard time: EST = UTC-5
# Daylight time: EDT = UTC-4

# Conversion automatically accounts for DST
nylas timezone convert \
  --from America/New_York \
  --to UTC \
  --time "2025-03-09 03:00"  # Just after DST begins
```

3. **Specify exact date/time:**
```bash
# Include full date for accurate conversion
# (DST changes throughout the year)

# ✅ Complete date/time
nylas timezone convert \
  --from PST \
  --to EST \
  --time "2025-03-15 14:00"

# ✅ With timezone names
nylas timezone convert \
  --from America/Los_Angeles \
  --to America/New_York \
  --time "2025-03-15 14:00"
```

---

## DST Issues

### Understanding DST transitions:

**Spring Forward (Clocks ahead):**
- Occurs in March (Northern Hemisphere)
- Clocks jump forward 1 hour
- Time gap: 2:00 AM → 3:00 AM
- Times like 2:30 AM **don't exist**

**Fall Back (Clocks back):**
- Occurs in November (Northern Hemisphere)
- Clocks fall back 1 hour
- Time repeat: 2:00 AM → 1:00 AM
- Times like 1:30 AM happen **twice**

### Issue: Event time doesn't exist due to DST

**Symptoms:**
```
⚠️ Warning: This time will not exist due to Daylight Saving Time
DST transition creates invalid time
```

**Example:**
```bash
# In America/New_York, on March 9, 2025:
# 2:00 AM jumps to 3:00 AM (spring forward)
# So 2:30 AM doesn't exist!

nylas calendar events create \
  --title "Early Meeting" \
  --start "2025-03-09 02:30" \
  --timezone America/New_York

# ⚠️ Warning: Time 02:30 will not exist
```

**Solutions:**

1. **Avoid DST transition times:**
```bash
# Check DST transitions first
nylas timezone dst --zone America/New_York --year 2025

# Output shows:
# Spring forward: Mar 9, 2025 at 2:00 AM → 3:00 AM
# Fall back: Nov 2, 2025 at 2:00 AM → 1:00 AM

# Schedule before or after transition
nylas calendar events create \
  --title "Early Meeting" \
  --start "2025-03-09 03:00"  # After transition ✅
```

2. **Use different timezone:**
```bash
# Use UTC (no DST)
nylas calendar events create \
  --title "Meeting" \
  --start "2025-03-09 07:30" \
  --timezone UTC
```

3. **CLI will warn you:**
```bash
# CLI automatically detects DST issues
# Shows warning before creating event
# Allows you to adjust time
```

---

### Issue: Event displays wrong time after DST change

**Symptoms:**
- Event was at 2:00 PM, now shows 1:00 PM
- Time changed after DST transition
- Offset is different

**Causes:**
- Event was stored with fixed offset
- Timezone rule changed
- DST transition occurred

**Solutions:**

1. **Check event timezone:**
```bash
# View event with timezone info
nylas calendar events show <event-id> --show-tz

# Verify timezone is correct
# Update if needed
```

2. **Understand timezone storage:**
   - Events stored with timezone (e.g., "America/New_York") automatically adjust for DST ✅
   - Events stored with fixed offset (e.g., "UTC-5") don't adjust ❌

3. **Use named timezones, not offsets:**
```bash
# ✅ Good - Adjusts for DST
--timezone America/New_York

# ❌ Avoid - Fixed offset
--timezone UTC-5
```

---

## Timezone Conversion

### Converting times between zones:

```bash
# Basic conversion
nylas timezone convert \
  --from America/New_York \
  --to Europe/London

# With specific time
nylas timezone convert \
  --from America/New_York \
  --to Europe/London \
  --time "2025-03-15 14:00"

# Multiple conversions
nylas timezone convert \
  --from America/Los_Angeles \
  --to "Europe/London,Asia/Tokyo,Australia/Sydney"
```

### Conversion tips:

1. **Always use IANA names**
2. **Include full date** (for DST accuracy)
3. **Check DST status** if near transition
4. **Use --show-offset** to see UTC offset

---

## Calendar Timezone Display

### Viewing events in different timezones:

```bash
# List events in specific timezone
nylas calendar events list --timezone America/Los_Angeles

# Show timezone abbreviations
nylas calendar events list --show-tz

# Both combined
nylas calendar events list \
  --timezone Europe/London \
  --show-tz

# View specific event in different timezone
nylas calendar events show <event-id> \
  --timezone Asia/Tokyo
```

### Issue: Events show in wrong timezone

**Solutions:**

1. **Specify explicit timezone:**
```bash
# Force display in specific timezone
nylas calendar events list --timezone America/New_York
```

2. **Check system timezone:**
```bash
# Events default to system timezone
# Check system timezone
date +%Z

# Override with --timezone flag
```

3. **Use --show-tz to see abbreviations:**
```bash
# Shows timezone info for each event
nylas calendar events list --show-tz

# Example output:
# 2:00 PM EST (America/New_York)
# 3:00 PM PST (America/Los_Angeles)
```

---

## Meeting Time Finder

### Finding optimal meeting times:

```bash
# Basic usage
nylas timezone find-meeting \
  --zones "America/New_York,Europe/London,Asia/Tokyo"

# With preferences
nylas timezone find-meeting \
  --zones "America/New_York,Europe/London,Asia/Tokyo" \
  --duration 60 \
  --earliest 9 \
  --latest 17
```

### Issue: No meeting times found

**Symptoms:**
- "No suitable meeting times found"
- All suggested times are outside working hours
- Times conflict with preferences

**Solutions:**

1. **Expand time range:**
```bash
# Allow earlier/later times
nylas timezone find-meeting \
  --zones "..." \
  --earliest 7 \
  --latest 20
```

2. **Reduce duration:**
```bash
# Shorter meetings have more slots
nylas timezone find-meeting \
  --zones "..." \
  --duration 30  # Instead of 60
```

3. **Consider async communication:**
```bash
# If no good time exists across timezones
# Consider asynchronous communication instead
```

4. **Check working hours in each zone:**
```bash
# Verify working hours overlap
# Example: New York (9-5) + Tokyo (9-5) = no overlap!
# NY: 9 AM = Tokyo: 11 PM (outside working hours)
```

---

## Advanced Troubleshooting

### Debugging timezone issues:

```bash
# Get detailed timezone info
nylas timezone info America/New_York

# Shows:
# - Current offset
# - DST status
# - Next DST transition
# - Abbreviation

# Check DST history
nylas timezone dst --zone America/New_York --year 2024
nylas timezone dst --zone America/New_York --year 2025

# Verify conversion
nylas timezone convert \
  --from America/New_York \
  --to UTC \
  --time "2025-03-09 02:30"  # DST transition

# Should show warning if time doesn't exist
```

### Common timezone data issues:

1. **Outdated timezone database:**
   - Timezone rules change over time
   - Governments modify DST rules
   - CLI uses system timezone data
   - Keep system updated

2. **Historical vs future DST:**
   - Past DST rules may differ
   - Future rules may change
   - CLI uses current known rules

3. **Timezone aliases:**
   - Some zones have multiple names
   - Use canonical IANA name
   - Check with `nylas timezone list`

---

## Best Practices

### For event creation:

1. **Use named timezones** (not offsets)
2. **Check DST transitions** near March/November
3. **Verify time exists** with `--show-tz`
4. **Test conversion** before creating events

### For multi-timezone coordination:

1. **Use UTC** as common reference
2. **Show multiple timezones** with `--timezone` flag
3. **Warn participants** about DST changes
4. **Use meeting finder** for optimal times

### For automation:

1. **Handle DST transitions** in scripts
2. **Use full IANA names** in code
3. **Test around DST dates** (March, November)
4. **Log timezone info** for debugging

---

## Timezone Resources

**List all timezones:**
```bash
nylas timezone list
```

**Search by region:**
```bash
nylas timezone list --filter America
nylas timezone list --filter Europe
nylas timezone list --filter Asia
```

**Get timezone details:**
```bash
nylas timezone info <timezone-name>
```

**Check DST:**
```bash
nylas timezone dst --zone <timezone> --year 2025
```

---

## Still Having Issues?

1. **Check FAQ:** [FAQ.md](../FAQ.md)
2. **Review timezone docs:** [TIMEZONE.md](../TIMEZONE.md)
3. **List available timezones:** `nylas timezone list`
4. **Report issue:** https://github.com/mqasimca/nylas/issues
