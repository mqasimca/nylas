package air

import (
	"net/url"
	"strconv"
)

// QueryParams wraps url.Values with convenience methods for common parsing patterns.
type QueryParams struct {
	values url.Values
}

// NewQueryParams creates a QueryParams wrapper from url.Values.
func NewQueryParams(values url.Values) *QueryParams {
	return &QueryParams{values: values}
}

// Get returns the first value for the given key, or empty string if not present.
func (q *QueryParams) Get(key string) string {
	return q.values.Get(key)
}

// GetInt parses an integer query parameter with bounds checking.
// Returns defaultVal if the parameter is missing, empty, invalid, or out of bounds.
func (q *QueryParams) GetInt(key string, defaultVal, minVal, maxVal int) int {
	s := q.values.Get(key)
	if s == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(s)
	if err != nil || parsed < minVal || parsed > maxVal {
		return defaultVal
	}
	return parsed
}

// GetLimit parses a "limit" parameter with standard bounds (1-200, default 50).
func (q *QueryParams) GetLimit(defaultVal int) int {
	return q.GetInt("limit", defaultVal, 1, 200)
}

// GetInt64 parses an int64 query parameter (e.g., Unix timestamps).
// Returns defaultVal if the parameter is missing, empty, or invalid.
func (q *QueryParams) GetInt64(key string, defaultVal int64) int64 {
	s := q.values.Get(key)
	if s == "" {
		return defaultVal
	}
	parsed, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return defaultVal
	}
	return parsed
}

// GetBool parses a boolean query parameter.
// Returns true only if the value is exactly "true".
func (q *QueryParams) GetBool(key string) bool {
	return q.values.Get(key) == "true"
}

// GetBoolPtr parses a boolean query parameter and returns a pointer.
// Returns nil if the parameter is not present, otherwise returns pointer to the bool value.
func (q *QueryParams) GetBoolPtr(key string) *bool {
	s := q.values.Get(key)
	if s == "" {
		return nil
	}
	val := s == "true"
	return &val
}

// GetString returns the parameter value, or defaultVal if empty.
func (q *QueryParams) GetString(key, defaultVal string) string {
	s := q.values.Get(key)
	if s == "" {
		return defaultVal
	}
	return s
}

// Has returns true if the parameter is present (even if empty).
func (q *QueryParams) Has(key string) bool {
	_, ok := q.values[key]
	return ok
}

// ParseLimit is a standalone helper for parsing limit with standard bounds.
// Deprecated: Use QueryParams.GetLimit() instead for new code.
func ParseLimit(query url.Values, defaultVal int) int {
	return NewQueryParams(query).GetLimit(defaultVal)
}

// ParseInt is a standalone helper for parsing an integer with bounds.
// Deprecated: Use QueryParams.GetInt() instead for new code.
func ParseInt(query url.Values, key string, defaultVal, minVal, maxVal int) int {
	return NewQueryParams(query).GetInt(key, defaultVal, minVal, maxVal)
}

// ParseInt64 is a standalone helper for parsing int64 values.
// Deprecated: Use QueryParams.GetInt64() instead for new code.
func ParseInt64(query url.Values, key string, defaultVal int64) int64 {
	return NewQueryParams(query).GetInt64(key, defaultVal)
}

// ParseBool is a standalone helper for parsing boolean values.
// Deprecated: Use QueryParams.GetBool() instead for new code.
func ParseBool(query url.Values, key string) bool {
	return NewQueryParams(query).GetBool(key)
}
