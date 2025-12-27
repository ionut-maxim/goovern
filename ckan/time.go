package ckan

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Time is a custom time type that handles CKAN's timestamp format
type Time struct {
	time.Time
}

// UnmarshalJSON handles JSON unmarshaling for CKAN timestamps
func (t *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}

	// Remove quotes
	s := string(data)
	if len(s) > 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	// Try parsing with timezone first (RFC3339)
	parsed, err := time.Parse(time.RFC3339, s)
	if err == nil {
		t.Time = parsed
		return nil
	}

	// Try parsing without timezone (CKAN format)
	parsed, err = time.Parse("2006-01-02T15:04:05.999999", s)
	if err == nil {
		// Assume UTC if no timezone
		t.Time = parsed.UTC()
		return nil
	}

	// Try without microseconds
	parsed, err = time.Parse("2006-01-02T15:04:05", s)
	if err == nil {
		t.Time = parsed.UTC()
		return nil
	}

	return fmt.Errorf("cannot parse time: %s", s)
}

// MarshalJSON handles JSON marshaling
func (t Time) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return []byte(`"` + t.Format(time.RFC3339) + `"`), nil
}

// Value implements the driver.Valuer interface for database insertion
func (t Time) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan implements the sql.Scanner interface for database retrieval
func (t *Time) Scan(value interface{}) error {
	if value == nil {
		t.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		t.Time = v
		return nil
	default:
		return fmt.Errorf("cannot scan type %T into Time", value)
	}
}
