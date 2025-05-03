package webui

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"cursor-roomer/loadtest/runner"
	"cursor-roomer/loadtest/types"
)

type Server struct {
	port int
	mu   sync.Mutex
	ctx  context.Context
}

func NewServer(port int) *Server {
	return &Server{
		port: port,
		ctx:  context.Background(),
	}
}

type SSEOutputHandler struct {
	w http.ResponseWriter
}

func (h *SSEOutputHandler) WriteLine(line string) {
	// Send each line as a separate SSE event to preserve formatting
	lines := strings.Split(line, "\n")
	for _, l := range lines {
		fmt.Fprintf(h.w, "data: %s\n\n", l)
		if f, ok := h.w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (s *Server) handleStopTest(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx != nil {
		// Cancel the current context
		if cancel, ok := s.ctx.Value("cancel").(context.CancelFunc); ok {
			cancel()
		}
		// Create a new context for future tests
		s.ctx = context.Background()
	}

	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRunTest(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create output handler
	output := &SSEOutputHandler{w: w}

	// Try to parse request body first
	var req struct {
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

	// Try to decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If JSON decode fails, try query parameters
		req.URL = r.URL.Query().Get("url")
		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		// Validate URL format
		if _, err := url.Parse(req.URL); err != nil {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}

		goroutines, err := strconv.Atoi(r.URL.Query().Get("goroutines"))
		if err != nil {
			http.Error(w, "Invalid goroutines value", http.StatusBadRequest)
			return
		}
		req.Goroutines = goroutines

		duration, err := strconv.Atoi(r.URL.Query().Get("duration"))
		if err != nil {
			http.Error(w, "Invalid duration value", http.StatusBadRequest)
			return
		}
		req.Duration = fmt.Sprintf("%ds", duration)

		req.MaxLatencyIncrease, err = strconv.ParseFloat(r.URL.Query().Get("latencyThreshold"), 64)
		if err != nil {
			http.Error(w, "Invalid latency threshold value", http.StatusBadRequest)
			return
		}

		req.MinRpsIncrease, err = strconv.ParseFloat(r.URL.Query().Get("rpsThreshold"), 64)
		if err != nil {
			http.Error(w, "Invalid RPS threshold value", http.StatusBadRequest)
			return
		}

		req.Debug = r.URL.Query().Get("debug") == "true"
		req.Method = r.URL.Query().Get("method")
		req.Body = r.URL.Query().Get("body")
		req.ClientType = r.URL.Query().Get("clientType")
		if req.ClientType == "" {
			req.ClientType = "k6" // Default to k6 if not specified
		}
	} else {
		// Validate URL format for JSON request
		if _, err := url.Parse(req.URL); err != nil {
			http.Error(w, "Invalid URL format", http.StatusBadRequest)
			return
		}
		if req.ClientType == "" {
			req.ClientType = "k6" // Default to k6 if not specified
		}
	}

	// Parse duration string
	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		http.Error(w, "Invalid duration format", http.StatusBadRequest)
		return
	}

	// Create a new context for this test
	s.mu.Lock()
	ctx, cancel := context.WithCancel(s.ctx)
	s.ctx = ctx
	s.ctx = context.WithValue(s.ctx, "cancel", cancel)
	s.mu.Unlock()

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

	if err := runner.RunLoadTest(config, output, req.ClientType); err != nil {
		fmt.Fprintf(w, "data: Error: %v\n\n", err)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("webui/templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// RunLoadTest runs a load test with the given configuration and returns the output
func RunLoadTest(config types.LoadTestConfig, clientType string) (string, error) {
	output := &StringOutputHandler{}
	if err := runner.RunLoadTest(config, output, clientType); err != nil {
		return "", err
	}
	return output.String(), nil
}

// StringOutputHandler implements types.OutputHandler and collects output as a string
type StringOutputHandler struct {
	output []string
}

func (h *StringOutputHandler) WriteLine(line string) {
	h.output = append(h.output, line)
}

func (h *StringOutputHandler) String() string {
	return fmt.Sprintf("%s\n", h.output)
}

func (s *Server) Start() error {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/run-test", s.handleRunTest)
	http.HandleFunc("/stop-test", s.handleStopTest)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting web server on %s", addr)
	return http.ListenAndServe(addr, nil)
}
