package runner

import (
	"fmt"

	"cursor-roomer/loadtest/client"
	"cursor-roomer/loadtest/parser"
	"cursor-roomer/loadtest/types"
)

// TestRunner handles the load test execution
type TestRunner struct {
	config     types.LoadTestConfig
	output     types.OutputHandler
	client     types.LoadTestClient
	initialP90 float64
	lastRPS    float64
}

func NewTestRunner(config types.LoadTestConfig, output types.OutputHandler, client types.LoadTestClient) *TestRunner {
	return &TestRunner{
		config: config,
		output: output,
		client: client,
	}
}

func (r *TestRunner) printResults(result *types.LoadTestResult, prefix string) {
	r.output.WriteLine(fmt.Sprintf("%s results:", prefix))
	r.output.WriteLine(fmt.Sprintf("RPS: %.2f", result.RPS))
	r.output.WriteLine(fmt.Sprintf("P50: %.2fms", result.P50))
	r.output.WriteLine(fmt.Sprintf("P75: %.2fms", result.P75))
	r.output.WriteLine(fmt.Sprintf("P90: %.2fms", result.P90))
	r.output.WriteLine(fmt.Sprintf("P99: %.2fms", result.P99))
}

func (r *TestRunner) runInitialTest() (*types.LoadTestResult, error) {
	// Force 1 virtual user for initial test
	r.config.Goroutines = 1

	r.output.WriteLine(fmt.Sprintf("Running initial test with 1 virtual user for %v...", r.config.Duration))
	r.output.WriteLine(fmt.Sprintf("Will stop if P90 latency increases by more than %.1f%% or RPS increase is less than %.1f%%\n",
		r.config.MaxLatencyIncrease, r.config.MinRpsIncrease))

	output, err := r.client.RunTest(r.config)
	if err != nil {
		return nil, fmt.Errorf("failed to run initial test: %v", err)
	}

	var result *types.LoadTestResult
	switch r.client.Name() {
	case "k6":
		result, err = parser.ParseK6Output(output)
	case "wrk":
		result, err = parser.ParseWRKOutput(output)
	case "ghz":
		return nil, fmt.Errorf("ghz client not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported client type: %s", r.client.Name())
	}

	if err != nil {
		return nil, err
	}

	r.printResults(result, "Initial")
	r.initialP90 = result.P90
	r.lastRPS = result.RPS

	return result, nil
}

func (r *TestRunner) calculateNextThreads(currentThreads int, originalThreads int) int {
	if currentThreads == 1 && originalThreads > 1 {
		// Second test: use configured goroutines
		return originalThreads
	}
	// Subsequent tests: increase by 50% with round up
	return int(float64(currentThreads)*1.5 + 0.5)
}

func (r *TestRunner) runIteration(currentThreads int) (*types.LoadTestResult, error) {
	r.config.Goroutines = currentThreads
	output, err := r.client.RunTest(r.config)
	if err != nil {
		if r.config.Ctx.Err() != nil {
			return nil, fmt.Errorf("test cancelled")
		}
		return nil, fmt.Errorf("failed to run test: %v", err)
	}

	var result *types.LoadTestResult
	switch r.client.Name() {
	case "k6":
		result, err = parser.ParseK6Output(output)
	case "wrk":
		result, err = parser.ParseWRKOutput(output)
	case "ghz":
		return nil, fmt.Errorf("ghz client not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported client type: %s", r.client.Name())
	}

	if err != nil {
		return nil, err
	}

	r.printResults(result, "Current")

	latencyIncrease := (result.P90 - r.initialP90) / r.initialP90 * 100
	rpsIncrease := (result.RPS - r.lastRPS) / r.lastRPS * 100

	r.output.WriteLine(fmt.Sprintf("P90 Latency increase: %.1f%%", latencyIncrease))
	r.output.WriteLine(fmt.Sprintf("RPS increase: %.1f%%\n", rpsIncrease))

	if latencyIncrease > r.config.MaxLatencyIncrease {
		return nil, fmt.Errorf("stopping: P90 latency increased by %.1f%% (threshold: %.1f%%)",
			latencyIncrease, r.config.MaxLatencyIncrease)
	}

	if rpsIncrease < r.config.MinRpsIncrease {
		return nil, fmt.Errorf("stopping: RPS increased by only %.1f%% (threshold: %.1f%%)",
			rpsIncrease, r.config.MinRpsIncrease)
	}

	r.lastRPS = result.RPS
	return result, nil
}

// RunLoadTest executes the load test with the given configuration
func RunLoadTest(config types.LoadTestConfig, output types.OutputHandler, clientType string) error {
	var testClient types.LoadTestClient
	switch clientType {
	case "k6":
		testClient = client.NewK6Client(output)
	case "wrk":
		testClient = client.NewWRKClient(output)
	case "ghz":
		testClient = client.NewGHZClient(output)
	default:
		return fmt.Errorf("unsupported client type: %s", clientType)
	}

	runner := NewTestRunner(config, output, testClient)
	originalThreads := config.Goroutines

	// Run initial test with 1 thread
	if _, err := runner.runInitialTest(); err != nil {
		return err
	}

	// Start with 1 thread for the initial test
	currentThreads := 1

	for {
		select {
		case <-config.Ctx.Done():
			output.WriteLine("\nTest terminated by user")
			return nil
		default:
			// Calculate next thread count
			currentThreads = runner.calculateNextThreads(currentThreads, originalThreads)

			if _, err := runner.runIteration(currentThreads); err != nil {
				output.WriteLine(err.Error())
				return nil
			}
		}
	}
}
