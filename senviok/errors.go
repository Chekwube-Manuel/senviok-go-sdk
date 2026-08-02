package senviok

import "fmt"

// Error is the base interface every Senviok error satisfies.
type Error interface {
	error
	StatusCode() int
}

// APIError is returned for any non-2xx response that is not otherwise
// classified (4xx/5xx).
type APIError struct {
	Message string
	Code    int
	Body    []byte
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("senviok: %s (HTTP %d)", e.Message, e.Code)
	}
	return fmt.Sprintf("senviok: request failed with HTTP %d", e.Code)
}

func (e *APIError) StatusCode() int { return e.Code }

// AuthenticationError indicates the API key is missing, invalid, or revoked.
type AuthenticationError struct {
	Message string
	Code    int
}

func (e *AuthenticationError) Error() string {
	msg := "senviok: authentication failed"
	if e.Message != "" {
		msg = "senviok: " + e.Message
	}
	return fmt.Sprintf("%s (HTTP %d)", msg, e.Code)
}

func (e *AuthenticationError) StatusCode() int { return e.Code }

// RateLimitError indicates the request was throttled (HTTP 429).
type RateLimitError struct {
	Message string
	Code    int
}

func (e *RateLimitError) Error() string {
	msg := "senviok: rate limit exceeded"
	if e.Message != "" {
		msg = "senviok: " + e.Message
	}
	return fmt.Sprintf("%s (HTTP %d)", msg, e.Code)
}

func (e *RateLimitError) StatusCode() int { return e.Code }
