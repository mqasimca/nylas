## FAQ

### Q: Is my data private with Ollama?

**A:** Yes! Ollama runs 100% locally on your machine. No data is ever sent to external servers. It's GDPR and HIPAA compliant.

### Q: Which AI provider is best?

**A:**
- **Privacy-first:** Ollama (local, free)
- **Best reasoning:** Claude (cloud, $$)
- **Most versatile:** OpenAI (cloud, $$$)
- **Fastest:** Groq (cloud, $)

Start with Ollama, upgrade to cloud if needed.

### Q: How much does cloud AI cost?

**A:**
- **Claude:** ~$0.02 per scheduling request
- **OpenAI:** ~$0.05 per request
- **Groq:** ~$0.005 per request

Ollama is completely free.

### Q: Can I use multiple providers?

**A:** Yes! Configure fallback chains:

```yaml
ai:
  fallback:
    enabled: true
    providers: [ollama, claude, openai]
```

AI will try Ollama first, fall back to Claude if needed.

### Q: Does AI work offline?

**A:** Yes, if you use Ollama (local provider). Cloud providers (Claude, OpenAI, Groq) require internet.

### Q: How accurate is AI scheduling?

**A:** AI suggestions are validated through:
- Timezone conversion checks
- DST transition validation
- Working hours verification
- Conflict detection

All suggestions include confidence scores.

### Q: How does pattern learning work?

**A:** Pattern learning analyzes your **historical** calendar events to identify:
- Meeting acceptance patterns (which days/times you prefer)
- Duration patterns (how long meetings actually run)
- Timezone preferences (which zones you collaborate with most)
- Productivity patterns (best times for focus work)

**Requirements:**
- Minimum 10-15 historical events for basic patterns
- 30+ days recommended for reliable insights
- Works with completed meetings only (past events)

**Usage:**
```bash
# Analyze last 30 days
nylas calendar ai analyze --days 30

# Quarterly review
nylas calendar ai analyze --days 90
```

**Privacy:** All analysis happens locally. LLM calls only for recommendations (opt-in).

### Q: Why am I getting "no events found" for pattern learning?

**A:** Pattern learning analyzes **past events only**. If you see "no events found":

1. **Check date range**: The default is last 30 days. Your calendar may only have future events.
2. **Try longer period**: `nylas calendar ai analyze --days 60`
3. **Verify calendar access**: Ensure the CLI can access your calendars

**Note:** Pattern learning needs historical data. It cannot analyze future events.

### Q: Does pattern learning analyze all my events?

**A:** Pattern learning fetches up to **200 events per calendar** (Nylas API v3 limit). For most users, this covers 1-3 months of history.

**If you have many calendars:**
- Analysis includes events from all accessible calendars
- Each calendar contributes up to 200 events
- Total analyzed = 200 × number of calendars

**Example:** If you have 3 calendars, pattern learning can analyze up to 600 total events.

**Tip:** For very active calendars (500+ events/month), the analysis focuses on the most recent 200 events per calendar within the specified date range.

### Q: Can I delete my AI data?

**A:** Yes! Run:

```bash
nylas ai clear-data
```

This deletes all learned patterns and preferences.

### Q: Is AI required for timezone features?

**A:** No! All timezone features work without AI:

```bash
# Timezone conversion (no AI)
nylas timezone convert --from PST --to IST

# Calendar with timezone display (no AI)
nylas calendar events list --timezone America/Los_Angeles --show-tz
```

AI adds natural language and predictive features.

---

