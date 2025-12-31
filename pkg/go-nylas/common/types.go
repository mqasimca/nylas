// Package common provides shared types and utilities for go-nylas.
package common

import "time"

// ListOptions contains common pagination options.
type ListOptions struct {
	Limit  int
	Offset int
}

// Participant represents an email participant (sender, recipient, etc).
type Participant struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email"`
}

// Attachment represents an email attachment.
type Attachment struct {
	ID          string `json:"id,omitempty"`
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
	ContentID   string `json:"content_id,omitempty"`
	Data        []byte `json:"data,omitempty"`
}

// Error represents a Nylas API error.
type Error struct {
	StatusCode int
	Message    string
	RequestID  string
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.RequestID != "" {
		return e.Message + " (request_id: " + e.RequestID + ")"
	}
	return e.Message
}

// DateTime represents a date-time value.
type DateTime struct {
	Time     time.Time
	Timezone string
}

// Date represents a date without time.
type Date struct {
	Year  int
	Month int
	Day   int
}

// Timespan represents a time range.
type Timespan struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}
