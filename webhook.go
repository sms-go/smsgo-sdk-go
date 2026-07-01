package smsgo

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
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
