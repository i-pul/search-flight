package main

import (
	"log/slog"
	"os"

	"github.com/fasthttp/router"
	"github.com/i-pul/search-flight/internal/config"
	flighth "github.com/i-pul/search-flight/internal/handler/flight"
	"github.com/i-pul/search-flight/internal/middleware"
	"github.com/i-pul/search-flight/internal/repository/flight"
	"github.com/i-pul/search-flight/internal/repository/flight/airasia"
	"github.com/i-pul/search-flight/internal/repository/flight/batikair"
	"github.com/i-pul/search-flight/internal/repository/flight/garuda"
	"github.com/i-pul/search-flight/internal/repository/flight/lionair"
	"github.com/i-pul/search-flight/internal/slogx"
	flightuc "github.com/i-pul/search-flight/internal/usecase/flight"
	"github.com/joho/godotenv"
	"github.com/valyala/fasthttp"
)

func main() {
	godotenv.Load()

	slog.SetDefault(slog.New(slogx.NewContextHandler(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))))

	cfg := config.Load()

	repos := []flight.Repository{
		garuda.New(),
		lionair.New(),
		batikair.New(),
		airasia.New(),
	}

	uc := flightuc.New(repos)
	h := flighth.New(uc)

	r := router.New()
	r.POST("/api/v1/flights/search", middleware.Trace(h.SearchFlights))

	slog.Info("flight search service starting", "addr", cfg.Addr)
	if err := fasthttp.ListenAndServe(cfg.Addr, r.Handler); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
