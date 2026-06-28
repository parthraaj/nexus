package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// svcClient is the IBM SVC REST API client.
// All SVC REST calls are POST-only.
type svcClient struct {
	host    string
	tokens  *tokenManager
	http    *http.Client
}

var svc *svcClient

func initSVCClient() error {
	host := os.Getenv("SVC_HOST")
	username := os.Getenv("SVC_USERNAME")
	password := os.Getenv("SVC_PASSWORD")

	if host == "" || username == "" || password == "" {
		return fmt.Errorf("SVC_HOST, SVC_USERNAME, SVC_PASSWORD env vars are required")
	}

	// Skip TLS verification — SVC uses self-signed certs
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	svc = &svcClient{
		host:   host,
		tokens: newTokenManager(host, username, password, httpClient),
		http:   httpClient,
	}

	// Validate credentials immediately at startup
	_, err := svc.tokens.getToken()
	return err
}

// post sends a POST request to the SVC REST API.
// endpoint: e.g. "lsvdisk" or "lsvdisk/vol-001"
// body: JSON-serializable struct or nil for empty body
// Returns raw JSON response bytes.
func (c *svcClient) post(ctx context.Context, endpoint string, body interface{}) ([]byte, error) {
	token, err := c.tokens.getToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get auth token: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader([]byte("{}"))
	}

	url := fmt.Sprintf("https://%s:7443/rest/v1/%s", c.host, endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("X-Auth-Token", token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SVC request to %q failed: %w", endpoint, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 403 means token expired mid-flight — force refresh and retry once
	if resp.StatusCode == http.StatusForbidden {
		svc.tokens.mu.Lock()
		svc.tokens.token = ""
		svc.tokens.mu.Unlock()

		token, err = svc.tokens.getToken()
		if err != nil {
			return nil, fmt.Errorf("token refresh failed: %w", err)
		}
		req2, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
		req2.Header.Set("X-Auth-Token", token)
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Accept", "application/json")
		resp2, err := c.http.Do(req2)
		if err != nil {
			return nil, fmt.Errorf("retry after token refresh failed: %w", err)
		}
		defer resp2.Body.Close()
		respBody, _ = io.ReadAll(resp2.Body)
		resp = resp2
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("SVC returned HTTP %d for %q: %s", resp.StatusCode, endpoint, string(respBody))
	}

	return respBody, nil
}
