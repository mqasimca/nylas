package cache

import (
	"database/sql"
	"fmt"
	"time"
)

// CachedCalendar represents a calendar stored in the cache.
type CachedCalendar struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	IsPrimary   bool      `json:"is_primary"`
	ReadOnly    bool      `json:"read_only"`
	HexColor    string    `json:"hex_color,omitempty"`
	CachedAt    time.Time `json:"cached_at"`
}

// CalendarStore provides calendar caching operations.
type CalendarStore struct {
	db *sql.DB
}

// NewCalendarStore creates a calendar store for a database.
func NewCalendarStore(db *sql.DB) *CalendarStore {
	return &CalendarStore{db: db}
}

// Put stores a calendar in the cache.
func (s *CalendarStore) Put(calendar *CachedCalendar) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO calendars (
			id, name, description, is_primary, read_only, hex_color, cached_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		calendar.ID, calendar.Name, calendar.Description,
		boolToInt(calendar.IsPrimary), boolToInt(calendar.ReadOnly),
		calendar.HexColor, time.Now().Unix(),
	)
	return err
}

// PutBatch stores multiple calendars in a transaction.
func (s *CalendarStore) PutBatch(calendars []*CachedCalendar) error {
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
		INSERT OR REPLACE INTO calendars (
			id, name, description, is_primary, read_only, hex_color, cached_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare statement: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	now := time.Now().Unix()
	for _, calendar := range calendars {
		_, err = stmt.Exec(
			calendar.ID, calendar.Name, calendar.Description,
			boolToInt(calendar.IsPrimary), boolToInt(calendar.ReadOnly),
			calendar.HexColor, now,
		)
		if err != nil {
			return fmt.Errorf("insert calendar %s: %w", calendar.ID, err)
		}
	}

	return tx.Commit()
}

// Get retrieves a calendar by ID.
func (s *CalendarStore) Get(id string) (*CachedCalendar, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, is_primary, read_only, hex_color, cached_at
		FROM calendars WHERE id = ?
	`, id)

	return scanCalendar(row)
}

// GetPrimary retrieves the primary calendar.
func (s *CalendarStore) GetPrimary() (*CachedCalendar, error) {
	row := s.db.QueryRow(`
		SELECT id, name, description, is_primary, read_only, hex_color, cached_at
		FROM calendars WHERE is_primary = 1
	`)

	return scanCalendar(row)
}

// List retrieves all calendars.
func (s *CalendarStore) List() ([]*CachedCalendar, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, is_primary, read_only, hex_color, cached_at
		FROM calendars
		ORDER BY is_primary DESC, name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query calendars: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var calendars []*CachedCalendar
	for rows.Next() {
		calendar, err := scanCalendarRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan calendar: %w", err)
		}
		calendars = append(calendars, calendar)
	}

	return calendars, rows.Err()
}

// ListWritable retrieves calendars that can be written to.
func (s *CalendarStore) ListWritable() ([]*CachedCalendar, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, is_primary, read_only, hex_color, cached_at
		FROM calendars
		WHERE read_only = 0
		ORDER BY is_primary DESC, name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("query calendars: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var calendars []*CachedCalendar
	for rows.Next() {
		calendar, err := scanCalendarRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scan calendar: %w", err)
		}
		calendars = append(calendars, calendar)
	}

	return calendars, rows.Err()
}

// Delete removes a calendar from the cache.
func (s *CalendarStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM calendars WHERE id = ?", id)
	return err
}

// Count returns the number of cached calendars.
func (s *CalendarStore) Count() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM calendars").Scan(&count)
	return count, err
}

func scanCalendar(row *sql.Row) (*CachedCalendar, error) {
	var calendar CachedCalendar
	var isPrimary, readOnly int
	var cachedAtUnix int64

	err := row.Scan(
		&calendar.ID, &calendar.Name, &calendar.Description,
		&isPrimary, &readOnly, &calendar.HexColor, &cachedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	calendar.IsPrimary = isPrimary == 1
	calendar.ReadOnly = readOnly == 1
	calendar.CachedAt = time.Unix(cachedAtUnix, 0)
	return &calendar, nil
}

func scanCalendarRow(rows *sql.Rows) (*CachedCalendar, error) {
	var calendar CachedCalendar
	var isPrimary, readOnly int
	var cachedAtUnix int64

	err := rows.Scan(
		&calendar.ID, &calendar.Name, &calendar.Description,
		&isPrimary, &readOnly, &calendar.HexColor, &cachedAtUnix,
	)
	if err != nil {
		return nil, err
	}

	calendar.IsPrimary = isPrimary == 1
	calendar.ReadOnly = readOnly == 1
	calendar.CachedAt = time.Unix(cachedAtUnix, 0)
	return &calendar, nil
}
