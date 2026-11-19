package senviok

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
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

func TestUploadThenSendAttachment(t *testing.T) {
	var uploadSeen, sendSeen bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer svk_test_abc" {
			t.Errorf("unexpected auth header: %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/emails/attachments" {
			if err := r.ParseMultipartForm(1024); err != nil {
				t.Errorf("parse multipart form: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			file, header, err := r.FormFile("file")
			if err != nil {
				t.Errorf("read file field: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			defer file.Close()
			content, _ := io.ReadAll(file)
			if header.Filename != "invoice.pdf" || string(content) != "pdf" {
				t.Errorf("unexpected uploaded file %q: %q", header.Filename, content)
			}
			uploadSeen = true
			_, _ = w.Write([]byte(`{"id":"att_1","filename":"invoice.pdf","contentType":"application/pdf","sizeBytes":3,"bucket":"test","key":"test/att_1"}`))
			return
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode send body: %v", err)
		}
		ids, ok := payload["attachmentIds"].([]any)
		if r.URL.Path != "/v1/emails" || !ok || len(ids) != 1 || ids[0] != "att_1" {
			t.Errorf("unexpected send request: %s %v", r.URL.Path, payload)
		}
		sendSeen = true
		_, _ = w.Write([]byte(`{"id":"msg_1"}`))
	}))
	defer srv.Close()

	client := New("svk_test_abc", WithBaseURL(srv.URL))
	ctx := context.Background()
	uploaded, err := client.Emails.UploadAttachment(ctx, "invoice.pdf", []byte("pdf"), "application/pdf")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Emails.Send(ctx, EmailOptions{
		From: "billing@example.com", To: "customer@example.com", Subject: "Invoice",
		AttachmentIDs: []string{uploaded.ID},
	}); err != nil {
		t.Fatal(err)
	}
	if !uploadSeen || !sendSeen {
		t.Fatalf("expected upload and send requests, got upload=%v send=%v", uploadSeen, sendSeen)
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
