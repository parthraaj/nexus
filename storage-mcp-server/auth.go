package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// tokenManager handles SVC JWT token lifecycle.
// Token is fetched once, cached, and refreshed 5 minutes before expiry.
type tokenManager struct {
	mu       sync.Mutex
	token    string
	expiry   time.Time
	host     string
	username string
	password string
	client   *http.Client
}

func newTokenManager(host, username, password string, client *http.Client) *tokenManager {
	return &tokenManager{
		host:     host,
		username: username,
		password: password,
		client:   client,
	}
}

// getToken returns a valid token, refreshing if needed.
func (tm *tokenManager) getToken() (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Refresh if no token or expiring within 5 minutes
	if tm.token == "" || time.Now().After(tm.expiry.Add(-5*time.Minute)) {
		if err := tm.fetchToken(); err != nil {
			return "", err
		}
	}
	return tm.token, nil
}

// fetchToken authenticates and stores the new token + expiry.
// Must be called with mu held.
func (tm *tokenManager) fetchToken() error {
	url := fmt.Sprintf("https://%s:7443/rest/v1/auth", tm.host)
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("auth request build failed: %w", err)
	}
	req.Header.Set("X-Auth-Username", tm.username)
	req.Header.Set("X-Auth-Password", tm.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := tm.client.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("auth response read failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("auth response parse failed: %w", err)
	}
	if result.Token == "" {
		return fmt.Errorf("auth returned empty token")
	}

	// Decode JWT payload (middle segment) to extract exp
	expiry, err := extractJWTExpiry(result.Token)
	if err != nil {
		// Fallback: assume 55 minutes if decode fails
		expiry = time.Now().Add(55 * time.Minute)
	}

	tm.token = result.Token
	tm.expiry = expiry
	return nil
}

// extractJWTExpiry decodes the JWT payload and extracts the exp field.
// No external JWT library needed — payload is just base64-encoded JSON.
func extractJWTExpiry(token string) (time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return time.Time{}, fmt.Errorf("invalid JWT format")
	}

	// JWT uses base64url encoding without padding
	payload := parts[1]
	// Add padding if needed
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		return time.Time{}, fmt.Errorf("JWT payload decode failed: %w", err)
	}

	var claims struct {
		Exp int64 `json:"exp"`
	}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return time.Time{}, fmt.Errorf("JWT claims parse failed: %w", err)
	}

	return time.Unix(claims.Exp, 0), nil
}
