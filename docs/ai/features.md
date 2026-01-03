## Features

### Natural Language Scheduling

**Status:** ✅ Available

Schedule meetings using natural language instead of manual date/time selection.

#### Example Usage (Local AI - Privacy Mode)

```bash
$ nylas calendar ai schedule "30-minute meeting with john@example.com next Tuesday afternoon"

🤖 AI Scheduling Assistant (Privacy Mode: Ollama Mistral 7B)

Processing locally... ✓

Analyzing request:
  ✓ Detected: 30-minute meeting
  ✓ Participant: john@example.com (timezone: America/New_York, auto-detected)
  ✓ Timeframe: Next Tuesday afternoon (Jan 21, 2025)
  ✓ Your timezone: America/Los_Angeles

Calling tools:
  → getAvailability(participants=[john@example.com], startTime=2025-01-21T12:00:00-08:00, ...)
  → findMeetingTime(participants=[you, john@example.com], duration=30, ...)

Top 3 AI-Suggested Times:

1. 🟢 Tuesday, Jan 21, 2:00 PM PST (Score: 94/100)
   You: 2:00 PM - 2:30 PM PST (Mid-afternoon, good energy)
   John: 5:00 PM - 5:30 PM EST (End of day, still acceptable)

   Why this is good:
   • Both in working hours
   • No conflicts detected
   • Your calendar shows high productivity at 2 PM historically

2. 🟡 Tuesday, Jan 21, 1:00 PM PST (Score: 82/100)
   You: 1:00 PM - 1:30 PM PST (Post-lunch, moderate energy)
   John: 4:00 PM - 4:30 PM EST (Late afternoon)

Create meeting with option #1? [y/N/2/3]: y

Creating event...
✓ Event created: event-789
  Title: Meeting with John
  When: Tuesday, Jan 21, 2025, 2:00 PM - 2:30 PM PST
  Participants: john@example.com

🔒 Privacy: All processing done locally, no data sent to cloud.
```

#### Example Usage (Cloud AI - Advanced)

```bash
$ nylas calendar ai schedule \
    --provider claude \
    "Find time for 1-hour planning session with team next week, preferably morning"

🤖 AI Scheduling Assistant (Claude Sonnet 4.5)

Analyzing request using Claude...

Detected context:
  • Meeting type: Planning session (suggests need for focus)
  • Duration: 1 hour
  • Participants: Checking team members from your calendar...
  • Timeframe: Next week (Jan 20-24, 2025)
  • Preference: Morning

Found 5 team members across 3 timezones:
  • alice@team.com (PST)
  • bob@team.com (EST)
  • carol@team.com (GMT)
  • david@team.com (IST)
  • eve@team.com (PST)

AI Insight: Planning sessions typically need high focus. Recommending morning
slots to ensure high energy for all participants.

Top Recommendations:

1. 🟢 Wed, Jan 22, 9:00 AM PST (Score: 91/100)
   Timezone overlap: Excellent

   PST (Alice, Eve): 9:00 AM - 10:00 AM ✓ Morning (peak energy)
   EST (Bob): 12:00 PM - 1:00 PM ⚠️ Lunch time
   GMT (Carol): 5:00 PM - 6:00 PM ✓ End of day
   IST (David): 10:30 PM - 11:30 PM ✗ Late night

   AI Reasoning: 60% of team in ideal hours. Consider rotating time
   monthly to be fair to David (IST timezone).

2. 🔄 Rotating Schedule Suggestion:
   Week 1-2: 9:00 AM PST (good for US/Europe)
   Week 3-4: 8:00 PM PST (good for Asia/Europe)

   Claude's Analysis: This ensures fairness. David (IST) will have good
   times 50% of the time, rather than never.

Apply rotating schedule? [y/N/view-more]

💰 Cost: ~$0.02 (Claude API tokens used)
```

---

### Smart Meeting Finder

**Status:** ✅ Partially Complete (Task 2.1 in progress)

Find optimal meeting times across multiple timezones using 100-point scoring algorithm.

#### Scoring Algorithm

| Factor | Points | Details |
|--------|--------|---------|
| Working Hours Coverage | 40 pts | All participants in 9 AM - 5 PM local time |
| Time Quality | 25 pts | Mid-morning (10 AM) better than early/late |
| Cultural Considerations | 15 pts | Avoid Friday PM (Middle East), lunch hours |
| Weekday vs Weekend | 10 pts | Weekdays preferred |
| Holiday Avoidance | 10 pts | No major holidays in any timezone |

#### Example Usage

```bash
$ nylas calendar find-time \
    --participants alice@example.com,bob@example.com \
    --duration 1h \
    --dates "next week"

🌍 Multi-Timezone Meeting Finder

Participants:
  • You: America/Los_Angeles (PST)
  • Alice: America/New_York (EST)
  • Bob: Europe/London (GMT)

Top 3 Suggested Times:

1. 🟢 Tuesday, Jan 7, 10:00 AM PST (Score: 94/100)
   You:   10:00 AM - 11:00 AM PST (Mid-morning ✨)
   Alice:  1:00 PM -  2:00 PM EST (Early afternoon)
   Bob:    6:00 PM -  7:00 PM GMT (End of day ⚠️)

   Score Breakdown:
   • Working Hours: 40/40 (all participants)
   • Time Quality: 22/25 (good for 2/3)
   • Cultural: 15/15 (no conflicts)
   • Weekday: 10/10
   • Holidays: 7/10 (Bob's New Year observed)

💡 Recommendation: Book option #1 for best overall experience
```

---

### Predictive Scheduling

**Status:** ✅ Available

Learn from your calendar history and predict optimal scheduling patterns.

#### Example Usage

```bash
$ nylas calendar ai analyze --learn-patterns

🧠 AI Calendar Pattern Analysis

Analyzing your last 90 days of meetings...

Discovered Patterns:

📊 Meeting Acceptance Patterns:
  • Monday 9-11 AM: 92% accept rate (you like morning meetings)
  • Friday 3-5 PM: 34% accept rate (you avoid late Friday meetings)
  • Tuesday/Wednesday 2-4 PM: 88% accept rate (your sweet spot)

⏱️ Meeting Duration Patterns:
  • 1-on-1s with John: Average 32 min (scheduled 30 min + 2 min overrun)
  • Team standups: Average 18 min (scheduled 30 min, usually shorter)
  • Client calls: Average 48 min (scheduled 60 min, often end early)

🌍 Timezone Patterns:
  • You schedule 67% of meetings with EST timezone participants
  • Your preferred time for cross-timezone: 2-4 PM PST
  • You avoid early morning (<9 AM) for Asia timezone meetings

🎯 Productivity Insights:
  • Peak focus: Tuesday/Thursday 10 AM - 12 PM (2-hour blocks)
  • Low energy: Monday mornings, Friday afternoons
  • Best for creative work: Tuesday 2-4 PM (fewest interruptions)

💡 AI Recommendations:
  1. Block Tuesday/Thursday 10 AM - 12 PM for focus time
  2. Schedule 1-on-1s on Wednesday afternoons (your preference)
  3. Suggest 45-minute default for client calls (not 60 min)
  4. Decline Friday PM meetings automatically (low productivity)

Apply these AI-optimized rules to your calendar? [y/N/customize]
```

---

### Conflict Resolution

**Status:** ✅ Available

Intelligently detect and resolve scheduling conflicts.

#### Example Usage

```bash
$ nylas calendar ai reschedule evt_123 --reason "Conflict with urgent task"

🤖 AI Rescheduling Assistant

Current meeting:
  • "Product Review" with 5 participants
  • Scheduled: Wed, Jan 22, 2:00 PM PST
  • Duration: 1 hour

Reason: Conflict with urgent task

Finding alternative times...
  ✓ Analyzed participant calendars
  ✓ Checked historical preferences
  ✓ Applied timezone optimization
  ✓ Considered meeting priority (Medium)

Top Reschedule Options:

1. 🟢 Same day, later (Score: 94/100)
   Thu, Jan 23, 3:30 PM PST

   Pros:
   • Same week (minimal disruption)
   • All participants available
   • Better time for 3/5 participants
   • No other meetings nearby (buffer time)

   Cons:
   • 1 day delay

2. 🟡 Next week, preferred slot (Score: 89/100)
   Tue, Jan 28, 2:00 PM PST

   Pros:
   • Everyone's preferred time (Tuesday afternoon)
   • More preparation time

   Cons:
   • 6 day delay (might impact decision timeline)

💡 AI Insight: Option #1 is best. Minimal delay + better participant fit.

Auto-send reschedule request with option #1? [y/N]
```

---

### Focus Time Protection

**Status:** ✅ Available

AI-adaptive deep work protection that learns and adjusts.

#### Example Usage

```bash
$ nylas calendar ai focus-time --enable

🧠 AI Focus Time Protection

Analyzing your productivity patterns...

Discovered Focus Patterns:
  • Peak productivity: Tuesday/Thursday 10 AM - 12 PM
  • Deep work sessions: Average 2.5 hours
  • Most productive day: Wednesday (fewest interruptions)
  • Least productive: Friday PM (high meeting density)

AI-Recommended Focus Time Blocks:

📅 Weekly Schedule:
  Monday:    ░░░░░░░░░░░░████████░░░░░░░░ 2-4 PM (2 hrs)
  Tuesday:   ████████████░░░░░░░░░░░░░░░░ 10 AM-12 PM (2 hrs) ⭐ Peak
  Wednesday: ████████████████████░░░░░░░░ 9 AM-1 PM (4 hrs) ⭐ Deep Work Day
  Thursday:  ████████████░░░░░░░░░░░░░░░░ 10 AM-12 PM (2 hrs) ⭐ Peak
  Friday:    ░░░░░░░░░░░░░░░░░░░░████████ 3-5 PM (2 hrs) ⚠️ Low energy

Total: 14 hours/week protected for focus time

🛡️ Protection Rules:
  1. Auto-decline meeting requests during focus blocks
  2. Suggest alternative times when requests come in
  3. Allow override for "urgent" meetings (you approve)
  4. Dynamically adjust if deadline pressure increases

Apply AI focus time protection? [y/N/customize]
```

#### Break Time Awareness

**Status:** ✅ Available

The AI focus time protection respects configured break blocks and working hours:

**Break Integration:**
- **Hard Breaks:** AI cannot schedule focus time during lunch, coffee breaks, or other configured break blocks
- **Soft Working Hours:** AI prefers working hours but can suggest outside hours if needed
- **Break Patterns:** AI learns from your break preferences and avoids scheduling near natural break times

**Example with Configured Breaks:**

If your config has lunch breaks:
```yaml
working_hours:
  default:
    breaks:
      - name: "Lunch"
        start: "12:00"
        end: "13:00"
        type: "lunch"
```

AI focus time suggestions will **automatically exclude** lunch hours:

```bash
$ nylas calendar ai focus-time --enable

AI-Recommended Focus Time Blocks (respecting lunch 12-1 PM):

Monday:    ████████░░░░████████░░░░░░░░ 9-11 AM, 1-3 PM (4 hrs)
           Lunch Break: 12:00-13:00 (protected)

Tuesday:   ████████████░░░░░░░░████████ 9 AM-12 PM, 1-3 PM (5 hrs)
           Lunch Break: 12:00-13:00 (protected)
```

**Break-Aware Scheduling:**
1. **Respects all break blocks** - Lunch, coffee, custom breaks cannot be overridden
2. **Works around breaks** - Schedules focus blocks before/after breaks naturally
3. **Learns break patterns** - Identifies when you typically take breaks even without explicit config
4. **Suggests break-friendly blocks** - Prefers longer morning/afternoon sessions with lunch in between

**Configuration:**
For working hours and break configuration, see [Timezone & Working Hours Guide](../commands/timezone.md#working-hours--break-management).

---

### Recurring Pattern Learning

**Status:** ✅ Available (Phase 4)

AI-powered analysis of your historical calendar data to learn scheduling patterns and provide personalized recommendations.

#### What It Analyzes

The pattern learner examines your past meetings to discover:

- **Acceptance Patterns**: Which days/times you most frequently accept meetings
- **Duration Patterns**: Actual vs scheduled meeting lengths by type
- **Timezone Preferences**: Cross-timezone meeting patterns
- **Productivity Insights**: Peak focus times based on meeting density
- **Participant Patterns**: Per-person scheduling preferences

#### Example Usage

**Basic Analysis (Last 30 Days):**

```bash
$ nylas calendar ai analyze --days 30

🔍 Analyzing 30 days of meeting history...

📊 Analysis Period: 2025-11-21 to 2025-12-21
📅 Total Meetings Analyzed: 122

✅ Meeting Acceptance Patterns
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Overall Acceptance Rate: 100.0%

By Day of Week:
     Monday: 100.0% ████████████████████
    Tuesday: 100.0% ████████████████████
  Wednesday: 100.0% ████████████████████
   Thursday: 100.0% ████████████████████
     Friday: 100.0% ████████████████████

By Time of Day (working hours):
  09:00: 100.0% ████████████████████
  10:00: 100.0% ████████████████████
  11:00: 100.0% ████████████████████
  12:00: 100.0% ████████████████████

⏱️  Meeting Duration Patterns
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Average Scheduled: 56 minutes
Average Actual: 56 minutes
Overrun Rate: 0.0%

🌍 Timezone Distribution
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  America/Toronto: 72 meetings (59%)
  UTC: 18 meetings (15%)
  America/New_York: 14 meetings (11%)
  America/Los_Angeles: 10 meetings (8%)

🎯 Productivity Insights
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Peak Focus Times (recommended for deep work):
  1. Monday 09:00-11:00 (score: 80/100)
  2. Monday 10:00-12:00 (score: 90/100)
  3. Monday 11:00-13:00 (score: 90/100)

Meeting Density by Day:
     Monday: 1.0 meetings/day  ⭐ Best for focus
    Tuesday: 1.8 meetings/day
  Wednesday: 2.8 meetings/day  ⚠️ Busiest
   Thursday: 2.5 meetings/day
     Friday: 2.0 meetings/day

💡 AI Recommendations
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. 🔴 Block Monday 09:00-11:00 for focus time
   Historical data shows you have few meetings during this time,
   making it ideal for deep work.

   📌 Action: Create recurring focus time block
   📈 Impact: Increase productivity by 20-30%
   🎯 Confidence: 80%

📝 Key Insights
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. You accept 100% of meetings on Wednesdays (your best day)
2. Peak focus time: Monday 09:00-11:00 (fewest meetings)
3. Most meetings in America/Toronto timezone (72 meetings)
4. Analyzed 122 meetings over 30 days
```

**Custom Analysis Period:**

```bash
# Analyze last 60 days
$ nylas calendar ai analyze --days 60

# Analyze last 90 days (quarterly patterns)
$ nylas calendar ai analyze --days 90
```

#### How It Works

1. **Data Collection**: Fetches historical events from all your calendars
2. **Pattern Detection**: Uses statistical analysis to identify trends
3. **AI Enhancement**: Optional LLM analysis for deeper insights and recommendations
4. **Privacy**: All analysis happens locally; LLM calls only for recommendations (opt-in)

#### Technical Details

**API Limits:**
- Nylas API v3 has a maximum limit of **200 events per request per calendar**
- Pattern learner automatically iterates through all calendars
- For accounts with many events, analysis fetches up to 200 events per calendar

**Event Scope:**
- Analyzes events from **all accessible calendars** (primary + shared)
- Includes confirmed, tentative, and declined meetings
- Excludes recurring event rules (analyzes individual instances)

**Performance:**
- Typical analysis (30 days): 2-5 seconds
- Large accounts (500+ events): 10-15 seconds
- All processing happens locally for privacy

#### Use Cases

**📊 Quarterly Review:**
```bash
# Analyze 90 days to understand quarterly patterns
$ nylas calendar ai analyze --days 90
```

**🎯 Optimize Schedule:**
- Identify when you're most available for focus work
- Discover your preferred meeting times
- Understand meeting overrun patterns

**🌍 Multi-Timezone Teams:**
- Identify which timezones you collaborate with most
- Find optimal times that work across time zones
- Detect timezone-related scheduling patterns

**⏰ Meeting Efficiency:**
- Compare scheduled vs actual meeting durations
- Identify which meeting types run long
- Optimize future meeting time allocations

#### Privacy & Data

- ✅ **Historical data only**: Analyzes past meetings (no future predictions exposed)
- ✅ **Local processing**: Pattern detection happens on your machine
- ✅ **Opt-in LLM**: AI recommendations require explicit consent
- ✅ **No storage**: Patterns computed on-demand, not stored
- ✅ **GDPR compliant**: All data processing follows privacy regulations

#### Integration with Other Features

Pattern learning enhances other AI features:

- **Focus Time Protection**: Uses learned patterns to suggest optimal focus blocks
- **Smart Meeting Finder**: Considers your acceptance patterns when suggesting times
- **Conflict Resolution**: Understands which meetings are more flexible based on history

---

### Meeting Context Analysis

**Status:** 🔄 Planned (Phase 3, Task 3.5)

Extract meeting context from email threads and auto-generate agendas.

#### Example Usage

```bash
$ nylas calendar ai analyze-thread --email thread_456

🤖 AI Email Thread Analysis

Analyzing email thread...
  Thread: "Q1 Planning Discussion" (12 messages)
  Participants: 5 people
  Duration: 3 days
  Latest: 2 hours ago

AI-Detected Context:

📋 Meeting Purpose:
  Primary topic: Q1 budget planning and resource allocation
  Secondary topics:
    • Hiring priorities for engineering team
    • Marketing campaign timeline
    • Product roadmap alignment

⏱️ Suggested Duration:
  Recommended: 90 minutes

  Reasoning:
  • 3 major topics detected → needs 30 min each
  • 5 participants → more discussion time needed
  • Complex topic (budget) → requires detail

  Compare to: Similar meetings averaged 82 minutes

🎯 Detected Priority: High
  Urgency indicators:
  • Deadline mentioned: "need to finalize by Jan 25"
  • CEO copied on thread
  • 3 follow-ups in 24 hours → high interest

👥 Key Participants:
  Required:
  • alice@team.com (Finance lead - mentioned in 8 messages)
  • bob@team.com (Engineering manager - decision maker)

  Optional:
  • carol@team.com (mentioned once, FYI)

📝 Auto-Generated Agenda:

Meeting Agenda: Q1 Planning Discussion
Duration: 90 minutes
Priority: High

1. Q1 Budget Review (30 min)
   - Review proposed budget allocation
   - Discuss cost savings opportunities
   - Decision: Approve final budget

2. Engineering Hiring Plan (30 min)
   - Frontend vs Backend hiring priorities
   - Timeline for new hires
   - Decision: Approve job postings

3. Marketing & Product Alignment (20 min)
   - Campaign timeline for Q1
   - Product roadmap milestones
   - Decision: Confirm launch dates

4. Action Items & Next Steps (10 min)
   - Assign owners
   - Set follow-up dates

💡 AI Actions:
  [1] Create meeting with suggested time & agenda
  [2] Send calendar invite to required attendees
  [3] Mark Carol as optional (minimal thread involvement)
  [4] Set 90-minute duration (AI-recommended)

Create AI-optimized meeting? [y/N/customize]
```

---

