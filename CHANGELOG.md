# Changelog

All notable changes to this package are documented here.
The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/)
and versioning follows [SemVer](https://semver.org/).

## [0.3.0] - 2026-07-01

### Added

- First release of the official **Go** SDK for SMSGo, at parity with the public
  `v1` API and the Node SDK (`@orynlabs/smsgo`).
- Transparent two-step auth (SMSGo-key → 48h Bearer token) with an in-memory,
  mutex-guarded token cache; the token is refreshed early (47h) and refreshed
  once + retried once on HTTP 401.
- `Client` with:
  - SMS: `Send`, `SendBulk`, `List`, `Get`, `GetNumbers`, `GetSMSTypes`.
  - Account: `GetBalance`, `GetAutoRecharge`, `SetAutoRecharge`, `GetWebhook`, `SetWebhook`.
  - Mode: `Mode`, `ResolveMode` — `live`/`test` from a `test_` key.
  - Namespaces `Client.Billing` (`Plans`, `Cards`, `Invoices`, `Purchase`),
    `Client.Contacts` and `Client.Lists` (`List`/`Create`/`Get`/`Update`/`Delete`).
- Single error type `smsgo.Error` (`Status`, `Code`, `Message`, `Details`,
  `FieldErrors`) with a stable code map, plus the `AsError` helper.
- `VerifyWebhookSignature` — constant-time HMAC-SHA256 verification of the
  `X-SMSGo-Signature` header over the raw body.
- Zero runtime dependencies (standard library only). Requires Go 1.21+.

### Divergence from the Node SDK

- Every network method takes a `context.Context` as its first argument — the
  idiomatic Go way to carry deadlines and cancellation.
