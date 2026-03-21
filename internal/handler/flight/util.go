package flight

import (
	"encoding/json"

	"github.com/valyala/fasthttp"
)

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func writeError(ctx *fasthttp.RequestCtx, code int, errType, message string) error {
	return writeJSON(ctx, code, errorResponse{
		Error:   errType,
		Message: message,
		Code:    code,
	})
}

func writeJSON(ctx *fasthttp.RequestCtx, code int, v any) error {
	ctx.SetContentType("application/json; charset=utf-8")
	ctx.SetStatusCode(code)
	enc := json.NewEncoder(ctx)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
