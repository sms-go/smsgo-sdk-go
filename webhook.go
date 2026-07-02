package smsgo

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"time"
)

// VerifyWebhookSignature reports whether signatureHeader is a valid signature
// for body given secret.
//
// The expected value is "sha256=" followed by the lowercase hex HMAC-SHA256 of
// the raw body keyed with secret. The comparison is constant-time. It never
// panics: a tampered body, wrong secret, or empty/truncated signature simply
// returns false.
//
// Pass the raw request body bytes (exactly as received) and the value of the
// X-SMSGo-Signature header.
func VerifyWebhookSignature(body []byte, signatureHeader, secret string) bool {
	if signatureHeader == "" || secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	// Guard length first so ConstantTimeCompare gets equal-length inputs.
	if len(expected) != len(signatureHeader) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signatureHeader)) == 1
}

// VerifyWebhookSignatureWithFreshness reports whether the signature is valid AND
// (anti-replay) the body's sentAt is within toleranceSeconds of now. The
// signature is checked exactly like VerifyWebhookSignature; then, if
// toleranceSeconds > 0, the body is parsed and a stale or unparsable sentAt
// makes it return false. Pass toleranceSeconds <= 0 to skip the freshness check
// (equivalent to VerifyWebhookSignature). Deduplicating on the body's id field
// for idempotency remains the receiver's responsibility. Never panics.
func VerifyWebhookSignatureWithFreshness(body []byte, signatureHeader, secret string, toleranceSeconds int) bool {
	if !VerifyWebhookSignature(body, signatureHeader, secret) {
		return false
	}
	if toleranceSeconds <= 0 {
		return true
	}

	var payload struct {
		SentAt string `json:"sentAt"`
	}
	if err := json.Unmarshal(body, &payload); err != nil || payload.SentAt == "" {
		return false
	}
	sent, err := time.Parse(time.RFC3339, payload.SentAt)
	if err != nil {
		return false
	}
	diff := time.Since(sent)
	if diff < 0 {
		diff = -diff
	}
	return diff <= time.Duration(toleranceSeconds)*time.Second
}
