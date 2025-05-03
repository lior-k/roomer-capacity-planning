package types

import (
	"context"
	"time"
)

// LoadTestConfig contains the configuration for a load test
type LoadTestConfig struct {
	URL                string
	Goroutines         int
	Duration           time.Duration
	MaxLatencyIncrease float64
	MinRpsIncrease     float64
	Debug              bool
	Ctx                context.Context
	Method             string    // HTTP method (GET, POST, etc.)
	Body               string    // Request body for POST requests
}

// LoadTestResult contains the results of a load test
type LoadTestResult struct {
	RPS    float64
	P50    float64
	P75    float64
	P90    float64
	P99    float64
}

// LoadTestClient interface for different load testing tools
type LoadTestClient interface {
	RunTest(config LoadTestConfig) (string, error)
	Name() string
}

// OutputHandler interface for handling output
type OutputHandler interface {
	WriteLine(line string)
} 