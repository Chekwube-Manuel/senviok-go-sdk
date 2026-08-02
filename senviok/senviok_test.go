package senviok

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestVerifyWebhookSignature(t *testing.T) {
	body := []byte(`{"event":"message.created"}`)
	secret := "whsec_test"
	got := VerifyWebhookSignature(body, hmacSHA256(body, secret), secret)
	if !got {
		t.Fatal("expected signature to verify")
	}
	if VerifyWebhookSignature(body, "deadbeef", secret) {
		t.Fatal("expected wrong signature to fail")
	}
	if VerifyWebhookSignature(body, "", secret) {
		t.Fatal("expected empty signature to fail")
	}
}

// hmacSHA256 recomputes the HMAC-SHA256 hex for test fixtures.
func hmacSHA256(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// TestEmailOptionsSerialization pins the exact lowercase-camelCase wire format
// the Python SDK emits (to_camel), so the Go SDK stays 1:1.
func TestEmailOptionsSerialization(t *testing.T) {
	opts := EmailOptions{
		From:    "hi@example.com",
		Subject: "Hello",
		To:      "user@example.com",
		HTML:    "<p>Hi</p>",
	}
	b, err := json.Marshal(opts)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"from":"hi@example.com"`) {
		t.Errorf("expected lowercase 'from' key, got %s", s)
	}
	if strings.Contains(s, "addUnsubscribeFooter") {
		t.Errorf("addUnsubscribeFooter should be omitted when nil, got %s", s)
	}
}

func TestClientAuthHeader(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"m1"}`))
	}))
	defer srv.Close()

	c := New("svk_test_abc", WithBaseURL(srv.URL))
	resp, err := c.SMS.Send(t.Context(), "555", "Senviok", "Hi")
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != "m1" {
		t.Fatalf("expected m1, got %s", resp.ID)
	}
	if gotAuth != "Bearer svk_test_abc" {
		t.Fatalf("expected Authorization header, got %q", gotAuth)
	}
}

func TestClientRateLimitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"Message":"Too many requests"}`))
	}))
	defer srv.Close()

	c := New("svk_test_abc", WithBaseURL(srv.URL))
	_, err := c.SMS.Send(context.Background(), "1", "2", "3")
	if err == nil {
		t.Fatal("expected error")
	}
	var rl *RateLimitError
	if !errors.As(err, &rl) {
		t.Fatalf("expected *RateLimitError, got %T: %v", err, err)
	}
}

func TestMessagesListQuery(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":[{"id":"l1","channel":"whatsapp","status":"sent"}]}`))
	}))
	defer srv.Close()

	c := New("svk_test_abc", WithBaseURL(srv.URL))
	logs, err := c.Messages.List(context.Background(), MessageListParams{Channel: "whatsapp"})
	if err != nil {
		t.Fatal(err)
	}
	if len(logs) != 1 || logs[0].ID != "l1" {
		t.Fatalf("unexpected logs: %+v", logs)
	}
	for _, want := range []string{"take=20", "sortOrder=desc", "channel=whatsapp"} {
		if !strings.Contains(gotQuery, want) {
			t.Errorf("query missing %q: %s", want, gotQuery)
		}
	}
}
