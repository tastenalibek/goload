# goload

A fast, concurrent HTTP load testing tool written in Go.

```
Sending 500 GET requests to https://api.example.com (concurrency: 25)

  [██████████████████████████████] 100% (500/500)
─────────────────────────────────────────
 Summary
─────────────────────────────────────────
  Total Requests  500
  Successful      498
  Failed          2
  Total Time      4.32s
  Requests/sec    115.74
─────────────────────────────────────────
 Latency
─────────────────────────────────────────
  Min   48.21 ms
  Mean  213.44 ms
  p50   195.30 ms
  p90   380.12 ms
  p99   512.88 ms
  Max   891.04 ms
─────────────────────────────────────────
```

## Features

- **Concurrent workers** — configurable goroutine pool fires requests in parallel
- **Live progress bar** — real-time `[████░░░] 67%` display updated every 100ms
- **Latency percentiles** — p50, p90, p99 plus min/mean/max
- **JSON output** — pipe results into `jq` or other tools with `--json`
- **Zero runtime dependencies** beyond the standard library (cobra is CLI-only)

## Installation

```bash
go install github.com/tastenalibek/goload@latest
```

Or build from source:

```bash
git clone https://github.com/tastenalibek/goload
cd goload
go build -o goload .
```

## Usage

```
goload <url> [flags]

Flags:
  -n, --requests int      total number of requests to send (default 100)
  -c, --concurrency int   number of concurrent workers (default 10)
  -m, --method string     HTTP method (default "GET")
  -t, --timeout int       per-request timeout in seconds (default 30)
      --json              output results as JSON
  -h, --help              help for goload
```

### Examples

```bash
# Quick smoke test — 100 requests, 10 workers
goload https://example.com

# Stress test — 1000 requests, 50 workers
goload https://api.example.com/health -n 1000 -c 50

# POST endpoint
goload https://api.example.com/submit -m POST -n 200 -c 20

# Machine-readable output
goload https://example.com -n 500 --json | jq '.rps'
```

## Project structure

```
goload/
├── main.go                  # entry point
├── cmd/
│   └── root.go              # CLI flags and command wiring (cobra)
└── internal/
    ├── runner/
    │   └── runner.go        # concurrent request engine + stats
    └── report/
        └── report.go        # table and JSON output formatters
```

## How it works

1. A buffered **jobs channel** is pre-filled with `n` tokens (one per request).
2. `c` goroutines drain the channel concurrently — each goroutine fires one HTTP request per token and sends the result to a **results channel**.
3. A separate goroutine reads an **atomic counter** every 100ms and renders the progress bar with `\r`.
4. Once all workers finish, the results channel is drained to compute latency percentiles via sorting.

## License

MIT
