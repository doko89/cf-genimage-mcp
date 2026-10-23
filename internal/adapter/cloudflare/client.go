// Package cloudflare adapts Cloudflare Workers AI to the domain.ImageGenerator port.
package cloudflare

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cf-genimage/internal/domain"
)

// DefaultBaseURL is the Cloudflare API root.
const DefaultBaseURL = "https://api.cloudflare.com/client/v4"

const (
	defaultTimeout = 120 * time.Second
	maxBodyBytes   = 32 << 20
)

// Client calls the Workers AI text-to-image endpoint.
type Client struct {
	accountID string
	apiToken  string
	baseURL   string
	http      *http.Client
}

// Option customizes a Client.
type Option func(*Client)

// WithBaseURL overrides the API root (useful for tests).
func WithBaseURL(url string) Option {
	return func(c *Client) { c.baseURL = url }
}

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.http = h }
}

// New builds a Cloudflare client. accountID and apiToken are required.
func New(accountID, apiToken string, opts ...Option) (*Client, error) {
	if accountID == "" || apiToken == "" {
		return nil, fmt.Errorf("cloudflare: accountID and apiToken must not be empty")
	}
	c := &Client{
		accountID: accountID,
		apiToken:  apiToken,
		baseURL:   DefaultBaseURL,
		http:      &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

type generatePayload struct {
	Prompt string `json:"prompt"`
	Steps  int    `json:"steps"`
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type generateResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Image string `json:"image"`
	} `json:"result"`
	Errors []apiError `json:"errors"`
}

// Generate implements domain.ImageGenerator using the hardcoded model and step count.
func (c *Client) Generate(ctx context.Context, prompt string) (domain.Image, error) {
	body, err := json.Marshal(generatePayload{Prompt: prompt, Steps: domain.Steps})
	if err != nil {
		return domain.Image{}, fmt.Errorf("cloudflare: encode request: %w", err)
	}
	url := c.baseURL + "/accounts/" + c.accountID + "/ai/run/" + domain.ModelID
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return domain.Image{}, fmt.Errorf("cloudflare: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return domain.Image{}, fmt.Errorf("cloudflare: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBodyBytes))
	if err != nil {
		return domain.Image{}, fmt.Errorf("cloudflare: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return domain.Image{}, fmt.Errorf("cloudflare: HTTP %d: %s", resp.StatusCode, truncate(string(data), 300))
	}
	var out generateResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return domain.Image{}, fmt.Errorf("cloudflare: decode response: %w", err)
	}
	if !out.Success {
		return domain.Image{}, fmt.Errorf("cloudflare: API error: %s", formatErrors(out.Errors))
	}
	if out.Result.Image == "" {
		return domain.Image{}, fmt.Errorf("cloudflare: response has no result.image")
	}
	raw, err := base64.StdEncoding.DecodeString(out.Result.Image)
	if err != nil {
		return domain.Image{}, fmt.Errorf("cloudflare: decode image: %w", err)
	}
	return domain.Image{Data: raw, MIMEType: "image/jpeg"}, nil
}

func formatErrors(errs []apiError) string {
	if len(errs) == 0 {
		return "unknown error"
	}
	msgs := make([]string, 0, len(errs))
	for _, e := range errs {
		msgs = append(msgs, fmt.Sprintf("[%d] %s", e.Code, e.Message))
	}
	return strings.Join(msgs, "; ")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
