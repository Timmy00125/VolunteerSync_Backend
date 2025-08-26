package generated

import (
	"context"
	"fmt"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

// _Time marshals a time.Time to RFC3339 string.
func (ec *executionContext) _Time(ctx context.Context, sel ast.SelectionSet, v *time.Time) graphql.Marshaler {
	if v == nil {
		return graphql.Null
	}
	tt := v.UTC()
	return graphql.MarshalString(tt.Format(time.RFC3339Nano))
}

// unmarshalInputTime parses a GraphQL input into time.Time.
func (ec *executionContext) unmarshalInputTime(ctx context.Context, obj any) (time.Time, error) {
	switch val := obj.(type) {
	case string:
		if ts, err := time.Parse(time.RFC3339Nano, val); err == nil {
			return ts, nil
		}
		if ts, err := time.Parse(time.RFC3339, val); err == nil {
			return ts, nil
		}
		return time.Time{}, fmt.Errorf("invalid time format: %q", val)
	default:
		return time.Time{}, fmt.Errorf("time must be a string, got %T", obj)
	}
}
