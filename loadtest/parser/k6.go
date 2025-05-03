package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"cursor-roomer/loadtest/types"
)

// ParseK6Output parses the output from k6 command into a LoadTestResult
func ParseK6Output(output string) (*types.LoadTestResult, error) {
	lines := strings.Split(output, "\n")
	result := &types.LoadTestResult{}

	// Regular expressions for parsing k6 output
	rpsRegex := regexp.MustCompile(`http_reqs\.*:\s*(\d+\.?\d*)\s*(\d+\.?\d*)/s`)
	p50Regex := regexp.MustCompile(`http_req_duration.*p\(50\)\s*=\s*(\d+\.?\d*)(ms|s)`)
	p75Regex := regexp.MustCompile(`http_req_duration.*p\(75\)\s*=\s*(\d+\.?\d*)(ms|s)`)
	p90Regex := regexp.MustCompile(`http_req_duration.*p\(90\)\s*=\s*(\d+\.?\d*)(ms|s)`)
	p99Regex := regexp.MustCompile(`http_req_duration.*p\(99\)\s*=\s*(\d+\.?\d*)(ms|s)`)

	// Try to find the summary section
	summaryStart := -1
	for i, line := range lines {
		if strings.Contains(line, "http_req_duration") {
			summaryStart = i
			break
		}
	}

	if summaryStart == -1 {
		return nil, fmt.Errorf("could not find summary section in k6 output")
	}

	// Parse the summary section
	for i := summaryStart; i < len(lines); i++ {
		line := lines[i]
		// Parse RPS
		if matches := rpsRegex.FindStringSubmatch(line); len(matches) > 2 {
			fmt.Printf("Found RPS line: %s\n", line)
			fmt.Printf("Matches: %v\n", matches)
			rps, err := strconv.ParseFloat(matches[2], 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse RPS: %v", err)
			}
			result.RPS = rps
		} else if strings.Contains(line, "http_reqs") {
			fmt.Printf("Found http_reqs line but didn't match: %s\n", line)
		}

		// Parse P50
		if matches := p50Regex.FindStringSubmatch(line); len(matches) > 2 {
			val, err := parseK6Latency(matches[1], matches[2])
			if err != nil {
				return nil, err
			}
			result.P50 = val
		}

		// Parse P75
		if matches := p75Regex.FindStringSubmatch(line); len(matches) > 2 {
			val, err := parseK6Latency(matches[1], matches[2])
			if err != nil {
				return nil, err
			}
			result.P75 = val
		}

		// Parse P90
		if matches := p90Regex.FindStringSubmatch(line); len(matches) > 2 {
			val, err := parseK6Latency(matches[1], matches[2])
			if err != nil {
				return nil, err
			}
			result.P90 = val
		}

		// Parse P99
		if matches := p99Regex.FindStringSubmatch(line); len(matches) > 2 {
			val, err := parseK6Latency(matches[1], matches[2])
			if err != nil {
				return nil, err
			}
			result.P99 = val
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

// parseK6Latency converts a k6 latency value to milliseconds
func parseK6Latency(value, unit string) (float64, error) {
	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}

	switch unit {
	case "ms":
		return val, nil
	case "s":
		return val * 1000, nil
	default:
		return 0, fmt.Errorf("unknown time unit in k6 latency value: %s", unit)
	}
}
