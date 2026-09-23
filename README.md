# BookCabin Flight Search

Go service that queries four mocked airline providers concurrently, normalizes their responses into one model, and returns filtered, de-duplicated, ranked results. Supports one-way, round-trip, and multi-city search.

## Quick start

Requirements: Go 1.26+.

```bash
make run     # build to bin/bookcabin and start on :8080
make test    # go test -race ./...
```

Config: `files/etc/bookcabin.development.yaml` (timeouts, cache TTL, per-provider latency, failure rate, rate limit, retry).

## Usage

All request variants, with tests, are in the Postman collection: [`docs/postman/bookcabin.postman_collection.json`](docs/postman/bookcabin.postman_collection.json).

**One-way:** `GET` or `POST /search/flight/v1`

```bash
curl "localhost:8080/search/flight/v1?origin=CGK&destination=DPS&departureDate=2025-12-15&passengers=1&cabinClass=economy&maxStops=0&sortBy=price"
```

- Filters: `minPrice`, `maxPrice`, `maxStops`, `maxDuration` (minutes), `airlines` (`GA,JT`), `departureFrom`/`To`, `arrivalFrom`/`To` (`HH:MM`, airport local time)
- Sort: `sortBy` = `best_value` (default), `price`, `duration`, `departure_time`, `arrival_time`; `sortOrder` = `asc`/`desc`

**Round-trip / multi-city:** `POST /multi-search/flight/v1` with 2–6 legs, each shaped like a one-way request. Returns `{"legs": [...]}`, one search result per leg, in order.

```json
{ "legs": [
  { "origin": "CGK", "destination": "DPS", "departureDate": "2025-12-15", "passengers": 1, "cabinClass": "economy" },
  { "origin": "DPS", "destination": "CGK", "departureDate": "2025-12-20", "passengers": 1, "cabinClass": "economy" }
] }
```

## Design choices

- **Clean architecture:** `delivery` (binding, validation, status codes) -> `usecase` (orchestration) -> `repository` (one adapter per provider, mapping each format to a unified `Flight`).
- **Fan-out / fan-in:** providers are queried concurrently under a 2s global timeout; each goroutine writes to its own slot, results merge after `Wait()`. Latency ≈ slowest provider. Multi-search fans out over legs the same way.
- **Resilience:** per-provider token-bucket rate limiter (`x/time/rate`) and retry with exponential backoff + jitter (`avast/retry-go`) for transient failures only.
- **Search vs. filter:** route/date/cabin select the flights (mocks ignore params, so the aggregator enforces them); user filters and sorting run per request after the cache. Dates use the departure airport's local time.
- **Price comparison:** de-duplicate by flight number + departure time, keep the cheapest.
- **Cache:** in-memory, 1 min TTL, keyed by route/date/cabin (filters excluded, so they still hit). Partial results are not cached.
- **Best value:** `0.5·price + 0.3·duration + 0.2·stops`, each **min-max normalized** to 0..1 over the filtered set (lower is better). Without normalization rupiah dominates and the ranking equals price order.
- **Time zones:** three formats parsed into offset-aware times (RFC 3339, `+0700`, local time + IANA zone). Durations use absolute time, so WIB/WITA/WIT differences and layovers are counted correctly.

## Mock data notes

Provided mocks only cover CGK->DPS on 2025-12-15; a few flights were appended (DPS->CGK, DPS->SUB, SUB->CGK) to demonstrate multi-search. The original 13 flights are unchanged.

| Issue in provided mocks | Handling |
|---|---|
| Batik ID7042 `travelTime` 3h 5m, actual 4h 5m | Duration from timestamps |
| Garuda GA315 top level says arrival SUB, 0 stops | Use segments: DPS, 1 stop, 3h 45m |
| Garuda GA332 segment 90 min, actual 30 | Provider durations ignored |
| `expected_result.json` timestamp is for 2024 | Computed from parsed times |

Assumptions: GA315->GA332 is one CGK->DPS product; Garuda baggage numbers are pieces; Batik class `Y` is economy.

## Complexity

O(n) for normalization, matching, de-duplication, filtering, and scoring; O(n log n) sort; O(1) cache lookup.