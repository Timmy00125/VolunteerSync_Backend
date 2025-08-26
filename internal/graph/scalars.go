package graph

import (
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

// MarshalTime converts time.Time to a GraphQL string in RFC3339 format.
func MarshalTime(t time.Time) graphql.Marshaler {
	// Ensure UTC for consistency
	tt := t.UTC()
	return graphql.MarshalString(tt.Format(time.RFC3339Nano))
}

// UnmarshalTime parses a GraphQL input into time.Time.
func UnmarshalTime(v interface{}) (time.Time, error) {
	switch val := v.(type) {
	case string:
		// Try RFC3339 variants
		if ts, err := time.Parse(time.RFC3339Nano, val); err == nil {
			return ts, nil
		}
		if ts, err := time.Parse(time.RFC3339, val); err == nil {
			return ts, nil
		}
		return time.Time{}, fmt.Errorf("invalid time format: %q", val)
	default:
		return time.Time{}, fmt.Errorf("time must be a string, got %T", v)
	}
}
