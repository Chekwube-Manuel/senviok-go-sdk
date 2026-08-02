// Package senviok is the official Go client for the Senviok messaging API
// (email, SMS and WhatsApp, plus the resources around them). It mirrors the
// official JavaScript/Python SDKs' public facade 1:1.
//
// The full HTTP surface is implemented with the Go standard library only — no
// external dependencies.
package senviok

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// DefaultBaseURL is the production API endpoint.
	DefaultBaseURL = "https://api.senviok.live"
	// LocalBaseURL is used for local development / testing.
	LocalBaseURL = "http://localhost:5033"
)

// Client is a Senviok API client. Create one with New, then call the resource
// fields (Emails, SMS, WhatsApp, Messages, Webhooks, Templates, Domains,
// Audiences, Contacts, Suppressions, APIKeys).
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client

	Emails       *EmailsResource
	SMS          *SMSResource
	WhatsApp     *WhatsAppResource
	Messages     *MessagesResource
	Webhooks     *WebhooksResource
	Templates    *TemplatesResource
	Domains      *DomainsResource
	Audiences    *AudiencesResource
	Contacts     *ContactsResource
	Suppressions *SuppressionsResource
	APIKeys      *APIKeysResource
}

// New returns a Senviok client. apiKey is required (svk_live_* / svk_test_*).
//
//	client := senviok.New("svk_live_xxx")
//	client := senviok.New("svk_live_xxx", senviok.WithBaseURL("http://localhost:5033"))
func New(apiKey string, opts ...Option) *Client {
	if apiKey == "" {
		panic("senviok: apiKey is required")
	}
	c := &Client{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(DefaultBaseURL, "/"),
		http:    &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.Emails = &EmailsResource{client: c}
	c.SMS = &SMSResource{client: c}
	c.WhatsApp = &WhatsAppResource{client: c}
	c.Messages = &MessagesResource{client: c}
	c.Webhooks = &WebhooksResource{client: c}
	c.Templates = &TemplatesResource{client: c}
	c.Domains = &DomainsResource{client: c}
	c.Audiences = &AudiencesResource{client: c}
	c.Contacts = &ContactsResource{client: c}
	c.Suppressions = &SuppressionsResource{client: c}
	c.APIKeys = &APIKeysResource{client: c}
	return c
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API base URL (e.g. a local dev instance).
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient overrides the default *http.Client (timeout, transport, etc.).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.http = h }
}

// CloseIdleConnections closes the client's idle HTTP connections.
func (c *Client) CloseIdleConnections() {
	if tr, ok := c.http.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
}

// Do is the shared request path. It applies the auth header, sends the request,
// and normalizes non-2xx responses into typed errors.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any) ([]byte, error) {
	u := c.baseURL + "/v1" + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("senviok: marshal request: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, fmt.Errorf("senviok: build request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	c.applyAuth(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("senviok: read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, responseError(resp.StatusCode, raw)
	}
	return raw, nil
}

func (c *Client) applyAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
}

// responseError maps a non-2xx status into a typed error. raw may be empty.
func responseError(status int, raw []byte) error {
	var payload struct {
		Message string `json:"Message"`
	}
	_ = json.Unmarshal(raw, &payload)
	msg := payload.Message
	if msg == "" && len(raw) > 0 {
		msg = string(raw)
	}
	switch status {
	case http.StatusUnauthorized, http.StatusForbidden:
		return &AuthenticationError{Message: msg, Code: status}
	case http.StatusTooManyRequests:
		return &RateLimitError{Message: msg, Code: status}
	default:
		return &APIError{Message: msg, Code: status, Body: raw}
	}
}

// asBytes helper for webhook signature verification. Kept as a free function to
// mirror VerifySignature on the Webhooks resource.
