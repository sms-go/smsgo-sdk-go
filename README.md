# smsgo (Go)

[![Go Reference](https://pkg.go.dev/badge/github.com/SMSFy/smsgo-sdk-go.svg)](https://pkg.go.dev/github.com/SMSFy/smsgo-sdk-go)
[![CI](https://github.com/SMSFy/smsgo-sdk-go/actions/workflows/ci.yml/badge.svg)](https://github.com/SMSFy/smsgo-sdk-go/actions/workflows/ci.yml)
[![license](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

Official **Go** SDK for [SMSGo](https://smsgo.com.br) — the simple SMS API for Brazil. Send **OTP/2FA, transactional alerts and campaigns** in a few lines of Go.

- ⚡ **Integrates in minutes** — auth handled for you (no manual token ritual).
- 💸 **No monthly fee** — prepaid credits that don't expire, priced in BRL.
- 🇧🇷 **Brazil-first** — delivery to every carrier, LGPD native.
- 🟢 **Zero dependencies** — standard library only. Fully typed.
- 🎁 **R$ 10 free** on sign-up — test without a card.

> New account and key at **[smsgo.com.br](https://smsgo.com.br)** → panel → **My account → API**.

## Requirements

Go **1.21+** (uses generics for `Paginated[T]`).

## Install

```bash
go get github.com/SMSFy/smsgo-sdk-go@latest
```

```go
import smsgo "github.com/SMSFy/smsgo-sdk-go"
```

The import path is `github.com/SMSFy/smsgo-sdk-go`; the package name is `smsgo`.

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	smsgo "github.com/SMSFy/smsgo-sdk-go"
)

func main() {
	client := smsgo.New(smsgo.Options{APIKey: os.Getenv("SMSGO_KEY")})

	res, err := client.Send(context.Background(), smsgo.SendParams{
		Phone:   "+5511999990000",
		Message: "Olá do SMSGo",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res.ID, res.Status) // -> "a1b2c3...", "queued"
}
```

You pass only the `APIKey`. The SDK exchanges it for a Bearer token (valid 48h),
caches it in memory (guarded by a mutex) and refreshes it automatically when it
expires or the API returns 401.

### Context, everywhere

Unlike the Node SDK, **every network method takes a `context.Context` as its
first argument**. This is an intentional, idiomatic divergence — it lets you set
deadlines and cancel in-flight requests:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
res, err := client.Send(ctx, smsgo.SendParams{ /* ... */ })
```

## Send an OTP (2FA)

```go
n, _ := rand.Int(rand.Reader, big.NewInt(900000))
code := fmt.Sprintf("%06d", n.Int64()+100000)

_, err := client.Send(ctx, smsgo.SendParams{
	Phone:   user.Phone,
	Message: fmt.Sprintf("Seu código SMSGo é %s. Válido por 5 minutos.", code),
})
// store `code` (with a TTL) and compare it on verification
```

## Bulk send

```go
_, err := client.SendBulk(ctx, smsgo.SendBulkParams{
	Messages: []smsgo.BulkMessage{
		{Phone: "+5511999990000", Message: "Oi, Ana!"},
		{Phone: "+5521988887777", Message: "Oi, Bruno!"},
	},
	URLCallback: "https://yourapp.com/webhooks/smsgo", // delivery status (optional)
})
```

## Query sends

```go
page, _ := client.List(ctx, smsgo.ListParams{Page: 1}) // { Meta, Data []SendListItem }
one, _ := client.Get(ctx, "a1b2c3-...")                // detail + Summary{Total,Delivered,Failed,InProgress,Done}

// Track a large send without downloading everything — numbers by bucket, paginated:
failed, _ := client.GetNumbers(ctx, "a1b2c3-...", smsgo.NumbersParams{Status: "failed", Page: 1})
```

## Test mode (sandbox)

Use the **test key** (prefix `test_`, from the panel → My account → API) as
`APIKey`. Nothing changes in your code: sends are **not dispatched and don't
debit balance**, responses mirror production (`Test == true`), and webhooks fire
with the same flag.

```go
sandbox := smsgo.New(smsgo.Options{APIKey: os.Getenv("SMSGO_TEST_KEY")})
r, _ := sandbox.Send(ctx, smsgo.SendParams{Phone: "+5511999990000", Message: "Teste"})
r.Test // true

mode, _ := sandbox.ResolveMode(ctx) // smsgo.ModeTest  (or sandbox.Mode() after the 1st call)
```

## Balance and catalog

```go
bal, _ := client.GetBalance(ctx)     // { Balance: 9.3, Currency: "BRL", Company }
types, _ := client.GetSMSTypes(ctx)  // []SMSTypeItem{ ID, Name, Price, Sale } — ID goes in SMSTypeID
```

## Buy credits (off-session)

Charges a **saved card** without opening the panel (the card is registered in
the panel via Stripe; the API only charges an already-saved one).

```go
plans, _ := client.Billing.Plans(ctx) // tiers by range
cards, _ := client.Billing.Cards(ctx) // last 4 digits

receipt, _ := client.Billing.Purchase(ctx, smsgo.PurchaseParams{Quantity: 5000})
receipt.Status // "succeeded" already credited the balance | "processing" confirms via webhook

invoices, _ := client.Billing.Invoices(ctx, smsgo.InvoicesParams{Page: 1})
```

> **Idempotency:** each `Purchase` creates a new charge. On timeout, query
> `Billing.Invoices` before retrying — **do not blindly retry**.

## Automatic recharge + balance alert

Optional fields are pointers so unset values are stripped from the request:

```go
enabled, threshold, qty := true, 5.0, 5000
alertOn, alertAt := true, 15.0
cardID := "<uuid>"

_, err := client.SetAutoRecharge(ctx, smsgo.AutoRechargeUpdate{
	Enabled:        &enabled,
	Threshold:      &threshold, // recharge when balance ≤ R$ 5
	PlanQuantity:   &qty,       // credits per recharge
	CardID:         &cardID,    // required to enable
	AlertEnabled:   &alertOn,
	AlertThreshold: &alertAt,   // e-mail when balance ≤ R$ 15
})
cfg, _ := client.GetAutoRecharge(ctx)
```

## Outbound webhooks (DLR + replies)

```go
url := "https://yourapp.com/webhooks/smsgo"
cfg, _ := client.SetWebhook(ctx, smsgo.WebhookUpdate{URL: &url}) // store cfg.Secret

rotate := true
client.SetWebhook(ctx, smsgo.WebhookUpdate{RotateSecret: &rotate}) // rotate the secret

empty := ""
client.SetWebhook(ctx, smsgo.WebhookUpdate{URL: &empty}) // disable
```

Each request carries `X-SMSGo-Signature: sha256=<hmac>` — the HMAC-SHA256 of the
**raw body** with your `secret`. Always verify it (constant-time):

```go
func handler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body) // the RAW bytes — the signature is over these
	if !smsgo.VerifyWebhookSignature(body, r.Header.Get("X-SMSGo-Signature"), secret) {
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}
	// ... process the DLR / reply payload
}
```

`VerifyWebhookSignature` never panics: a tampered body, wrong secret, or an
empty/truncated signature returns `false`.

## Contacts and lists

```go
list, _ := client.Lists.Create(ctx, smsgo.ListInput{Name: "Clientes VIP"})
contactID, _ := client.Contacts.Create(ctx, smsgo.ContactInput{
	FullName: "Ana Souza",
	Phone:    "+5511999990000",
	Email:    "ana@exemplo.com",
	Lists:    []string{list.ID},
})

client.Contacts.List(ctx, smsgo.ContactsListParams{Page: 1, Search: "ana"}) // { Meta, Data }
client.Contacts.Update(ctx, contactID, smsgo.ContactInput{FullName: "Ana S.", Phone: "+5511999990000"})
client.Contacts.Delete(ctx, contactID)
```

## Error handling

Every non-2xx response becomes a `*smsgo.Error` with a `Status` and a stable
`Code`. Extract it with `smsgo.AsError`:

```go
_, err := client.Send(ctx, smsgo.SendParams{Phone: "+5511999990000", Message: "Olá"})
if e, ok := smsgo.AsError(err); ok {
	switch e.Code {
	case "insufficient_balance": // 402 — out of balance
	case "rate_limited":         // 429 — too many requests (see e.Details)
	case "validation_error":     // 422 — invalid data (see e.FieldErrors)
	default:
		log.Println(e.Status, e.Code, e.Message)
	}
}
```

On validation failures (422), `e.FieldErrors` carries per-field detail
(`[]FieldError{ Field, Message }`). Transport/network failures have `Status == 0`
and `Code == "network_error"`.

| `Code`                    | HTTP | Meaning                          |
| ------------------------- | ---- | -------------------------------- |
| `validation_error`        | 422  | Invalid request data             |
| `unauthorized`            | 401  | Invalid key/token                |
| `insufficient_balance`    | 402  | Not enough balance               |
| `provider_out_of_stock`   | 409  | Provider stock unavailable       |
| `rate_limited`            | 429  | Rate limit reached               |
| `card_declined`           | 402  | Card declined on purchase        |
| `authentication_required` | 402  | Card needs authentication (SCA)  |
| `card_required`           | 400  | No chargeable card               |
| `payment_unavailable`     | 503  | Payment gateway unavailable      |
| `network_error`           | 0    | Transport failure (no response)  |

(The API-driven codes above come straight from the response body; the SDK maps
unknown statuses to `http_<status>`.)

## API reference

### `smsgo.New(opts smsgo.Options) *smsgo.Client`

| Field        | Type           | Default                    | Description                        |
| ------------ | -------------- | -------------------------- | ---------------------------------- |
| `APIKey`     | `string`       | —                          | **Required.** Your SMSGo-key.      |
| `BaseURL`    | `string`       | `https://api.smsgo.com.br` | Only change if SMSGo tells you to. |
| `HTTPClient` | `*http.Client` | `http.DefaultClient`       | Inject a custom client/transport.  |

`New` never panics. If `APIKey` is empty, methods return a `*smsgo.Error`
(`network_error`, "apiKey is required") on first use.

### Methods

**SMS**

- `Send(ctx, SendParams) (*SendResult, error)` — Fields: `Phone`, `Message`, `Schedule?` (ISO-8601), `Reference?`, `From?`, `SMSTypeID?`.
- `SendBulk(ctx, SendBulkParams) (*SendResult, error)` — up to 5000 messages.
- `List(ctx, ListParams) (*Paginated[SendListItem], error)`.
- `Get(ctx, id) (*SendDetail, error)` — with `Summary`.
- `GetNumbers(ctx, id, NumbersParams) (*Paginated[SendNumberItem], error)`.
- `GetSMSTypes(ctx) ([]SMSTypeItem, error)`.

**Account**

- `GetBalance(ctx) (*Balance, error)`.
- `GetAutoRecharge(ctx)` / `SetAutoRecharge(ctx, AutoRechargeUpdate) (*AutoRechargeConfig, error)`.
- `GetWebhook(ctx)` / `SetWebhook(ctx, WebhookUpdate) (*WebhookConfig, error)`.
- `Mode() AuthMode` / `ResolveMode(ctx) (AuthMode, error)`.

**Billing** (`client.Billing`)

- `Plans(ctx) ([]Plan, error)` · `Cards(ctx) ([]Card, error)` · `Invoices(ctx, InvoicesParams) (*Paginated[InvoiceItem], error)`.
- `Purchase(ctx, PurchaseParams) (*PurchaseResult, error)` — off-session, **not idempotent**.

**Contacts** (`client.Contacts`) and **Lists** (`client.Lists`)

- `List` · `Create` · `Get` · `Update` · `Delete`.

**Webhook helper (top-level)**

- `VerifyWebhookSignature(body []byte, signatureHeader, secret string) bool`.

## Examples

Runnable programs under [`examples/`](./examples):

```bash
SMSGO_KEY=yourkey go run ./examples/send-otp +5511999990000
```

- [`send-sms`](./examples/send-sms) — simple send
- [`send-otp`](./examples/send-otp) — OTP/2FA code
- [`check-status`](./examples/check-status) — bulk send + status query
- [`check-balance`](./examples/check-balance) — balance + SMS-type catalog
- [`buy-credits`](./examples/buy-credits) — off-session purchase + auto-recharge
- [`configure-webhook`](./examples/configure-webhook) — configure the outbound webhook
- [`receive-dlr-webhook`](./examples/receive-dlr-webhook) — receive + verify delivery callbacks

Runnable `Example` functions also live in [`example_test.go`](./example_test.go)
and render on [pkg.go.dev](https://pkg.go.dev/github.com/SMSFy/smsgo-sdk-go).

## Migrating from TotalVoice / Twilio?

SMSGo focuses on **simple DX and BRL pricing**. No sender registration to start,
no dollar billing, credits that don't expire. Full API docs:
**[smspulse.apidog.io](https://smspulse.apidog.io/)**.

## License

MIT © SMSGo
