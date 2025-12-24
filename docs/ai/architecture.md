## Architecture

### Multi-Provider LLM Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Nylas AI Assistant CLI                   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  User Input: "Schedule 30-min call with John next Tuesday" │
│       ↓                                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Unified LLM Interface (Provider-Agnostic)          │   │
│  │  - Input validation                                 │   │
│  │  - Prompt templating                                │   │
│  │  - Function calling setup                           │   │
│  │  - Response parsing                                 │   │
│  └─────────────────────────────────────────────────────┘   │
│       ↓                                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Provider Router (selects based on config)          │   │
│  └─────────────────────────────────────────────────────┘   │
│       ↓         ↓           ↓          ↓                   │
│  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐              │
│  │ Ollama │ │ Claude │ │OpenAI  │ │ Groq   │              │
│  │ (Local)│ │ (Cloud)│ │(Cloud) │ │(Cloud) │              │
│  │ FREE   │ │  $$    │ │  $$    │ │  $     │              │
│  └────────┘ └────────┘ └────────┘ └────────┘              │
│       ↓                                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Function Calling Tools                             │   │
│  │  - findMeetingTime(participants, duration, ...)     │   │
│  │  - checkDST(time, timezone)                         │   │
│  │  - validateWorkingHours(time, timezone)             │   │
│  │  - createEvent(title, time, participants, ...)      │   │
│  │  - analyzePatterns(userId, lookbackDays)            │   │
│  │  - suggestFocusTime(userId, weekStart)              │   │
│  └─────────────────────────────────────────────────────┘   │
│       ↓                                                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  Calendar & Timezone Services                       │   │
│  │  - Nylas API (events, availability, calendars)      │   │
│  │  - Timezone Service (conversions, DST, scoring)     │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Function Calling Tools

AI agents have access to specialized tools:

| Tool | Description | Parameters |
|------|-------------|------------|
| `findMeetingTime` | Find optimal meeting times across timezones | participants, duration, dateRange, workingHours |
| `checkDST` | Check if time falls during DST transition | time, timezone |
| `validateWorkingHours` | Check if time is within working hours | time, timezone, workStart, workEnd |
| `createEvent` | Create calendar event | title, startTime, endTime, participants, timezone |
| `getAvailability` | Get participant availability | participants, startTime, endTime |
| `analyzePatterns` | Analyze historical calendar patterns | userId, lookbackDays |
| `suggestFocusTime` | Suggest focus time blocks | userId, weekStart |

---

