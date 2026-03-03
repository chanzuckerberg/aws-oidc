package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	pushEndpoint   = "/loki/api/v1/push"
	initialBackoff = 100 * time.Millisecond
	maxErrorBody   = 1024
)

// Entry is a single log line with a timestamp for the Loki push API.
type Entry struct {
	Timestamp time.Time
	Line      string
}

// Client sends log entries to a Loki server via the push API.
type Client struct {
	baseURL    string
	userAgent  string
	username   string
	password   string
	httpClient *http.Client
	maxRetries int
}

// NewClient creates a new Loki HTTP client. baseURL is the Loki server base (e.g. https://loki.example.com).
// userAgent is sent on each request (e.g. "aws-oidc/1.2.3"); if empty, "aws-oidc" is used.
// username and password are used for Basic Auth when both are non-empty.
func NewClient(baseURL, userAgent, username, password string, timeout time.Duration, maxRetries int) *Client {
	if maxRetries < 0 {
		maxRetries = 0
	}
	if userAgent == "" {
		userAgent = "aws-oidc"
	}
	return &Client{
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		userAgent: userAgent,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: maxRetries,
	}
}

// Push sends log entries to Loki with the given stream labels. Entries are sent in one request.
func (c *Client) Push(ctx context.Context, labels map[string]string, entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	payload, err := c.buildPayload(labels, entries)
	if err != nil {
		return fmt.Errorf("build payload: %w", err)
	}
	return c.sendWithRetry(ctx, payload)
}

type pushRequest struct {
	Streams []stream `json:"streams"`
}

type stream struct {
	Stream map[string]string `json:"stream"`
	Values [][2]string       `json:"values"`
}

func (c *Client) buildPayload(labels map[string]string, entries []Entry) ([]byte, error) {
	values := make([][2]string, 0, len(entries))
	for _, e := range entries {
		ts := strconv.FormatInt(e.Timestamp.UnixNano(), 10)
		values = append(values, [2]string{ts, e.Line})
	}
	req := pushRequest{
		Streams: []stream{{
			Stream: labels,
			Values: values,
		}},
	}
	return json.Marshal(req)
}

func (c *Client) sendWithRetry(ctx context.Context, payload []byte) error {
	var lastErr error
	backoff := initialBackoff
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
			backoff *= 2
		}
		lastErr = c.send(ctx, payload)
		if lastErr == nil {
			return nil
		}
	}
	return fmt.Errorf("push failed after %d retries: %w", c.maxRetries+1, lastErr)
}

func (c *Client) send(ctx context.Context, payload []byte) error {
	url := c.baseURL + pushEndpoint
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.username != "" && c.password != "" {
		req.SetBasicAuth(c.username, c.password)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	limited := io.LimitReader(resp.Body, maxErrorBody)
	body, _ := io.ReadAll(limited)
	return fmt.Errorf("loki returned status %d: %s", resp.StatusCode, string(body))
}
