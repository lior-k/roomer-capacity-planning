package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"cursor-roomer/loadtest/runner"
	"cursor-roomer/loadtest/types"
)

type StdoutHandler struct{}

func (h *StdoutHandler) WriteLine(line string) {
	fmt.Println(line)
}

func main() {
	url := flag.String("url", "", "URL to test")
	initialGoroutines := flag.Int("goroutines", 10, "Initial number of goroutines")
	duration := flag.Duration("duration", 10*time.Second, "Duration for each test cycle (e.g. 10s, 1m)")
	latencyThreshold := flag.Float64("max-latency-increase", 15.0, "Maximum allowed latency increase in percentage (e.g. 15.0 for 15%)")
	rpsThreshold := flag.Float64("min-rps-increase", 4.0, "Minimum required RPS increase in percentage (e.g. 4.0 for 4%)")
	debug := flag.Bool("debug", false, "Enable debug logging to show raw k6 output")
	clientType := flag.String("client", "k6", "Load testing client to use (k6, wrk, ghz)")
	flag.Parse()

	if *url == "" {
		log.Fatal("Please provide a URL using -url flag")
	}

	config := types.LoadTestConfig{
		URL:                *url,
		Goroutines:         *initialGoroutines,
		Duration:           *duration,
		MaxLatencyIncrease: *latencyThreshold,
		MinRpsIncrease:     *rpsThreshold,
		Debug:              *debug,
		Ctx:                context.Background(),
	}

	output := &StdoutHandler{}
	if err := runner.RunLoadTest(config, output, *clientType); err != nil {
		log.Fatal(err)
	}
}
