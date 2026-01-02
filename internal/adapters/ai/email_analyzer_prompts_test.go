//go:build !integration

package ai

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEmailAnalyzer_BuildAnalysisPrompt(t *testing.T) {
	analyzer := &EmailAnalyzer{}

	tests := []struct {
		name           string
		threadContext  string
		req            *domain.EmailAnalysisRequest
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:          "basic prompt without options",
			threadContext: "Test thread content",
			req: &domain.EmailAnalysisRequest{
				IncludeAgenda: false,
				IncludeTime:   false,
			},
			wantContains: []string{
				"primary purpose",
				"Key topics discussed",
				"Priority level",
				"Suggested meeting duration",
				"PURPOSE:",
				"TOPICS:",
				"PRIORITY:",
				"DURATION:",
				"Test thread content",
			},
			wantNotContain: []string{
				"structured meeting agenda",
				"Best time for the meeting",
				"AGENDA:",
			},
		},
		{
			name:          "prompt with agenda",
			threadContext: "Thread with agenda request",
			req: &domain.EmailAnalysisRequest{
				IncludeAgenda: true,
				IncludeTime:   false,
			},
			wantContains: []string{
				"structured meeting agenda",
				"AGENDA:",
				"## [Agenda Title]",
			},
		},
		{
			name:          "prompt with time suggestion",
			threadContext: "Thread with time request",
			req: &domain.EmailAnalysisRequest{
				IncludeAgenda: false,
				IncludeTime:   true,
			},
			wantContains: []string{
				"Best time for the meeting",
				"participant timezones",
			},
		},
		{
			name:          "prompt with all options",
			threadContext: "Full featured thread",
			req: &domain.EmailAnalysisRequest{
				IncludeAgenda: true,
				IncludeTime:   true,
			},
			wantContains: []string{
				"structured meeting agenda",
				"Best time for the meeting",
				"AGENDA:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.buildAnalysisPrompt(tt.threadContext, tt.req)

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}

			for _, notWant := range tt.wantNotContain {
				assert.NotContains(t, result, notWant)
			}
		})
	}
}

func TestEmailAnalyzer_ParseAnalysisResponse(t *testing.T) {
	analyzer := &EmailAnalyzer{}

	tests := []struct {
		name            string
		response        string
		req             *domain.EmailAnalysisRequest
		wantPurpose     string
		wantTopics      []string
		wantPriority    domain.MeetingPriority
		wantDuration    int
		wantAgendaTitle string
		wantAgendaItems int
	}{
		{
			name: "parses complete response",
			response: `PURPOSE: Schedule a meeting to discuss the product roadmap
TOPICS:
- Feature prioritization
- Q1 milestones
- Resource allocation
PRIORITY: high - Critical for Q1 planning
DURATION: 60 minutes - Complex discussion required`,
			req:          &domain.EmailAnalysisRequest{IncludeAgenda: false},
			wantPurpose:  "Schedule a meeting to discuss the product roadmap",
			wantTopics:   []string{"Feature prioritization", "Q1 milestones", "Resource allocation"},
			wantPriority: domain.PriorityHigh,
			wantDuration: 60,
		},
		{
			name: "parses response with agenda",
			response: `PURPOSE: Team sync meeting
TOPICS:
- Status updates
PRIORITY: medium - Regular sync
DURATION: 30 minutes - Quick sync
AGENDA:
## Team Sync Agenda
### Item 1: Status Updates (15 min)
Each team member shares progress
### Item 2: Blockers (10 min)
Discuss any blockers
### Item 3: Action Items (5 min)
Assign next steps`,
			req:             &domain.EmailAnalysisRequest{IncludeAgenda: true},
			wantPurpose:     "Team sync meeting",
			wantPriority:    domain.PriorityMedium,
			wantDuration:    30,
			wantAgendaTitle: "Team Sync Agenda",
			wantAgendaItems: 3,
		},
		{
			name: "handles lowercase priority",
			response: `PURPOSE: Quick chat
TOPICS:
- Catch up
PRIORITY: low - just a chat
DURATION: 15 minutes - brief`,
			req:          &domain.EmailAnalysisRequest{},
			wantPurpose:  "Quick chat",
			wantPriority: domain.PriorityLow,
			wantDuration: 15,
		},
		{
			name: "handles urgent priority",
			response: `PURPOSE: Emergency response
TOPICS:
- Incident handling
PRIORITY: urgent - system down
DURATION: 45 minutes - incident response`,
			req:          &domain.EmailAnalysisRequest{},
			wantPriority: domain.MeetingPriority("urgent"),
			wantDuration: 45,
		},
		{
			name:         "defaults for missing values",
			response:     `Just some random text without proper formatting`,
			req:          &domain.EmailAnalysisRequest{},
			wantPriority: domain.PriorityMedium, // Default
			wantDuration: 30,                    // Default
		},
		{
			name: "handles malformed duration",
			response: `PURPOSE: Test
TOPICS:
- Test
PRIORITY: medium - test
DURATION: invalid duration`,
			req:          &domain.EmailAnalysisRequest{},
			wantPriority: domain.PriorityMedium, // Priority is parsed correctly
			wantDuration: 30,                    // Default when duration parsing fails
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.parseAnalysisResponse(tt.response, tt.req)

			require.NoError(t, err)

			if tt.wantPurpose != "" {
				assert.Equal(t, tt.wantPurpose, result.Purpose)
			}

			if len(tt.wantTopics) > 0 {
				assert.Equal(t, tt.wantTopics, result.Topics)
			}

			assert.Equal(t, tt.wantPriority, result.Priority)
			assert.Equal(t, tt.wantDuration, result.SuggestedDuration)

			if tt.wantAgendaTitle != "" && result.Agenda != nil {
				assert.Equal(t, tt.wantAgendaTitle, result.Agenda.Title)
			}

			if tt.wantAgendaItems > 0 && result.Agenda != nil {
				assert.Len(t, result.Agenda.Items, tt.wantAgendaItems)
			}
		})
	}
}

func TestEmailAnalyzer_ParseAgenda(t *testing.T) {
	analyzer := &EmailAnalyzer{}

	tests := []struct {
		name      string
		lines     []string
		wantTitle string
		wantItems int
		wantFirst string
		wantDur   int
	}{
		{
			name: "parses full agenda",
			lines: []string{
				"## Product Review Agenda",
				"### Item 1: Introduction (5 min)",
				"Welcome and overview",
				"### Item 2: Demo (30 min)",
				"Product demonstration",
			},
			wantTitle: "Product Review Agenda",
			wantItems: 2,
			wantFirst: "Item 1: Introduction",
			wantDur:   5,
		},
		{
			name: "handles agenda without duration",
			lines: []string{
				"## Simple Agenda",
				"### Discussion Point",
				"Details about the discussion",
			},
			wantTitle: "Simple Agenda",
			wantItems: 1,
			wantFirst: "Discussion Point",
			wantDur:   0,
		},
		{
			name: "handles item with description",
			lines: []string{
				"## Meeting Agenda",
				"### Review (10 min)",
				"Review the proposal",
				"Consider alternatives",
			},
			wantTitle: "Meeting Agenda",
			wantItems: 1,
		},
		{
			name:      "handles empty lines",
			lines:     []string{},
			wantTitle: "",
			wantItems: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.parseAgenda(tt.lines)

			assert.Equal(t, tt.wantTitle, result.Title)
			assert.Len(t, result.Items, tt.wantItems)

			if tt.wantItems > 0 && tt.wantFirst != "" {
				assert.Equal(t, tt.wantFirst, result.Items[0].Title)
			}

			if tt.wantItems > 0 && tt.wantDur > 0 {
				assert.Equal(t, tt.wantDur, result.Items[0].Duration)
			}
		})
	}
}

// Note: AnalyzeInbox tests requiring LLM router mock are in integration tests

func TestEmailAnalyzer_BuildInboxPrompt(t *testing.T) {
	analyzer := &EmailAnalyzer{}

	tests := []struct {
		name         string
		messages     []domain.Message
		wantContains []string
	}{
		{
			name: "builds prompt with all message details",
			messages: []domain.Message{
				{
					From:    []domain.EmailParticipant{{Name: "Alice", Email: "alice@example.com"}},
					Subject: "Important Update",
					Date:    time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
					Snippet: "This is the message preview",
					Unread:  true,
					Starred: true,
				},
			},
			wantContains: []string{
				"Analyze these 1 emails",
				"From: Alice <alice@example.com>",
				"Subject: Important Update",
				"Preview: This is the message preview",
				"Status: UNREAD",
				"Status: STARRED",
			},
		},
		{
			name: "handles multiple messages",
			messages: []domain.Message{
				{From: []domain.EmailParticipant{{Email: "a@example.com"}}, Subject: "A", Date: time.Now()},
				{From: []domain.EmailParticipant{{Email: "b@example.com"}}, Subject: "B", Date: time.Now()},
				{From: []domain.EmailParticipant{{Email: "c@example.com"}}, Subject: "C", Date: time.Now()},
			},
			wantContains: []string{
				"Analyze these 3 emails",
				"--- Email 1 ---",
				"--- Email 2 ---",
				"--- Email 3 ---",
			},
		},
		{
			name: "handles sender with no name",
			messages: []domain.Message{
				{From: []domain.EmailParticipant{{Email: "noname@example.com"}}, Subject: "Test", Date: time.Now()},
			},
			wantContains: []string{
				"From: noname@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.buildInboxPrompt(tt.messages)

			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}
		})
	}
}

func TestEmailAnalyzer_ParseInboxResponse(t *testing.T) {
	analyzer := &EmailAnalyzer{}

	tests := []struct {
		name            string
		content         string
		wantErr         bool
		wantSummary     string
		wantCategories  int
		wantActionItems int
		wantHighlights  int
	}{
		{
			name: "parses valid JSON",
			content: `{
				"summary": "You have 5 unread emails",
				"categories": [
					{"name": "Work", "count": 3, "subjects": ["Meeting", "Report", "Update"]},
					{"name": "Personal", "count": 2, "subjects": ["Lunch", "Trip"]}
				],
				"action_items": [
					{"subject": "Meeting", "from": "boss@work.com", "urgency": "high", "reason": "Tomorrow"}
				],
				"highlights": ["Important meeting scheduled", "Report due"]
			}`,
			wantSummary:     "You have 5 unread emails",
			wantCategories:  2,
			wantActionItems: 1,
			wantHighlights:  2,
		},
		{
			name: "parses JSON wrapped in markdown",
			content: "Here's your analysis:\n```json\n" + `{
				"summary": "Test summary",
				"categories": [],
				"action_items": [],
				"highlights": []
			}` + "\n```",
			wantSummary: "Test summary",
		},
		{
			name:    "returns error for no JSON",
			content: "This is just plain text with no JSON at all",
			wantErr: true,
		},
		{
			name:    "returns error for invalid JSON",
			content: `{invalid json content}`,
			wantErr: true,
		},
		{
			name:    "returns error for malformed JSON",
			content: `{"summary": "test", "categories": [}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.parseInboxResponse(tt.content)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSummary, result.Summary)
			assert.Len(t, result.Categories, tt.wantCategories)
			assert.Len(t, result.ActionItems, tt.wantActionItems)
			assert.Len(t, result.Highlights, tt.wantHighlights)
		})
	}
}

func TestFormatInboxParticipants(t *testing.T) {
	tests := []struct {
		name         string
		participants []domain.EmailParticipant
		want         string
	}{
		{
			name:         "returns Unknown for empty list",
			participants: []domain.EmailParticipant{},
			want:         "Unknown",
		},
		{
			name: "formats participant with name",
			participants: []domain.EmailParticipant{
				{Name: "Alice Smith", Email: "alice@example.com"},
			},
			want: "Alice Smith <alice@example.com>",
		},
		{
			name: "formats participant without name",
			participants: []domain.EmailParticipant{
				{Email: "noname@example.com"},
			},
			want: "noname@example.com",
		},
		{
			name: "formats multiple participants",
			participants: []domain.EmailParticipant{
				{Name: "Alice", Email: "alice@example.com"},
				{Email: "bob@example.com"},
			},
			want: "Alice <alice@example.com>, bob@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatInboxParticipants(tt.participants)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		maxLen int
		want   string
	}{
		{
			name:   "returns string as-is if under limit",
			input:  "short",
			maxLen: 10,
			want:   "short",
		},
		{
			name:   "truncates with ellipsis",
			input:  "this is a long string that needs truncation",
			maxLen: 20,
			want:   "this is a long st...",
		},
		{
			name:   "handles exact length",
			input:  "exact",
			maxLen: 5,
			want:   "exact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateStr(tt.input, tt.maxLen)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestInboxSummaryTypes(t *testing.T) {
	t.Run("EmailCategory serializes correctly", func(t *testing.T) {
		cat := EmailCategory{
			Name:     "Work",
			Count:    5,
			Subjects: []string{"Meeting", "Report"},
		}

		data, err := json.Marshal(cat)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"name":"Work"`)
		assert.Contains(t, string(data), `"count":5`)
	})

	t.Run("ActionItem serializes correctly", func(t *testing.T) {
		item := ActionItem{
			Subject: "Urgent Meeting",
			From:    "boss@work.com",
			Urgency: "high",
			Reason:  "Response needed today",
		}

		data, err := json.Marshal(item)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"urgency":"high"`)
	})

	t.Run("InboxSummaryResponse contains all fields", func(t *testing.T) {
		resp := InboxSummaryResponse{
			Summary:      "Test summary",
			Categories:   []EmailCategory{{Name: "Test"}},
			ActionItems:  []ActionItem{{Subject: "Action"}},
			Highlights:   []string{"Highlight 1"},
			ProviderUsed: "claude",
			TokensUsed:   100,
		}

		data, err := json.Marshal(resp)
		require.NoError(t, err)
		assert.Contains(t, string(data), `"provider_used":"claude"`)
		assert.Contains(t, string(data), `"tokens_used":100`)
	})
}
