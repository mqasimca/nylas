package air

import (
	"time"

	"github.com/mqasimca/nylas/internal/air/cache"
	"github.com/mqasimca/nylas/internal/domain"
)

// domainMessageToCached converts a domain message to a cached email.
func domainMessageToCached(msg *domain.Message) *cache.CachedEmail {
	var fromName, fromEmail string
	if len(msg.From) > 0 {
		fromName = msg.From[0].Name
		fromEmail = msg.From[0].Email
	}

	var folderID string
	if len(msg.Folders) > 0 {
		folderID = msg.Folders[0]
	}

	return &cache.CachedEmail{
		ID:             msg.ID,
		ThreadID:       msg.ThreadID,
		FolderID:       folderID,
		Subject:        msg.Subject,
		Snippet:        msg.Snippet,
		FromName:       fromName,
		FromEmail:      fromEmail,
		To:             participantsToStrings(msg.To),
		CC:             participantsToStrings(msg.Cc),
		BCC:            participantsToStrings(msg.Bcc),
		Date:           msg.Date,
		Unread:         msg.Unread,
		Starred:        msg.Starred,
		HasAttachments: len(msg.Attachments) > 0,
		BodyHTML:       msg.Body,
		BodyText:       msg.Body, // Simplified
		CachedAt:       time.Now(),
	}
}

// domainEventToCached converts a domain event to a cached event.
func domainEventToCached(evt *domain.Event, calendarID string) *cache.CachedEvent {
	return &cache.CachedEvent{
		ID:           evt.ID,
		CalendarID:   calendarID,
		Title:        evt.Title,
		Description:  evt.Description,
		Location:     evt.Location,
		StartTime:    time.Unix(evt.When.StartTime, 0),
		EndTime:      time.Unix(evt.When.EndTime, 0),
		AllDay:       evt.When.Object == "date" || evt.When.Object == "datespan",
		Status:       evt.Status,
		Busy:         evt.Busy,
		Participants: eventParticipantsToStrings(evt.Participants),
		CachedAt:     time.Now(),
	}
}

// domainContactToCached converts a domain contact to a cached contact.
func domainContactToCached(c *domain.Contact) *cache.CachedContact {
	var email, phone, company, jobTitle string
	if len(c.Emails) > 0 {
		email = c.Emails[0].Email
	}
	if len(c.PhoneNumbers) > 0 {
		phone = c.PhoneNumbers[0].Number
	}
	if len(c.CompanyName) > 0 {
		company = c.CompanyName
	}
	if len(c.JobTitle) > 0 {
		jobTitle = c.JobTitle
	}

	return &cache.CachedContact{
		ID:          c.ID,
		Email:       email,
		GivenName:   c.GivenName,
		Surname:     c.Surname,
		DisplayName: c.GivenName + " " + c.Surname,
		Phone:       phone,
		Company:     company,
		JobTitle:    jobTitle,
		Notes:       c.Notes,
		CachedAt:    time.Now(),
	}
}

// participantsToStrings converts email participants to a slice of strings.
func participantsToStrings(participants []domain.EmailParticipant) []string {
	result := make([]string, 0, len(participants))
	for _, p := range participants {
		if p.Name != "" {
			result = append(result, p.Name+" <"+p.Email+">")
		} else {
			result = append(result, p.Email)
		}
	}
	return result
}

// eventParticipantsToStrings converts event participants to a slice of strings.
func eventParticipantsToStrings(participants []domain.Participant) []string {
	result := make([]string, 0, len(participants))
	for _, p := range participants {
		if p.Name != "" {
			result = append(result, p.Name+" <"+p.Email+">")
		} else {
			result = append(result, p.Email)
		}
	}
	return result
}
