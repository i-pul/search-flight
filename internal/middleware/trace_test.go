package middleware_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/i-pul/search-flight/internal/middleware"
	"github.com/i-pul/search-flight/internal/slogx"
)

func TestTrace_SetsTraceIDHeader(t *testing.T) {
	var called bool
	handler := middleware.Trace(func(_ *fasthttp.RequestCtx) {
		called = true
	})

	var ctx fasthttp.RequestCtx
	handler(&ctx)

	assert.True(t, called, "next handler must be called")

	traceID := string(ctx.Response.Header.Peek("X-Trace-Id"))
	assert.NotEmpty(t, traceID)
	assert.Regexp(t,
		`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
		traceID,
		"X-Trace-Id must be a UUID v4",
	)
}

func TestTrace_InjectsTraceIDIntoContext(t *testing.T) {
	var capturedTraceID string
	handler := middleware.Trace(func(ctx *fasthttp.RequestCtx) {
		reqCtx := middleware.RequestContext(ctx)
		capturedTraceID, _ = reqCtx.Value(slogx.TraceIDKey).(string)
	})

	var ctx fasthttp.RequestCtx
	handler(&ctx)

	headerTraceID := string(ctx.Response.Header.Peek("X-Trace-Id"))
	require.NotEmpty(t, headerTraceID)
	assert.Equal(t, headerTraceID, capturedTraceID,
		"trace ID in context must match the response header")
}

func TestTrace_EachRequestGetsDifferentTraceID(t *testing.T) {
	ids := make([]string, 5)
	handler := middleware.Trace(func(ctx *fasthttp.RequestCtx) {})

	for i := range ids {
		var ctx fasthttp.RequestCtx
		handler(&ctx)
		ids[i] = string(ctx.Response.Header.Peek("X-Trace-Id"))
		require.NotEmpty(t, ids[i])
	}

	for i := 1; i < len(ids); i++ {
		assert.NotEqual(t, ids[0], ids[i], "every request should get a unique trace ID")
	}
}

func TestRequestContext_ReturnsFasthttpCtxWhenNotSet(t *testing.T) {
	var ctx fasthttp.RequestCtx
	// No Trace middleware — UserValue is not set.
	result := middleware.RequestContext(&ctx)
	assert.NotNil(t, result)
	// Should fall back to the fasthttp ctx itself (not panic).
}
