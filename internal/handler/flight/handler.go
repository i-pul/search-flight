package flight

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/i-pul/search-flight/internal/middleware"
	flightuc "github.com/i-pul/search-flight/internal/usecase/flight"
	"github.com/valyala/fasthttp"
)

type Handler struct {
	uc flightuc.FlightUsecase
}

func New(uc flightuc.FlightUsecase) *Handler {
	return &Handler{uc: uc}
}

// SearchFlights handles POST /api/v1/flights/search.
func (h *Handler) SearchFlights(ctx *fasthttp.RequestCtx) {
	reqCtx := middleware.RequestContext(ctx)
	start := time.Now()

	var req searchHTTPRequest
	if err := json.Unmarshal(ctx.PostBody(), &req); err != nil {
		slog.WarnContext(reqCtx, "invalid request body", "error", err)
		if err := writeError(ctx, fasthttp.StatusBadRequest, "invalid_request", "request body is not valid JSON: "+err.Error()); err != nil {
			slog.ErrorContext(reqCtx, "failed to encode error response", "error", err)
		}
		return
	}

	slog.InfoContext(reqCtx, "search request", "request", req)

	if err := validateSearchRequest(req); err != nil {
		slog.WarnContext(reqCtx, "validation error", "error", err)
		if err := writeError(ctx, fasthttp.StatusBadRequest, "validation_error", err.Error()); err != nil {
			slog.ErrorContext(reqCtx, "failed to encode error response", "error", err)
		}
		return
	}

	fp, err := toFilterParams(req.Filters, req.DepartureDate)
	if err != nil {
		slog.WarnContext(reqCtx, "invalid filter params", "error", err)
		if err := writeError(ctx, fasthttp.StatusBadRequest, "invalid_filter", err.Error()); err != nil {
			slog.ErrorContext(reqCtx, "failed to encode error response", "error", err)
		}
		return
	}

	sp := toSortParams(req.Sort)

	resp, err := h.uc.Search(reqCtx, toSearchRequest(req), fp, sp)
	if err != nil {
		slog.ErrorContext(reqCtx, "search failed", "error", err)
		if err := writeError(ctx, fasthttp.StatusInternalServerError, "search_error", err.Error()); err != nil {
			slog.ErrorContext(reqCtx, "failed to encode error response", "error", err)
		}
		return
	}

	slog.InfoContext(reqCtx, "search response",
		"status", fasthttp.StatusOK,
		"flights", resp.Metadata.TotalResults,
		"providers_ok", resp.Metadata.ProvidersSucceeded,
		"providers_failed", resp.Metadata.ProvidersFailed,
		"duration_ms", time.Since(start).Milliseconds(),
	)

	if err := writeJSON(ctx, fasthttp.StatusOK, resp); err != nil {
		slog.ErrorContext(reqCtx, "failed to encode response", "error", err)
		if err := writeError(ctx, fasthttp.StatusInternalServerError, "response_error", "failed to encode response: "+err.Error()); err != nil {
			slog.ErrorContext(reqCtx, "failed to encode error response", "error", err)
		}
		return
	}
}
