package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"time"

	"cursor-roomer/loadtest/types"
	"cursor-roomer/webui"
)

type TestRequest struct {
	URL                string  `json:"url"`
	Goroutines         int     `json:"goroutines"`
	Duration           string  `json:"duration"`
	MaxLatencyIncrease float64 `json:"maxLatencyIncrease"`
	MinRpsIncrease     float64 `json:"minRpsIncrease"`
	Debug              bool    `json:"debug"`
	Method             string  `json:"method"`
	Body               string  `json:"body"`
	ClientType         string  `json:"clientType"`
}

func handleStartTest(w http.ResponseWriter, r *http.Request) {
	var req TestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		http.Error(w, "invalid duration format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure we clean up the context

	config := types.LoadTestConfig{
		URL:                req.URL,
		Goroutines:         req.Goroutines,
		Duration:           duration,
		MaxLatencyIncrease: req.MaxLatencyIncrease,
		MinRpsIncrease:     req.MinRpsIncrease,
		Debug:              req.Debug,
		Ctx:                ctx,
		Method:             req.Method,
		Body:               req.Body,
	}

	// Create a channel to receive the test result
	resultChan := make(chan string, 1)
	errChan := make(chan error, 1)

	// Start the test in a goroutine
	go func() {
		output, err := webui.RunLoadTest(config, req.ClientType)
		if err != nil {
			errChan <- err
			return
		}
		resultChan <- output
	}()

	// Wait for the test to complete or context cancellation
	select {
	case result := <-resultChan:
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(result))
	case err := <-errChan:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	case <-ctx.Done():
		http.Error(w, "test cancelled", http.StatusServiceUnavailable)
	}
}

func main() {
	port := flag.Int("port", 8080, "Port to run the web server on")
	flag.Parse()

	server := webui.NewServer(*port)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
