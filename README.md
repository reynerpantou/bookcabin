# BookCabin Flight Search

A flight search and aggregation service written in Go. It queries four (mocked) airline providers in parallel, normalizes their different response formats into one model, and returns filtered, de-duplicated, and ranked results.

## Quick start

Requirements: Go 1.26+.

```bash
make run     # build to bin/bookcabin and start on :8080
make test    # go test -race ./...
```
