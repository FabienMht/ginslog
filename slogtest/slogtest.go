package slogtest

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// Mock slog handler for testing.
type MockHandler struct {
	slog.Handler

	// Testing instance.
	testing *testing.T
	// Level to check with the record level.
	level slog.Level
	// Fields to check with the record fields.
	fields []slog.Attr
	// Ignore value check for these fields.
	skipFields []string
}

// NewMockHandler creates a new mock handler.
func NewMockHandler(h slog.Handler, t *testing.T, l slog.Level, f []slog.Attr, sf []string) *MockHandler {
	return &MockHandler{Handler: h, testing: t, level: l, fields: f, skipFields: sf}
}

// Handle implements Handler.Handle.
func (h *MockHandler) Handle(ctx context.Context, r slog.Record) error {
	// Check if the level matches.
	require.Equal(h.testing, h.level, r.Level)

	// Get fields names from fields to check.
	fieldsMap := make(map[string]any)
	for _, f := range h.fields {
		fieldsMap[f.Key] = f.Value
	}

	// Check each field in the record.
	r.Attrs(func(a slog.Attr) bool {
		// Check if the field exists in the expected fields.
		value, ok := fieldsMap[a.Key]
		require.True(h.testing, ok, fmt.Sprintf("field '%s' not found", a.Key))

		// Ignore value check.
		if slices.Contains(h.skipFields, a.Key) {
			return true
		}

		// Check if the field value matches with the expected value.
		require.Equal(
			h.testing, value, a.Value,
			fmt.Sprintf("field '%s' value '%s' does not match with '%s'", a.Key, a.Value, value),
		)
		return true
	})

	return h.Handler.Handle(ctx, r)
}
