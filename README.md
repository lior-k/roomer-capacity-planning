# Roomer Capacity Planning Tool

A load testing tool designed to help determine the optimal number of concurrent users a system can handle while maintaining acceptable performance metrics.

## Features

- Multiple load testing clients support:
  - k6
  - wrk
  - ghz (planned)
- Web UI for easy configuration and monitoring
- CLI interface for automation
- Automatic scaling of virtual users
- Performance metrics tracking:
  - Requests Per Second (RPS)
  - Latency percentiles (P50, P75, P90, P99)
- Configurable thresholds for:
  - Maximum latency increase
  - Minimum RPS increase

## Installation

### Prerequisites

- Go 1.16 or later
- k6 (for k6 client)
- wrk (for wrk client)

### Building

```bash
go build -o roomer ./cmd/cli
go build -o roomer-web ./cmd/web
```

## Usage

### Web UI

Start the web server:
```bash
./roomer-web
```

Access the web interface at `http://localhost:8080`

### CLI

```bash
./roomer -url http://example.com -goroutines 10 -duration 30s -max-latency-increase 50 -min-rps-increase 20 -client k6
```

### Parameters

- `url`: Target URL to test
- `goroutines`: Initial number of virtual users
- `duration`: Test duration (e.g., "30s", "1m")
- `max-latency-increase`: Maximum allowed latency increase percentage
- `min-rps-increase`: Minimum required RPS increase percentage
- `client`: Load testing client to use (k6, wrk, ghz)
- `method`: HTTP method (GET, POST, etc.)
- `body`: Request body for POST requests
- `debug`: Enable debug output

## Test Sequence

The tool follows a specific sequence to determine optimal capacity:

1. Initial test with 1 virtual user to establish baseline latency
2. Second test with configured number of virtual users
3. Subsequent tests with 50% increase in virtual users (rounded up)
4. Continues until either:
   - Latency threshold is exceeded
   - RPS increase threshold is not met
   - Test is cancelled

## Output

The tool provides detailed output including:
- Current number of virtual users
- RPS measurements
- Latency percentiles
- Percentage changes in metrics
- Test termination reason

## Development

### Project Structure

```
.
├── cmd/
│   ├── cli/        # Command-line interface
│   └── web/        # Web server
├── loadtest/
│   ├── client/     # Load testing clients
│   ├── parser/     # Output parsers
│   ├── runner/     # Test runner
│   └── types/      # Common types
└── webui/          # Web UI components
```

### Adding New Clients

To add a new load testing client:
1. Implement the `LoadTestClient` interface in `loadtest/types/types.go`
2. Create a new client package in `loadtest/client/`
3. Add client initialization in `RunLoadTest` function

## License

MIT License 