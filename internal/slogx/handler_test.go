package slogx_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/i-pul/search-flight/internal/slogx"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slogx.NewContextHandler(slog.NewJSONHandler(buf, nil)))
}

func parseRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var m map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &m))
	return m
}

// TestContextHandler_WithTraceID verifies that a trace ID stored in the context
// is injected as "trace_id" in every log record.
func TestContextHandler_WithTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	ctx := context.WithValue(context.Background(), slogx.TraceIDKey, "abc-123")
	logger.InfoContext(ctx, "hello")

	rec := parseRecord(t, &buf)
	assert.Equal(t, "abc-123", rec["trace_id"])
}

// TestContextHandler_WithoutTraceID verifies that no "trace_id" field is emitted
// when the context carries no trace ID.
func TestContextHandler_WithoutTraceID(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	logger.InfoContext(context.Background(), "hello")

	rec := parseRecord(t, &buf)
	assert.Nil(t, rec["trace_id"])
}

// TestContextHandler_DerivedContext verifies that a trace ID is still visible
// after the context is wrapped (as errgroup.WithContext or context.WithCancel do).
func TestContextHandler_DerivedContext(t *testing.T) {
	var buf bytes.Buffer
	logger := newTestLogger(&buf)

	base := context.WithValue(context.Background(), slogx.TraceIDKey, "derived-456")
	child, cancel := context.WithCancel(base)
	defer cancel()

	logger.InfoContext(child, "inside goroutine")

	rec := parseRecord(t, &buf)
	assert.Equal(t, "derived-456", rec["trace_id"], "trace_id must propagate through derived contexts")
}
