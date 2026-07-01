// Package smsgo is the official Go SDK for [SMSGo] — the simple SMS API for Brazil.
//
// It handles the two-step authentication (SMSGo-key -> 48h Bearer token)
// transparently: you only pass an API key. The token is fetched on demand,
// cached in memory and renewed automatically when it expires or the API
// returns HTTP 401.
//
// The SDK covers the whole public v1 API: sending SMS, querying sends, the SMS
// type catalog, account balance, billing (off-session credit purchase),
// automatic recharge, outbound webhooks, contacts and lists.
//
// # Quick start
//
//	client := smsgo.New(smsgo.Options{APIKey: os.Getenv("SMSGO_KEY")})
//
//	res, err := client.Send(context.Background(), smsgo.SendParams{
//		Phone:   "+5511999990000",
//		Message: "Olá do SMSGo",
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(res.ID, res.Status)
//
// # Context
//
// Unlike the Node SDK, every network method takes a [context.Context] as its
// first argument. This is an intentional, idiomatic divergence: it lets you set
// deadlines and cancel in-flight requests.
//
// # Test mode (sandbox)
//
// A key that starts with "test_" transparently selects sandbox mode: sends are
// not dispatched and do not debit balance, responses mirror production (with
// Test == true), and webhooks fire with the same flag. The detected mode is
// exposed via [Client.Mode] and [Client.ResolveMode].
//
// # Errors
//
// Every non-2xx response is returned as a [*Error] carrying a stable Code and
// the HTTP Status. Use [AsError] to inspect it:
//
//	if e, ok := smsgo.AsError(err); ok && e.Code == "insufficient_balance" {
//		// out of balance
//	}
//
// # Webhooks
//
// Outbound webhooks are signed with HMAC-SHA256 over the raw request body.
// Verify the X-SMSGo-Signature header with [VerifyWebhookSignature] before
// trusting a payload.
//
// [SMSGo]: https://smsgo.com.br
package smsgo
