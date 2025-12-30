package air

import (
	"net/url"
	"testing"
)

func TestQueryParams_GetInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		key        string
		defaultVal int
		minVal     int
		maxVal     int
		want       int
	}{
		{"valid value", "limit=25", "limit", 50, 1, 100, 25},
		{"missing key", "", "limit", 50, 1, 100, 50},
		{"empty value", "limit=", "limit", 50, 1, 100, 50},
		{"below min", "limit=0", "limit", 50, 1, 100, 50},
		{"above max", "limit=150", "limit", 50, 1, 100, 50},
		{"invalid string", "limit=abc", "limit", 50, 1, 100, 50},
		{"negative value", "limit=-5", "limit", 50, 1, 100, 50},
		{"at min boundary", "limit=1", "limit", 50, 1, 100, 1},
		{"at max boundary", "limit=100", "limit", 50, 1, 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			q := NewQueryParams(values)
			got := q.GetInt(tt.key, tt.defaultVal, tt.minVal, tt.maxVal)
			if got != tt.want {
				t.Errorf("GetInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestQueryParams_GetLimit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		defaultVal int
		want       int
	}{
		{"valid limit", "limit=25", 50, 25},
		{"missing limit", "", 50, 50},
		{"max limit", "limit=200", 50, 200},
		{"over max limit", "limit=500", 50, 50},
		{"zero limit", "limit=0", 50, 50},
		{"different default", "limit=abc", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			q := NewQueryParams(values)
			got := q.GetLimit(tt.defaultVal)
			if got != tt.want {
				t.Errorf("GetLimit() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestQueryParams_GetInt64(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		key        string
		defaultVal int64
		want       int64
	}{
		{"valid timestamp", "start=1704067200", "start", 0, 1704067200},
		{"missing key", "", "start", 0, 0},
		{"empty value", "start=", "start", 100, 100},
		{"invalid string", "start=abc", "start", 0, 0},
		{"negative value", "start=-1000", "start", 0, -1000},
		{"large value", "start=9999999999999", "start", 0, 9999999999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			q := NewQueryParams(values)
			got := q.GetInt64(tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("GetInt64() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestQueryParams_GetBool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		key   string
		want  bool
	}{
		{"true value", "unread=true", "unread", true},
		{"false value", "unread=false", "unread", false},
		{"missing key", "", "unread", false},
		{"empty value", "unread=", "unread", false},
		{"uppercase TRUE", "unread=TRUE", "unread", false},
		{"numeric 1", "unread=1", "unread", false},
		{"yes value", "unread=yes", "unread", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			q := NewQueryParams(values)
			got := q.GetBool(tt.key)
			if got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryParams_GetBoolPtr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		query   string
		key     string
		wantNil bool
		wantVal bool
	}{
		{"true value", "unread=true", "unread", false, true},
		{"false value", "unread=false", "unread", false, false},
		{"missing key", "", "unread", true, false},
		{"empty value", "unread=", "unread", true, false}, // Empty = not meaningfully set
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			q := NewQueryParams(values)
			got := q.GetBoolPtr(tt.key)
			if tt.wantNil {
				if got != nil {
					t.Errorf("GetBoolPtr() = %v, want nil", *got)
				}
			} else {
				if got == nil {
					t.Errorf("GetBoolPtr() = nil, want %v", tt.wantVal)
				} else if *got != tt.wantVal {
					t.Errorf("GetBoolPtr() = %v, want %v", *got, tt.wantVal)
				}
			}
		})
	}
}

func TestQueryParams_GetString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		query      string
		key        string
		defaultVal string
		want       string
	}{
		{"present value", "folder=inbox", "folder", "all", "inbox"},
		{"missing key", "", "folder", "all", "all"},
		{"empty value", "folder=", "folder", "all", "all"},
		{"special chars", "q=hello+world", "q", "", "hello world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			q := NewQueryParams(values)
			got := q.GetString(tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("GetString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestQueryParams_Has(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		key   string
		want  bool
	}{
		{"present with value", "limit=50", "limit", true},
		{"present empty", "limit=", "limit", true},
		{"missing", "other=value", "limit", false},
		{"empty query", "", "limit", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			values, _ := url.ParseQuery(tt.query)
			q := NewQueryParams(values)
			got := q.Has(tt.key)
			if got != tt.want {
				t.Errorf("Has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryParams_Get(t *testing.T) {
	t.Parallel()

	values, _ := url.ParseQuery("foo=bar&empty=")
	q := NewQueryParams(values)

	if got := q.Get("foo"); got != "bar" {
		t.Errorf("Get(foo) = %q, want %q", got, "bar")
	}
	if got := q.Get("empty"); got != "" {
		t.Errorf("Get(empty) = %q, want empty", got)
	}
	if got := q.Get("missing"); got != "" {
		t.Errorf("Get(missing) = %q, want empty", got)
	}
}

// Test standalone helper functions (deprecated but still need coverage)
func TestParseLimit(t *testing.T) {
	t.Parallel()

	values, _ := url.ParseQuery("limit=75")
	got := ParseLimit(values, 50)
	if got != 75 {
		t.Errorf("ParseLimit() = %d, want 75", got)
	}
}

func TestParseInt(t *testing.T) {
	t.Parallel()

	values, _ := url.ParseQuery("page=5")
	got := ParseInt(values, "page", 1, 1, 100)
	if got != 5 {
		t.Errorf("ParseInt() = %d, want 5", got)
	}
}

func TestParseInt64(t *testing.T) {
	t.Parallel()

	values, _ := url.ParseQuery("timestamp=1704067200")
	got := ParseInt64(values, "timestamp", 0)
	if got != 1704067200 {
		t.Errorf("ParseInt64() = %d, want 1704067200", got)
	}
}

func TestParseBool(t *testing.T) {
	t.Parallel()

	values, _ := url.ParseQuery("active=true")
	got := ParseBool(values, "active")
	if !got {
		t.Errorf("ParseBool() = %v, want true", got)
	}
}
