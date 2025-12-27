package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// CachedEvent represents a calendar event stored in the cache.
type CachedEvent struct {
	ID           string    `json:"id"`
	CalendarID   string    `json:"calendar_id,omitempty"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty"`
	Location     string    `json:"location,omitempty"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	AllDay       bool      `json:"all_day"`
	Recurring    bool      `json:"recurring"`
	RRule        string    `json:"rrule,omitempty"`
	Participants []string  `json:"participants,omitempty"`
	Status       string    `json:"status,omitempty"`
	Busy         bool      `json:"busy"`
	CachedAt     time.Time `json:"cached_at"`
}

// EventStore provides event caching operations.
type EventStore struct {
	db *sql.DB
}

// NewEventStore creates an event store for a database.
func NewEventStore(db *sql.DB) *EventStore {
	return &EventStore{db: db}
}

// Put stores an event in the cache.
func (s *EventStore) Put(event *CachedEvent) error {
	participantsJSON, _ := json.Marshal(event.Participants)

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO events (
			id, calendar_id, title, description, location,
			start_time, end_time, all_day, recurring, rrule,
			participants_json, status, busy, cached_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		event.ID, event.CalendarID, event.Title, event.Description, event.Location,
		event.StartTime.Unix(), event.EndTime.Unix(), boolToInt(event.AllDay),
		boolToInt(event.Recurring), event.RRule, string(participantsJSON),
		event.Status, boolToInt(event.Busy), time.Now().Unix(),
	)
	return err
}

// PutBatch stores multiple events in a transaction.
func (s *EventStore) PutBatch(events []*CachedEvent) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO events (
			id, calendar_id, title, description, location,
			start_time, end_time, all_day, recurring, rrule,
			participants_json, status, busy, cached_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	now := time.Now().Unix()
	for _, event := range events {
		participantsJSON, _ := json.Marshal(event.Participants)

		_, err = stmt.Exec(
			event.ID, event.CalendarID, event.Title, event.Description, event.Location,
			event.StartTime.Unix(), event.EndTime.Unix(), boolToInt(event.AllDay),
			boolToInt(event.Recurring), event.RRule, string(participantsJSON),
			event.Status, boolToInt(event.Busy), now,
		)
		if err != nil {
			return fmt.Errorf("insert event %s: %w", event.ID, err)
		}
	}

	return tx.Commit()
}

// Get retrieves an event by ID.
func (s *EventStore) Get(id string) (*CachedEvent, error) {
	row := s.db.QueryRow(`
		SELECT id, calendar_id, title, description, location,
			start_time, end_time, all_day, recurring, rrule,
			participants_json, status, busy, cached_at
		FROM events WHERE id = ?
	`, id)

	return scanEvent(row)
}

// EventListOptions configures event listing.
type EventListOptions struct {
	CalendarID string
	Start      time.Time
	End        time.Time
	Limit      int
	Offset     int
}

// List retrieves events with filtering.
func (s *EventStore) List(opts EventListOptions) ([]*CachedEvent, error) {
	query := `
		SELECT id, calendar_id, title, description, location,
			start_time, end_time, all_day, recurring, rrule,
			participants_json, status, busy, cached_at
		FROM events
		WHERE 1=1
	`
	var args []any

	if opts.CalendarID != "" {
		query += " AND calendar_id = ?"
		args = append(args, opts.CalendarID)
	}
	if !opts.Start.IsZero() {
		query += " AND end_time >= ?"
		args = append(args, opts.Start.Unix())
	}
	if !opts.End.IsZero() {
		query += " AND start_time <= ?"
		args = append(args, opts.End.Unix())
	}

	query += " ORDER BY start_time ASC"

	if opts.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", opts.Limit)
	}
	if opts.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", opts.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []*CachedEvent
	for rows.Next() {
		event, err := scanEventRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

// ListByDateRange retrieves events within a date range.
func (s *EventStore) ListByDateRange(start, end time.Time) ([]*CachedEvent, error) {
	return s.List(EventListOptions{Start: start, End: end})
}

// Search performs full-text search on events.
func (s *EventStore) Search(query string, limit int) ([]*CachedEvent, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(`
		SELECT e.id, e.calendar_id, e.title, e.description, e.location,
			e.start_time, e.end_time, e.all_day, e.recurring, e.rrule,
			e.participants_json, e.status, e.busy, e.cached_at
		FROM events e
		WHERE e.rowid IN (
			SELECT rowid FROM events_fts WHERE events_fts MATCH ?
		)
		ORDER BY e.start_time ASC
		LIMIT ?
	`, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search events: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var events []*CachedEvent
	for rows.Next() {
		event, err := scanEventRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

// Delete removes an event from the cache.
func (s *EventStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM events WHERE id = ?", id)
	return err
}

// DeleteByCalendar removes all events for a calendar.
func (s *EventStore) DeleteByCalendar(calendarID string) error {
	_, err := s.db.Exec("DELETE FROM events WHERE calendar_id = ?", calendarID)
	return err
}

// Count returns the number of cached events.
func (s *EventStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM events").Scan(&count)
	return count, err
}

// GetUpcoming retrieves upcoming events.
func (s *EventStore) GetUpcoming(limit int) ([]*CachedEvent, error) {
	now := time.Now()
	return s.List(EventListOptions{
		Start: now,
		Limit: limit,
	})
}

func scanEvent(row *sql.Row) (*CachedEvent, error) {
	var event CachedEvent
	var participantsJSON sql.NullString
	var startUnix, endUnix, cachedAtUnix int64
	var allDay, recurring, busy int

	err := row.Scan(
		&event.ID, &event.CalendarID, &event.Title, &event.Description, &event.Location,
		&startUnix, &endUnix, &allDay, &recurring, &event.RRule,
		&participantsJSON, &event.Status, &busy, &cachedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	event.StartTime = time.Unix(startUnix, 0)
	event.EndTime = time.Unix(endUnix, 0)
	event.CachedAt = time.Unix(cachedAtUnix, 0)
	event.AllDay = allDay == 1
	event.Recurring = recurring == 1
	event.Busy = busy == 1

	if participantsJSON.Valid {
		_ = json.Unmarshal([]byte(participantsJSON.String), &event.Participants)
	}

	return &event, nil
}

func scanEventRow(rows *sql.Rows) (*CachedEvent, error) {
	var event CachedEvent
	var participantsJSON sql.NullString
	var startUnix, endUnix, cachedAtUnix int64
	var allDay, recurring, busy int

	err := rows.Scan(
		&event.ID, &event.CalendarID, &event.Title, &event.Description, &event.Location,
		&startUnix, &endUnix, &allDay, &recurring, &event.RRule,
		&participantsJSON, &event.Status, &busy, &cachedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	event.StartTime = time.Unix(startUnix, 0)
	event.EndTime = time.Unix(endUnix, 0)
	event.CachedAt = time.Unix(cachedAtUnix, 0)
	event.AllDay = allDay == 1
	event.Recurring = recurring == 1
	event.Busy = busy == 1

	if participantsJSON.Valid {
		_ = json.Unmarshal([]byte(participantsJSON.String), &event.Participants)
	}

	return &event, nil
}
