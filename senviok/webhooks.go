package senviok

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// VerifyWebhookSignature is a constant-time HMAC-SHA256 check of an incoming
// webhook. Pass the raw request body (exactly as sent), the signature received
// (commonly the X-Senviok-Signature header) and the webhook's secret.
//
// It mirrors the Python SDK's WebhooksResource.verify_signature.
func VerifyWebhookSignature(rawBody []byte, signature, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	computed := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(computed), []byte(signature))
}
