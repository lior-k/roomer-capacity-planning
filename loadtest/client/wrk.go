package client

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cursor-roomer/loadtest/types"
)

// WRKClient implements LoadTestClient using the wrk command
type WRKClient struct {
	output types.OutputHandler
}

func NewWRKClient(output types.OutputHandler) types.LoadTestClient {
	return &WRKClient{output: output}
}

func (c *WRKClient) Name() string {
	return "wrk"
}

func (c *WRKClient) createLuaScript(body string, method string) (string, error) {
	// Create a temporary directory for the script
	tmpDir := os.TempDir()

	// Generate a random filename with timestamp and random number
	rand.Seed(time.Now().UnixNano())
	randomNum := rand.Intn(10000)
	timestamp := time.Now().Format("20060102150405")
	scriptPath := filepath.Join(tmpDir, fmt.Sprintf("wrk-script-%s-%d.lua", timestamp, randomNum))

	// Create the Lua script content
	scriptContent := fmt.Sprintf(`wrk.method = "%s"
wrk.headers["Content-Type"] = "application/json"
wrk.body = [[%s]]`, method, body)

	// Write the script to a temporary file
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create Lua script: %v", err)
	}

	return scriptPath, nil
}

func (c *WRKClient) RunTest(config types.LoadTestConfig) (string, error) {
	// Check if wrk is installed
	if _, err := exec.LookPath("wrk"); err != nil {
		return "", fmt.Errorf("wrk is not installed. Please install it first: %v", err)
	}

	maxThreads := runtime.NumCPU() * 2
	threads := config.Goroutines
	if threads > maxThreads {
		threads = maxThreads
	}

	c.output.WriteLine(fmt.Sprintf("Running test with %d threads and %d connections for %v...", threads, config.Goroutines, config.Duration))

	args := []string{
		"-t", fmt.Sprintf("%d", threads),
		"-c", fmt.Sprintf("%d", config.Goroutines),
		"-d", fmt.Sprintf("%.0f", config.Duration.Seconds()),
		"--latency",
	}

	// Handle non-GET requests using Lua script
	if config.Method != "" && config.Method != "GET" {
		scriptPath, err := c.createLuaScript(config.Body, config.Method)
		if err != nil {
			return "", err
		}
		defer os.Remove(scriptPath) // Clean up the temporary script
		args = append(args, "-s", scriptPath)
	}

	args = append(args, config.URL)

	cmd := exec.CommandContext(config.Ctx, "wrk", args...)

	if config.Debug {
		c.output.WriteLine("\nExecuting command:")
		c.output.WriteLine(fmt.Sprintf("wrk %s", strings.Join(args, " ")))
		c.output.WriteLine("---")
	}

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		if config.Ctx.Err() != nil {
			return "", fmt.Errorf("test cancelled")
		}
		// Log the command output for debugging
		c.output.WriteLine(fmt.Sprintf("\nCommand failed with error: %v", err))
		c.output.WriteLine("Command output:")
		c.output.WriteLine(string(outputBytes))
		return "", fmt.Errorf("failed to run wrk: %v", err)
	}

	if config.Debug {
		c.output.WriteLine("\nRaw wrk output:")
		for _, line := range strings.Split(string(outputBytes), "\n") {
			c.output.WriteLine(line)
		}
		c.output.WriteLine("---")
	}

	return string(outputBytes), nil
}
