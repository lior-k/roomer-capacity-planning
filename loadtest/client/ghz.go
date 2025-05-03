package client

import (
	"fmt"

	"cursor-roomer/loadtest/types"
)

// GHZClient implements LoadTestClient using the ghz command
type GHZClient struct {
	output types.OutputHandler
}

func NewGHZClient(output types.OutputHandler) types.LoadTestClient {
	return &GHZClient{output: output}
}

func (c *GHZClient) Name() string {
	return "ghz"
}

func (c *GHZClient) RunTest(config types.LoadTestConfig) (string, error) {
	// TODO: Implement GHZ client
	return "", fmt.Errorf("ghz client not implemented yet")
} 