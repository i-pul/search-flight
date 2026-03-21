package middleware

import (
	"context"

	"github.com/google/uuid"
	"github.com/i-pul/search-flight/internal/slogx"
	"github.com/valyala/fasthttp"
)

type reqCtxKey struct{}

func Trace(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		traceID := uuid.New().String()
		ctx.Response.Header.Set("X-Trace-Id", traceID)

		reqCtx, cancel := context.WithCancel(context.Background())
		defer cancel()
		reqCtx = context.WithValue(reqCtx, slogx.TraceIDKey, traceID)
		ctx.SetUserValue(reqCtxKey{}, reqCtx)

		next(ctx)
	}
}

func RequestContext(ctx *fasthttp.RequestCtx) context.Context {
	if reqCtx, ok := ctx.UserValue(reqCtxKey{}).(context.Context); ok {
		return reqCtx
	}
	return ctx
}
