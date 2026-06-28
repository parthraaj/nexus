package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

type ExecRequest struct {
	Command string            `json:"command"`
	Params  map[string]string `json:"params"`
}

type ExecResponse struct {
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
	IsError bool   `json:"is_error"`
}

// callDaemon sends a command to the node-daemon pod on the target node.
// It resolves the pod IP fresh on every call.
func callDaemon(ctx context.Context, nodeName, command string, params map[string]string) (string, error) {
	podIP, err := getDaemonPodIP(ctx, nodeName)
	if err != nil {
		return "", err
	}

	req := ExecRequest{Command: command, Params: params}
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://%s:9090/exec", podIP)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to call node-daemon on node %q: %w", nodeName, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var execResp ExecResponse
	if err := json.Unmarshal(respBody, &execResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if execResp.IsError {
		return "", fmt.Errorf("node-daemon error: %s", execResp.Error)
	}

	return execResp.Output, nil
}
