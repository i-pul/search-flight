# Search Flight App

A flight search and aggregation service written in Go that fetches and normalizes flight data from multiple airline providers, then exposes a unified REST API for searching, filtering, and sorting results.

## Features

- **Multi-provider aggregation**: Garuda Indonesia, Lion Air, Batik Air, AirAsia
- **Parallel queries**: All providers queried concurrently via `errgroup`
- **Data normalization**: Handles 4 different response formats and time representations
- **Search, filter & sort**: Price range, stops, time windows, airlines, duration
- **Round-trip search**: Both legs searched in parallel; `return_flights` in response
- **Best value scoring**: Weighted score (0–100) combining price, duration, and stops; configurable via env vars
- **IDR currency formatting**: Locale-aware display (e.g. `Rp1.500.000`) via `golang.org/x/text/currency`
- **Timezone display**: Human-readable timezone labels (`WIB`, `WITA`, `WIT`) on every departure/arrival point
- **Flight validation**: Rejects flights where arrival is not after departure
- **Per-provider caching**: In-process cache (`go-cache`) keyed per provider + search request; errors never cached so unreliable providers always retry fresh; TTL configurable via env vars
- **Retry with exponential backoff**: Transient provider failures retried automatically; configurable attempts and base delay via env vars
- **Mock providers**: Embedded JSON data with realistic delays and AirAsia 10% failure simulation
- **Structured logging**: `log/slog` with JSON output; trace ID injected into every log line
- **Request tracing**: UUID v4 trace ID per request via `X-Trace-Id` header
- **Environment config**: Env vars loaded from `.env` via `godotenv`; mapped to a typed struct via `envconfig`

## Requirements

- Go 1.21+

## Setup

```bash
git clone https://github.com/i-pul/search-flight.git
cd search-flight
go mod download
cp .env.example .env   # edit as needed
```

## Running

```bash
make run        # start server on :8080
make build      # compile to bin/api
make test       # run all tests with race detector
make test-verbose
```

Or directly:

```bash
go run ./cmd/api
```

## Environment Variables

Defined in `.env` (copy from `.env.example`). All variables are optional — defaults apply when unset.

| Variable | Default | Description |
|---|---|---|
| `ADDR` | `:8080` | TCP address the server listens on |
| `PROVIDER_TIMEOUT_MS` | `2000` | Max milliseconds to wait for all provider responses |
| `RETRY_MAX_ATTEMPTS` | `3` | Max attempts per provider call (set to `1` to disable retries) |
| `RETRY_BASE_DELAY_MS` | `100` | Base backoff delay in ms; doubles each attempt with ±25% jitter |
| `BEST_VALUE_WEIGHT_PRICE` | `0.50` | Price factor weight for best-value scoring |
| `BEST_VALUE_WEIGHT_DURATION` | `0.30` | Duration factor weight for best-value scoring |
| `BEST_VALUE_WEIGHT_STOPS` | `0.20` | Stops factor weight for best-value scoring |
| `CACHE_TTL_SECONDS` | `300` | How long a successful provider response is cached (0 = cache never expires until cleanup) |
| `CACHE_CLEANUP_SECONDS` | `300` | How often expired cache entries are purged from memory |

> Weights are automatically normalised to sum to 1, so you can supply any positive combination (e.g. `5 3 2` is equivalent to `0.5 0.3 0.2`).

## API

### `POST /api/v1/flights/search`

Search for available flights across all providers.

**Request Body** (JSON):

```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departureDate": "2025-12-15",
  "returnDate": null,
  "passengers": 1,
  "cabinClass": "economy",
  "filters": {
    "minPrice": 500000,
    "maxPrice": 1000000,
    "maxStops": 0,
    "maxDuration": 180,
    "airlines": ["GA", "JT"],
    "departAfter": "06:00",
    "departBefore": "20:00",
    "arriveAfter": "08:00",
    "arriveBefore": "22:00"
  },
  "sort": {
    "by": "price_asc"
  }
}
```

**Top-level fields:**

| Field | Type | Required | Description |
|---|---|---|---|
| `origin` | string | yes | 3-letter IATA airport code |
| `destination` | string | yes | 3-letter IATA airport code |
| `departureDate` | string | yes | Date in `YYYY-MM-DD` format |
| `returnDate` | string | no | Date in `YYYY-MM-DD` format |
| `passengers` | int | yes | Number of passengers (>= 1) |
| `cabinClass` | string | yes | e.g. `economy`, `business` |
| `filters` | object | no | Optional filter criteria (see below) |
| `sort` | object | no | Optional sort order (see below) |

**`filters` fields** (all optional):

| Field | Type | Example | Description |
|---|---|---|---|
| `minPrice` | float | `500000` | Minimum price in IDR |
| `maxPrice` | float | `1000000` | Maximum price in IDR |
| `maxStops` | int | `0` | Maximum number of stops (0 = direct only) |
| `maxDuration` | int | `180` | Maximum flight duration in minutes |
| `airlines` | []string | `["GA","JT"]` | IATA airline codes to include |
| `departAfter` | string | `"06:00"` | Earliest departure time (HH:MM, WIB) |
| `departBefore` | string | `"20:00"` | Latest departure time (HH:MM, WIB) |
| `arriveAfter` | string | `"08:00"` | Earliest arrival time (HH:MM, WIB) |
| `arriveBefore` | string | `"22:00"` | Latest arrival time (HH:MM, WIB) |

**`sort.by` options**: `price_asc` (default), `price_desc`, `duration_asc`, `duration_desc`, `departure_time`, `arrival_time`, `best_value`

**Response headers:**

| Header | Description |
|---|---|
| `X-Trace-Id` | UUID v4 identifying this request; present in all log lines for correlation |

**Example — Basic search:**

```bash
curl -s -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{"origin":"CGK","destination":"DPS","departureDate":"2025-12-15","passengers":1,"cabinClass":"economy"}' | jq .
```

**Example — Direct flights only, sorted by price:**

```bash
curl -s -X POST http://localhost:8080/api/v1/flights/search \
  -H "Content-Type: application/json" \
  -d '{
    "origin": "CGK",
    "destination": "DPS",
    "departureDate": "2025-12-15",
    "passengers": 1,
    "cabinClass": "economy",
    "filters": { "maxStops": 0, "maxPrice": 1000000 },
    "sort": { "by": "price_asc" }
  }' | jq .
```

**Example Response:**

```json
{
  "search_criteria": {
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "return_date": "2025-12-22",
    "passengers": 1,
    "cabin_class": "economy"
  },
  "metadata": {
    "total_results": 13,
    "return_results": 11,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 412,
    "cache_hit": false
  },
  "flights": [ ... ],
  "return_flights": [ ... ]
}
```

> For one-way searches, `return_date`, `return_results`, and `return_flights` are omitted.

**Error Response:**

```json
{
  "error": "validation_error",
  "message": "origin must be a 3-letter IATA code",
  "code": 400
}
```

## Project Structure

```
.
├── cmd/api/main.go                          # Entry point — wires layers and starts server
├── internal/
│   ├── config/
│   │   └── config.go                        # Typed config struct; loaded via envconfig
│   ├── domain/                              # Shared data types
│   │   ├── flight.go                        # Flight, SearchRequest/Response, Matches()
│   │   └── filter.go                        # FilterParams, SortParams, SortBy constants
│   ├── handler/
│   │   └── flight/                          # HTTP layer (fasthttp)
│   │       ├── handler.go                   # SearchFlights endpoint
│   │       ├── dto.go                       # Request/response types, validation, body-to-domain mapping
│   │       ├── util.go                      # HTTP response helpers (writeJSON, writeError)
│   │       ├── handler_test.go              # Handler unit tests with mock usecase
│   │       └── util_test.go                 # Shared test helpers (writeJSON, decodeError, validBody)
│   ├── middleware/
│   │   └── trace.go                         # Trace middleware — UUID trace ID, proper context.Context bridge
│   ├── slogx/
│   │   ├── handler.go                       # ContextHandler — injects trace_id into every log record
│   │   └── handler_test.go                  # ContextHandler unit tests
│   ├── usecase/
│   │   └── flight/                          # Flight search business logic
│   │       ├── search.go                    # FlightUsecase interface + parallel fan-out orchestration
│   │       ├── search_test.go               # Usecase integration tests
│   │       ├── filter.go                    # ApplyFilters, ApplySort
│   │       ├── filter_test.go
│   │       ├── score.go                     # ComputeBestValueScores, ScoreWeights
│   │       └── score_test.go
│   ├── repository/
│   │   └── flight/                          # Flight data access layer
│   │       ├── repository.go                # Repository interface
│   │       ├── cached/                      # Standalone per-provider cache store (go-cache)
│   │       ├── garuda/                      # Garuda Indonesia (50-100ms delay)
│   │       ├── lionair/                     # Lion Air (100-200ms delay)
│   │       ├── batikair/                    # Batik Air (200-400ms delay)
│   │       └── airasia/                     # AirAsia (50-150ms, 10% failure rate)
│   ├── mockdata/                            # Embedded mock JSON files (compile-time)
│   │   └── mockdata.go
│   └── util/
│       ├── timeutil.go                      # Centralized time parsing helpers
│       ├── retry.go                         # Generic Retry[T] — exponential backoff with jitter
│       └── currency.go                      # FormatPrice — IDR locale-aware formatting
├── .env.example                             # Environment variable template
└── Makefile
```

## Design Decisions

### 1. Layered architecture: handler → usecase → repository
> _Technical: clear separation of concerns · clean, production-ready code_

> _Core: aggregate flight data · search & filter capabilities_

Separating HTTP concerns (handler), business logic (usecase), and data access (repository) keeps each layer independently testable and replaceable. The usecase depends only on the `flight.Repository` interface — swapping in a real HTTP client requires no changes to filtering, sorting, or the HTTP layer.

### 2. Centralized time parsing (`util/timeutil.go`)
> _Core: handle different time formats and time zones · validate flight data_

> _Bonus: timezone conversions (WIB, WITA, WIT)_

All 4 providers use different time formats — this is the highest-risk part of the normalisation pipeline:

| Provider | Format | Example |
|---|---|---|
| Garuda, AirAsia | RFC3339 with colon offset | `2025-12-15T06:00:00+07:00` |
| Lion Air | ISO8601 no-tz + separate IANA zone | `"2025-12-15T05:30:00"` + `"Asia/Jakarta"` |
| Batik Air | ISO8601 no-colon offset + `"1h 45m"` duration | `2025-12-15T07:15:00+0700` |

Centralizing parsing makes each function independently testable and keeps repository adapters focused on structural mapping.

### 3. Non-fatal provider failures
> _Core: aggregate flight data from multiple sources · proper error handling_

When a provider fails (e.g. AirAsia's 10% simulated failure rate), the usecase continues with remaining results and records the failure count in `metadata.providers_failed`. Partial results are returned rather than an error.

### 4. Parallel fan-out with `errgroup`
> _Technical: API performance_

> _Bonus: parallel provider queries_

All repositories are queried concurrently. A `results[i]` slice is pre-allocated so each goroutine writes to its own index — no mutex needed. Results are aggregated sequentially after `eg.Wait()`.

### 5. Structured logging with trace ID (`slogx`, `middleware`)
> _Technical: production-ready code · API performance (request tracing and diagnostics)_

Every request is assigned a UUID v4 trace ID by `middleware.Trace`. The ID is echoed in the `X-Trace-Id` response header and injected into every `slog` record for that request via `slogx.ContextHandler` — a thin wrapper around `slog.Handler` that reads the trace ID from `context.Value` before delegating to the underlying JSON handler.

### 6. IDR currency formatting (`util/currency.go`)
> _Bonus: currency display (IDR formatting with thousands separator)_

The `Price` domain model carries both a raw `amount float64` and a `formatted string` field populated at the repository adapter level (so every provider's flight already has it by the time it reaches the handler). Formatting is handled by `util.FormatPrice`, which:

1. Validates the ISO 4217 currency code via `golang.org/x/text/currency.ParseISO` — unknown codes fall back gracefully to `"CODE amount"` rather than panicking.
2. Extracts the locale-correct symbol (`"Rp"` for IDR, `"US$"` for USD, etc.) using the Indonesian locale printer with `currency.Symbol` — this delegates to CLDR data, keeping symbols accurate without a hardcoded table.
3. Formats the number independently with `message.NewPrinter(language.Indonesian).Sprintf`, which applies Indonesian thousands separators (`.`) and decimal separator (`,`) regardless of CLDR's currency precision rules. This decoupling is intentional: CLDR marks IDR as having 0 decimal places, so `currency.IDR.Amount(850000.53)` would silently round — by formatting the number separately we preserve whatever precision the provider supplies.

### 7. Best value scoring algorithm (`usecase/flight/score.go`)
> _Bonus: "best value" scoring algorithm (price + convenience)_

Scores are computed **after filtering** on the final candidate set. Computing before filtering would mean a flight's score reflects competition from options the user has already excluded — post-filter scores are relative to what the user actually sees.

**Algorithm** — weighted linear penalty model:

```
normFactor = (value - min) / (max - min)   // 0 = best, 1 = worst in set
penalty    = normPrice × wPrice + normDuration × wDuration + normStops × wStops
score      = round((1 - penalty) × 100, 1)  // 0–100, higher = better
```

**Default weights**: price 50%, duration 30%, stops 20%.

| Factor | Weight | Rationale |
|--------|--------|-----------|
| Price | 50% | Primary decision driver for most travelers; the main reason to use an aggregator |
| Duration | 30% | Holistic convenience signal — implicitly captures stop overhead since layover flights are almost always longer |
| Stops | 20% | Separate penalty for boarding/deboarding friction and missed-connection risk, not fully captured by duration |

**Configurability** — weights are read from env vars (`BEST_VALUE_WEIGHT_PRICE`, `BEST_VALUE_WEIGHT_DURATION`, `BEST_VALUE_WEIGHT_STOPS`) and automatically normalised to sum to 1 at startup. Any positive combination is valid: `{5, 3, 2}` produces identical results to `{0.5, 0.3, 0.2}`. All-zero weights fall back to defaults rather than dividing by zero. This lets operators tune the scoring for different audiences (e.g. business travellers who care more about duration than price) without a code change.

The `best_value` sort option is exposed alongside the existing price/duration/time options — it sorts by `BestValueScore` descending.

### 8. Timezone display (`util/timeutil.go`)
> _Bonus: timezone conversions (WIB, WITA, WIT)_

Every `FlightPoint` (departure and arrival) carries a `timezone` string field showing the human-readable Indonesian timezone abbreviation:

| Abbreviation | UTC offset | Provinces |
|---|---|---|
| `WIB` | +07:00 | Java, Sumatra, West Kalimantan |
| `WITA` | +08:00 | Bali, Sulawesi, East/South Kalimantan |
| `WIT` | +09:00 | Maluku, Papua |

**Implementation** — `util.TimezoneAbbr(t time.Time) string` reads the UTC offset in seconds via `t.Zone()` and maps it with a switch. This approach works uniformly across all 4 providers regardless of how their times were parsed:

- Garuda/AirAsia (RFC3339 `+07:00`) — `Zone()` returns offset 25200 → `"WIB"`
- Lion Air (IANA `Asia/Jakarta`) — `Zone()` also returns offset 25200 → `"WIB"`
- Batik Air (no-colon `+0700`) — same offset → `"WIB"`

Non-Indonesian offsets (e.g. UTC, `+05:30`) fall back to `±HH:MM` so the field is never empty and the function never panics.

### 9. Round-trip search (`usecase/flight/search.go`)
> _Bonus: support for round-trip searches_

When `returnDate` is supplied in the request, the usecase runs **two provider fan-outs in parallel** under the shared timeout context:

- Outbound: `origin → destination` on `departureDate`
- Return: `destination → origin` on `returnDate`

Both complete in roughly the same wall-clock time as a single one-way search because the 8 goroutines (4 providers × 2 legs) run concurrently.

**Response shape** — the existing `flights` array holds outbound results; a new `return_flights` array holds return results. Both are independently filtered, scored, and sorted with the same criteria. The only difference: departure/arrival time-of-day windows (`departAfter`, `departBefore`, etc.) are stripped from the return leg's filter because those windows were anchored to the *outbound* departure date — applying them literally to the return date would silently exclude valid flights. Price, stops, duration, and airline filters are applied to both legs unchanged.

For one-way searches all new fields (`return_date`, `return_results`, `return_flights`) are omitted from the JSON response, preserving full backward compatibility.

### 10. Retry with exponential backoff (`util/retry.go`)
> _Bonus: retry logic with exponential backoff for failed providers_

Each provider call inside `queryProviders` is wrapped with `util.Retry` — a generic function that retries any `func() (T, error)`:

```
attempt 1 → immediate
attempt 2 → baseDelay × [0.75, 1.25)
attempt 3 → baseDelay×2 × [0.75, 1.25)
…
```

**Design choices:**

- **Generic utility, not a struct wrapper** — `Retry[T any]` takes a function as a parameter, keeping the retry logic decoupled from the `Repository` interface. This makes it reusable for any future callsite (e.g. outbound HTTP calls, cache misses) without structural changes.
- **±25% jitter** — randomising each delay window prevents multiple providers from retrying in lock-step after a shared transient failure (thundering herd).
- **Context-aware** — the select on `ctx.Done()` stops retrying immediately when the shared timeout fires; no goroutine lingers beyond the deadline.
- **`RETRY_MAX_ATTEMPTS=1` disables retries** — zero-overhead path when retries are not desired; `Retry` returns the original repo unchanged in this case.
- **Configurable via env vars** (`RETRY_MAX_ATTEMPTS`, `RETRY_BASE_DELAY_MS`) so operators can tune aggressiveness without a code change — useful when switching from mock providers to real HTTP backends with different failure characteristics.

### 11. Per-provider caching (`repository/flight/cached`)

Each provider's successful responses are cached independently in `cached.Store` — a thin wrapper around `go-cache` that lives in its own package and knows nothing about how flights are fetched.

**Why per-provider, not aggregated:**

AirAsia has a 10% simulated failure rate. If the cache held aggregated results, a response with AirAsia missing would be served for the full TTL — every cache hit would silently under-report capacity. Per-provider caching means only successful responses are stored; a failed provider always retries on the next request regardless of whether the other three are cached.

**Separation of concerns — standalone `Store`, not a decorator:**

The cache is not a `Repository` wrapper (decorator pattern). Instead, `cached.Store` exposes `Get(providerName, req)` and `Set(providerName, req, flights)` and is passed as a dependency to the usecase. The repository layer stays focused on fetching; the orchestration layer decides when to read/write the cache. This keeps the two concerns independently changeable and independently testable.

**Cache key**: `providerName:origin:destination:departureDate:passengers:cabinClass` — all six fields that determine a unique result set. Filters and sort are **not** part of the key because they are applied post-fetch; a cached raw result can serve many different filter/sort combinations without a cache miss.

**Errors are never cached** — if a provider call fails (including after all retry attempts), nothing is written to the store. The next request will retry the provider fresh.

**TTL** is configurable via `CACHE_TTL_SECONDS` (default 5 min). `CACHE_CLEANUP_SECONDS` controls how often go-cache sweeps expired entries from memory (default 5 min).
