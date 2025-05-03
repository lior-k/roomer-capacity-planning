package client

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"cursor-roomer/loadtest/types"
)

// K6Client implements LoadTestClient using the k6 command
type K6Client struct {
	output types.OutputHandler
}

func NewK6Client(output types.OutputHandler) types.LoadTestClient {
	return &K6Client{output: output}
}

func (c *K6Client) Name() string {
	return "k6"
}

func (c *K6Client) createScript(config types.LoadTestConfig) (string, error) {
	// Create a temporary file for the k6 script
	tmpFile, err := os.CreateTemp("", "k6-script-*.js")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	// Don't defer the removal here, we'll do it after k6 has finished running

	// Write the k6 script to the temporary file
	script := fmt.Sprintf(`
import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  vus: %d,
  duration: '%s',
  thresholds: {
    http_req_duration: ['p(90)<1000'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'max', 'p(50)', 'p(75)', 'p(90)', 'p(99)'],
};

export default function() {
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  %s
}
`, config.Goroutines, config.Duration, func() string {
		if config.Method == "" || config.Method == "GET" {
			return fmt.Sprintf("const res = http.get('%s', params);", config.URL)
		}
		if config.Body != "" {
			return fmt.Sprintf("const res = http.%s('%s', %s, params);", strings.ToLower(config.Method), config.URL, config.Body)
		}
		return fmt.Sprintf("const res = http.%s('%s', null, params);", strings.ToLower(config.Method), config.URL)
	}())

	fmt.Printf("Generated k6 script:\n%s\n", script)
	fmt.Printf("Config: %+v\n", config)

	if _, err := tmpFile.WriteString(script); err != nil {
		return "", fmt.Errorf("failed to write k6 script: %v", err)
	}

	// Close the file so k6 can read it
	if err := tmpFile.Close(); err != nil {
		return "", fmt.Errorf("failed to close k6 script: %v", err)
	}

	return tmpFile.Name(), nil
}

func (c *K6Client) RunTest(config types.LoadTestConfig) (string, error) {
	// Check if k6 is installed
	if _, err := exec.LookPath("k6"); err != nil {
		return "", fmt.Errorf("k6 is not installed. Please install it first: %v", err)
	}

	c.output.WriteLine(fmt.Sprintf("Running test with %d virtual users for %v...",
		config.Goroutines, config.Duration))

	scriptPath, err := c.createScript(config)
	if err != nil {
		return "", err
	}
	// Clean up the temporary script after k6 has finished running
	// defer os.Remove(scriptPath)

	cmd := exec.CommandContext(config.Ctx, "k6", "run", scriptPath)

	if config.Debug {
		c.output.WriteLine("\nExecuting command:")
		c.output.WriteLine(fmt.Sprintf("k6 run %s", scriptPath))
		c.output.WriteLine("---")
	}

	// Capture both stdout and stderr separately
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		if config.Ctx.Err() != nil {
			return "", fmt.Errorf("test cancelled")
		}
		// Log the command output for debugging
		c.output.WriteLine(fmt.Sprintf("\nCommand failed with error: %v", err))
		c.output.WriteLine("Command stdout:")
		c.output.WriteLine(stdout.String())
		c.output.WriteLine("Command stderr:")
		c.output.WriteLine(stderr.String())
		return "", fmt.Errorf("failed to run k6: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Always show the output if debug is enabled
	if config.Debug {
		c.output.WriteLine("\nRaw k6 output:")
		c.output.WriteLine(stdout.String())
		if stderr.String() != "" {
			c.output.WriteLine("\nError output:")
			c.output.WriteLine(stderr.String())
		}
		c.output.WriteLine("---")
	}

	return stdout.String(), nil
}
