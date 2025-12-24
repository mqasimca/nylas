# Time Zone Utilities

Offline timezone tools for global team coordination, DST management, and meeting scheduling.

> **⚡ All timezone commands work 100% offline—no API access required.**

---

## Commands

### Convert Time Between Zones

```bash
nylas timezone convert --from PST --to IST
nylas timezone convert --from UTC --to America/New_York --time "2025-01-01T12:00:00Z"
```

**Flags:** `--from`, `--to`, `--time`, `--json`

---

### Check DST Transitions

```bash
nylas timezone dst --zone America/New_York --year 2026
```

**Flags:** `--zone`, `--year`, `--json`

---

### Find Meeting Times

```bash
nylas timezone find-meeting --zones "America/New_York,Europe/London,Asia/Tokyo"
```

**Flags:** `--zones`, `--duration`, `--start-hour`, `--end-hour`, `--exclude-weekends`, `--json`

> **Note:** Meeting finder algorithm is planned but not yet implemented.

---

### List Time Zones

```bash
nylas timezone list                    # List all zones (593 total)
nylas timezone list --filter America   # Filter by region
```

**Flags:** `--filter`, `--json`

---

### Get Zone Information

```bash
nylas timezone info America/New_York
nylas timezone info PST  # Abbreviations supported
```

**Flags:** `--zone`, `--time`, `--json`

**Common abbreviations:** PST, EST, CST, MST, GMT, IST, JST

---

## Calendar Integration

Calendar commands support timezone conversion and DST warnings:

```bash
nylas calendar events list --timezone America/Los_Angeles
nylas calendar events list --show-tz
```

### Working Hours & Breaks

Configure in `~/.nylas/config.yaml`:

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

**Enforcement:**
- Working hours: Soft warning (can override)
- Breaks: Hard block (protects your time)

---

## Tips

- Use abbreviations for faster typing (PST vs America/Los_Angeles)
- Add `--json` for scripting: `nylas timezone info UTC --json | jq`
- Works offline—no WiFi needed
- Check DST before scheduling recurring meetings

---

**Full guide:** [`docs/timezone/detailed.md`](timezone/detailed.md)
