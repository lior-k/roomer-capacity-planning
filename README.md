# Load Testing Tool

A flexible load testing tool that automatically adjusts the number of concurrent users based on performance metrics. The tool uses `k6` under the hood to generate load and provides both CLI and web UI interfaces.

## Features

- Automatic load adjustment based on latency and RPS thresholds
- Support for both CLI and web UI interfaces
- Real-time output streaming
- Configurable test duration and thresholds
- Debug mode for detailed output
- Support for different HTTP methods and request bodies

## Prerequisites

1. Install k6:
```bash
# macOS
brew install k6

# Linux
sudo apt-get update && sudo apt-get install k6

# Windows
choco install k6
```

## Installation

1. Clone the repository:
```bash
git clone https://github.com/lior/dev/cursor roomer.git
cd cursor roomer
```

2. Install dependencies:
```bash
go mod download
```

3. Build the executables:
```bash
# Build CLI version
go build -o loadtest-cli cmd/cli/main.go

# Build web UI version
go build -o loadtest-web cmd/web/main.go
```

## Usage

### CLI Version

Run load tests from the command line:

```bash
./loadtest-cli -url http://example.com -goroutines 10 -duration 10s -max-latency-increase 15.0 -min-rps-increase 4.0
```

Available flags:
- `-url`: URL to test (required)
- `-goroutines`: Initial number of virtual users (default: 10)
- `-duration`: Duration for each test cycle (default: 10s)
- `-max-latency-increase`: Maximum allowed latency increase in percentage (default: 15.0)
- `-min-rps-increase`: Minimum required RPS increase in percentage (default: 4.0)
- `-debug`: Enable debug logging to show raw k6 output

### Web UI Version

Start the web server:

```bash
./loadtest-web -port 8080
```

Then open your browser and navigate to `http://localhost:8080`. The web UI provides a form to configure and run load tests with real-time output streaming.

## How It Works

1. The tool starts with a specified number of virtual users (VUs)
2. For each test cycle:
   - Runs a load test for the specified duration
   - Measures latency and RPS metrics
   - If latency increases beyond the threshold, stops increasing load
   - If RPS increase is below the threshold, stops increasing load
   - Otherwise, increases the number of virtual users and continues

## License

MIT License 