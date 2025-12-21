package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mqasimca/nylas/internal/domain"
	"github.com/mqasimca/nylas/internal/ports"
)

// EmailAnalyzer analyzes email threads to extract meeting context.
type EmailAnalyzer struct {
	nylasClient ports.NylasClient
	llmRouter   ports.LLMRouter
}

// NewEmailAnalyzer creates a new email analyzer.
func NewEmailAnalyzer(nylasClient ports.NylasClient, llmRouter ports.LLMRouter) *EmailAnalyzer {
	return &EmailAnalyzer{
		nylasClient: nylasClient,
		llmRouter:   llmRouter,
	}
}

// AnalyzeThread analyzes an email thread and extracts meeting context.
func (a *EmailAnalyzer) AnalyzeThread(ctx context.Context, grantID, threadID string, req *domain.EmailAnalysisRequest) (*domain.EmailThreadAnalysis, error) {
	// 1. Fetch the thread from Nylas API
	thread, err := a.nylasClient.GetThread(ctx, grantID, threadID)
	if err != nil {
		return nil, fmt.Errorf("fetch thread: %w", err)
	}

	// 2. Fetch all messages in the thread
	messages, err := a.fetchThreadMessages(ctx, grantID, threadID)
	if err != nil {
		return nil, fmt.Errorf("fetch thread messages: %w", err)
	}

	if len(messages) == 0 {
		return nil, fmt.Errorf("thread has no messages")
	}

	// 3. Build analysis context
	threadContext := a.buildThreadContext(thread, messages)

	// 4. Use LLM to analyze the thread
	analysis, err := a.analyzeWithLLM(ctx, threadContext, req)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis: %w", err)
	}

	// 5. Analyze participants
	participants := a.analyzeParticipants(thread, messages)
	analysis.Participants = participants

	// 6. Detect urgency indicators
	urgencyIndicators := a.detectUrgencyIndicators(messages)
	analysis.UrgencyIndicators = urgencyIndicators

	return analysis, nil
}

// fetchThreadMessages fetches all messages in a thread.
func (a *EmailAnalyzer) fetchThreadMessages(ctx context.Context, grantID, threadID string) ([]domain.Message, error) {
	params := &domain.MessageQueryParams{
		ThreadID: threadID,
		Limit:    100, // Fetch up to 100 messages in the thread
	}

	messages, err := a.nylasClient.GetMessagesWithParams(ctx, grantID, params)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

// buildThreadContext builds a string representation of the thread for LLM analysis.
func (a *EmailAnalyzer) buildThreadContext(thread *domain.Thread, messages []domain.Message) string {
	var builder strings.Builder

	builder.WriteString(fmt.Sprintf("Email Thread: %s\n", thread.Subject))
	builder.WriteString(fmt.Sprintf("Participants: %d\n", len(thread.Participants)))
	builder.WriteString(fmt.Sprintf("Messages: %d\n\n", len(messages)))

	// Add participants
	builder.WriteString("Participants:\n")
	for _, p := range thread.Participants {
		if p.Name != "" {
			builder.WriteString(fmt.Sprintf("- %s <%s>\n", p.Name, p.Email))
		} else {
			builder.WriteString(fmt.Sprintf("- %s\n", p.Email))
		}
	}
	builder.WriteString("\n")

	// Add message summaries (most recent first)
	builder.WriteString("Message Thread:\n")
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		sender := "Unknown"
		if len(msg.From) > 0 {
			if msg.From[0].Name != "" {
				sender = msg.From[0].Name
			} else {
				sender = msg.From[0].Email
			}
		}

		// Format timestamp
		timestamp := msg.Date.Format("Jan 2, 2006 3:04 PM")

		builder.WriteString(fmt.Sprintf("\n[%s] %s:\n", timestamp, sender))

		// Add message body (truncate if too long)
		body := msg.Body
		if len(body) > 500 {
			body = body[:500] + "..."
		}
		builder.WriteString(body)
		builder.WriteString("\n")
	}

	return builder.String()
}

// analyzeWithLLM uses the LLM to analyze the thread and generate insights.
func (a *EmailAnalyzer) analyzeWithLLM(ctx context.Context, threadContext string, req *domain.EmailAnalysisRequest) (*domain.EmailThreadAnalysis, error) {
	// Build the analysis prompt
	prompt := a.buildAnalysisPrompt(threadContext, req)

	// Create chat request
	chatReq := &domain.ChatRequest{
		Messages: []domain.ChatMessage{
			{
				Role:    "system",
				Content: "You are an expert meeting scheduler and email analyst. Analyze email threads to extract meeting context, topics, priority, and participant involvement.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.3, // Lower temperature for more factual analysis
		MaxTokens:   2000,
	}

	// Call LLM
	response, err := a.llmRouter.Chat(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("LLM chat: %w", err)
	}

	// Parse LLM response into analysis
	analysis, err := a.parseAnalysisResponse(response.Content, req)
	if err != nil {
		return nil, fmt.Errorf("parse LLM response: %w", err)
	}

	return analysis, nil
}

// buildAnalysisPrompt builds the prompt for LLM analysis.
func (a *EmailAnalyzer) buildAnalysisPrompt(threadContext string, req *domain.EmailAnalysisRequest) string {
	var builder strings.Builder

	builder.WriteString("Analyze the following email thread and provide:\n\n")
	builder.WriteString("1. The primary purpose of the discussion (1 sentence)\n")
	builder.WriteString("2. Key topics discussed (list 2-5 topics)\n")
	builder.WriteString("3. Priority level (low, medium, high, or urgent) with reasoning\n")
	builder.WriteString("4. Suggested meeting duration in minutes\n")

	if req.IncludeAgenda {
		builder.WriteString("5. A structured meeting agenda with items and estimated durations\n")
	}

	if req.IncludeTime {
		builder.WriteString("6. Best time for the meeting considering participant timezones\n")
	}

	builder.WriteString("\nFormat your response as follows:\n")
	builder.WriteString("PURPOSE: [purpose]\n")
	builder.WriteString("TOPICS:\n- [topic 1]\n- [topic 2]\n")
	builder.WriteString("PRIORITY: [level] - [reasoning]\n")
	builder.WriteString("DURATION: [minutes] minutes - [reasoning]\n")

	if req.IncludeAgenda {
		builder.WriteString("AGENDA:\n## [Agenda Title]\n### Item 1: [title] ([duration] min)\n[description]\n")
	}

	builder.WriteString("\n---\n\n")
	builder.WriteString(threadContext)

	return builder.String()
}

// parseAnalysisResponse parses the LLM response into an EmailThreadAnalysis.
func (a *EmailAnalyzer) parseAnalysisResponse(response string, req *domain.EmailAnalysisRequest) (*domain.EmailThreadAnalysis, error) {
	analysis := &domain.EmailThreadAnalysis{
		ThreadID: req.ThreadID,
	}

	lines := strings.Split(response, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Parse PURPOSE
		if strings.HasPrefix(line, "PURPOSE:") {
			analysis.Purpose = strings.TrimSpace(strings.TrimPrefix(line, "PURPOSE:"))
		}

		// Parse TOPICS
		if strings.HasPrefix(line, "TOPICS:") {
			// Next lines starting with "- " are topics
			for j := i + 1; j < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[j]), "- "); j++ {
				topic := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(lines[j]), "- "))
				analysis.Topics = append(analysis.Topics, topic)
			}
		}

		// Parse PRIORITY
		if strings.HasPrefix(line, "PRIORITY:") {
			priorityLine := strings.TrimPrefix(line, "PRIORITY:")
			parts := strings.SplitN(priorityLine, "-", 2)
			if len(parts) > 0 {
				priorityStr := strings.TrimSpace(strings.ToLower(parts[0]))
				analysis.Priority = domain.MeetingPriority(priorityStr)
			}
		}

		// Parse DURATION
		if strings.HasPrefix(line, "DURATION:") {
			durationLine := strings.TrimPrefix(line, "DURATION:")
			// Extract number from "60 minutes - reasoning"
			parts := strings.Fields(durationLine)
			if len(parts) > 0 {
				var duration int
				_, _ = fmt.Sscanf(parts[0], "%d", &duration)
				analysis.SuggestedDuration = duration
			}
		}

		// Parse AGENDA (if requested)
		if req.IncludeAgenda && strings.HasPrefix(line, "AGENDA:") {
			agenda := a.parseAgenda(lines[i+1:])
			analysis.Agenda = agenda
		}
	}

	// Default values if parsing failed
	if analysis.SuggestedDuration == 0 {
		analysis.SuggestedDuration = 30 // Default 30 minutes
	}
	if analysis.Priority == "" {
		analysis.Priority = domain.PriorityMedium
	}

	return analysis, nil
}

// parseAgenda parses the agenda section from LLM response.
func (a *EmailAnalyzer) parseAgenda(lines []string) *domain.MeetingAgenda {
	agenda := &domain.MeetingAgenda{
		Items: []domain.AgendaItem{},
	}

	var currentItem *domain.AgendaItem

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Agenda title (##)
		if strings.HasPrefix(line, "## ") {
			agenda.Title = strings.TrimPrefix(line, "## ")
		}

		// Agenda item (###)
		if strings.HasPrefix(line, "### ") {
			if currentItem != nil {
				agenda.Items = append(agenda.Items, *currentItem)
			}

			// Parse "Item 1: Title (30 min)"
			itemLine := strings.TrimPrefix(line, "### ")
			currentItem = &domain.AgendaItem{}

			// Extract duration from parentheses
			if idx := strings.Index(itemLine, "("); idx != -1 {
				if endIdx := strings.Index(itemLine, ")"); endIdx > idx {
					durationStr := itemLine[idx+1 : endIdx]
					var duration int
					_, _ = fmt.Sscanf(durationStr, "%d", &duration)
					currentItem.Duration = duration
					itemLine = strings.TrimSpace(itemLine[:idx])
				}
			}

			currentItem.Title = itemLine
		}

		// Description (regular text after item)
		if currentItem != nil && !strings.HasPrefix(line, "#") && line != "" {
			if currentItem.Description != "" {
				currentItem.Description += " "
			}
			currentItem.Description += line
		}
	}

	// Add the last item
	if currentItem != nil {
		agenda.Items = append(agenda.Items, *currentItem)
	}

	return agenda
}

// analyzeParticipants analyzes participant involvement in the thread.
func (a *EmailAnalyzer) analyzeParticipants(thread *domain.Thread, messages []domain.Message) []domain.ParticipantInfo {
	participantMap := make(map[string]*domain.ParticipantInfo)

	// Initialize participants from thread
	for _, p := range thread.Participants {
		participantMap[p.Email] = &domain.ParticipantInfo{
			Email:        p.Email,
			Name:         p.Name,
			MentionCount: 0,
			MessageCount: 0,
			Required:     false,
			Involvement:  domain.InvolvementLow,
		}
	}

	// Count messages per participant
	for _, msg := range messages {
		for _, from := range msg.From {
			if p, exists := participantMap[from.Email]; exists {
				p.MessageCount++
				p.LastMessageAt = msg.Date.Format(time.RFC3339)
			}
		}

		// Count mentions in message body
		body := strings.ToLower(msg.Body)
		for email, p := range participantMap {
			if strings.Contains(body, strings.ToLower(email)) {
				p.MentionCount++
			}
		}
	}

	// Determine involvement level and required status
	totalMessages := len(messages)
	participants := make([]domain.ParticipantInfo, 0, len(participantMap))

	for _, p := range participantMap {
		// High involvement: sent >30% of messages or mentioned >3 times
		if totalMessages > 0 && (float64(p.MessageCount)/float64(totalMessages) > 0.3 || p.MentionCount > 3) {
			p.Involvement = domain.InvolvementHigh
			p.Required = true
		} else if p.MessageCount > 1 || p.MentionCount > 0 {
			p.Involvement = domain.InvolvementMedium
			p.Required = true
		} else {
			p.Involvement = domain.InvolvementLow
			p.Required = false
		}

		participants = append(participants, *p)
	}

	return participants
}

// detectUrgencyIndicators detects urgency signals in the email thread.
func (a *EmailAnalyzer) detectUrgencyIndicators(messages []domain.Message) []string {
	indicators := []string{}

	urgentKeywords := []string{
		"urgent", "asap", "immediately", "critical", "emergency",
		"deadline", "today", "tomorrow", "this week",
	}

	// Check for urgent keywords
	for _, msg := range messages {
		bodyLower := strings.ToLower(msg.Body)
		subjectLower := strings.ToLower(msg.Subject)

		for _, keyword := range urgentKeywords {
			if strings.Contains(bodyLower, keyword) || strings.Contains(subjectLower, keyword) {
				indicators = append(indicators, fmt.Sprintf("Contains urgent keyword: '%s'", keyword))
			}
		}
	}

	// Check for high message frequency
	if len(messages) > 5 {
		// Calculate time span
		if len(messages) > 1 {
			earliest := messages[0].Date
			latest := messages[len(messages)-1].Date
			duration := latest.Sub(earliest)

			if duration < 24*time.Hour && len(messages) > 3 {
				indicators = append(indicators, fmt.Sprintf("%d messages in %s (high activity)", len(messages), duration.Round(time.Hour)))
			}
		}
	}

	// Check for multiple participants (broad reach)
	participantEmails := make(map[string]bool)
	for _, msg := range messages {
		for _, from := range msg.From {
			participantEmails[from.Email] = true
		}
	}

	if len(participantEmails) > 5 {
		indicators = append(indicators, fmt.Sprintf("%d participants (broad reach)", len(participantEmails)))
	}

	// Remove duplicates
	seen := make(map[string]bool)
	uniqueIndicators := []string{}
	for _, indicator := range indicators {
		if !seen[indicator] {
			seen[indicator] = true
			uniqueIndicators = append(uniqueIndicators, indicator)
		}
	}

	return uniqueIndicators
}
