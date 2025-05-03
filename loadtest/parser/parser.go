package parser

import (
	"fmt"
	"strconv"
	"strings"

	"cursor-roomer/loadtest/types"
)

// ParseWRKOutput parses the output from wrk command into a LoadTestResult
func ParseWRKOutput(output string) (*types.LoadTestResult, error) {
	lines := strings.Split(output, "\n")
	result := &types.LoadTestResult{}

	for _, line := range lines {
		// Parse RPS
		if strings.HasPrefix(line, "Requests/sec:") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				rps, err := strconv.ParseFloat(parts[len(parts)-1], 64)
				if err != nil {
					return nil, fmt.Errorf("failed to parse RPS: %v", err)
				}
				result.RPS = rps
			}
		}
		// Parse Latencies from Latency Distribution section
		if strings.Contains(line, "Latency Distribution") {
			// Skip the header line
			continue
		}
		if strings.HasPrefix(line, "     50%") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				latencyStr := strings.TrimSpace(parts[1])
				if val, err := ParseLatency(latencyStr); err == nil {
					result.P50 = val
				}
			}
		}
		if strings.HasPrefix(line, "     75%") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				latencyStr := strings.TrimSpace(parts[1])
				if val, err := ParseLatency(latencyStr); err == nil {
					result.P75 = val
				}
			}
		}
		if strings.HasPrefix(line, "     90%") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				latencyStr := strings.TrimSpace(parts[1])
				if val, err := ParseLatency(latencyStr); err == nil {
					result.P90 = val
				}
			}
		}
		if strings.HasPrefix(line, "     99%") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				latencyStr := strings.TrimSpace(parts[1])
				if val, err := ParseLatency(latencyStr); err == nil {
					result.P99 = val
				}
			}
		}
	}

	// Verify we got all values
	if result.RPS == 0 {
		return nil, fmt.Errorf("failed to parse RPS from output")
	}
	if result.P50 == 0 || result.P75 == 0 || result.P90 == 0 || result.P99 == 0 {
		return nil, fmt.Errorf("failed to parse latency percentiles from output")
	}

	return result, nil
}

// ParseLatency converts a latency string (e.g. "123.45ms" or "1.03s") to milliseconds
func ParseLatency(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "ms") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(s, "ms"), 64)
		if err != nil {
			return 0, err
		}
		return val, nil
	} else if strings.HasSuffix(s, "s") {
		val, err := strconv.ParseFloat(strings.TrimSuffix(s, "s"), 64)
		if err != nil {
			return 0, err
		}
		return val * 1000, nil
	}
	return 0, fmt.Errorf("unknown time unit in latency value: %s", s)
} 