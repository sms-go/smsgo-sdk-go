package smsgo

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// newTestClient wires a Client to a test server's URL.
func newTestClient(t *testing.T, srv *httptest.Server, apiKey string) *Client {
	t.Helper()
	return New(Options{APIKey: apiKey, BaseURL: srv.URL, HTTPClient: srv.Client()})
}

// tokenHandler serves the /v1/auth/token exchange, returning the given mode.
func tokenHandler(t *testing.T, mode string, calls *int32) func(http.ResponseWriter, *http.Request) {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(calls, 1)
		if r.Header.Get("SMSGo-key") == "" {
			t.Errorf("missing SMSGo-key header on token exchange")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"token": "tok-abc", "mode": mode})
	}
}

func TestTokenExchangeAndMode(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		respMode string
		want     AuthMode
	}{
		{"live key", "live_key", "live", ModeLive},
		{"test key", "test_key", "test", ModeTest},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var tokenCalls int32
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/auth/token", tokenHandler(t, tc.respMode, &tokenCalls))
			srv := httptest.NewServer(mux)
			defer srv.Close()

			c := newTestClient(t, srv, tc.key)
			if c.Mode() != "" {
				t.Fatalf("Mode() before first call = %q, want empty", c.Mode())
			}
			got, err := c.ResolveMode(context.Background())
			if err != nil {
				t.Fatalf("ResolveMode: %v", err)
			}
			if got != tc.want {
				t.Fatalf("ResolveMode = %q, want %q", got, tc.want)
			}
			if c.Mode() != tc.want {
				t.Fatalf("Mode() = %q, want %q", c.Mode(), tc.want)
			}
			// Second call should reuse the cached token.
			if _, err := c.ResolveMode(context.Background()); err != nil {
				t.Fatal(err)
			}
			if tokenCalls != 1 {
				t.Fatalf("token exchanged %d times, want 1 (cache miss)", tokenCalls)
			}
		})
	}
}

func TestSendBodyMapping(t *testing.T) {
	var tokenCalls int32
	var gotBody map[string]any
	var gotAuth, gotContentType string

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/token", tokenHandler(t, "live", &tokenCalls))
	mux.HandleFunc("/v1/sms/send/single", func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		raw, _ := io.ReadAll(r.Body)
		json.Unmarshal(raw, &gotBody)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": "snd-1", "quantity": 1, "status": "queued", "test": true})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv, "test_key")
	res, err := c.Send(context.Background(), SendParams{
		Phone:     "+5511999990000",
		Message:   "Olá",
		SMSTypeID: 7,
	})
	if err != nil {
		t.Fatalf("Send: %v", err)
	}

	if gotAuth != "Bearer tok-abc" {
		t.Errorf("Authorization = %q, want Bearer tok-abc", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	// camelCase -> snake_case mapping and omitempty stripping.
	if gotBody["phone"] != "+5511999990000" || gotBody["message"] != "Olá" {
		t.Errorf("phone/message mismatch: %v", gotBody)
	}
	if gotBody["sms_type_id"].(float64) != 7 {
		t.Errorf("sms_type_id = %v, want 7", gotBody["sms_type_id"])
	}
	if _, present := gotBody["schedule"]; present {
		t.Errorf("empty schedule should be stripped, got: %v", gotBody)
	}
	if _, present := gotBody["from"]; present {
		t.Errorf("empty from should be stripped, got: %v", gotBody)
	}
	if res.ID != "snd-1" || res.Status != "queued" || !res.Test {
		t.Errorf("unexpected result: %+v", res)
	}
}

func TestUnauthorizedRefreshRetry(t *testing.T) {
	var tokenCalls int32
	var sendCalls int32

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/token", tokenHandler(t, "live", &tokenCalls))
	mux.HandleFunc("/v1/sms/send/single", func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&sendCalls, 1)
		if n == 1 {
			// First attempt: token rejected.
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"code": "unauthorized", "message": "expired"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": "snd-2", "quantity": 1, "status": "queued"})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv, "live_key")
	res, err := c.Send(context.Background(), SendParams{Phone: "+5511999990000", Message: "hi"})
	if err != nil {
		t.Fatalf("Send after refresh: %v", err)
	}
	if res.ID != "snd-2" {
		t.Fatalf("result ID = %q, want snd-2", res.ID)
	}
	if sendCalls != 2 {
		t.Fatalf("send attempted %d times, want 2 (retry once)", sendCalls)
	}
	if tokenCalls != 2 {
		t.Fatalf("token exchanged %d times, want 2 (initial + refresh)", tokenCalls)
	}
}

func TestErrorMapping(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		respBody   any
		wantCode   string
		wantMsg    string
		wantFields int
	}{
		{
			name:     "validation error with field errors",
			status:   422,
			respBody: map[string]any{"code": "validation_error", "message": "invalid", "errors": []any{map[string]any{"field": "phone", "message": "required"}}},
			wantCode: "validation_error", wantMsg: "invalid", wantFields: 1,
		},
		{
			name:     "insufficient balance from status only",
			status:   402,
			respBody: map[string]any{"message": "no funds"},
			wantCode: "insufficient_balance", wantMsg: "no funds",
		},
		{
			name:     "unknown status fallback code",
			status:   418,
			respBody: map[string]any{},
			wantCode: "http_418", wantMsg: "Erro na requisição.",
		},
		{
			name:     "string body becomes message",
			status:   500,
			respBody: "boom",
			wantCode: "http_500", wantMsg: "boom",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var tokenCalls int32
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/auth/token", tokenHandler(t, "live", &tokenCalls))
			mux.HandleFunc("/v1/account/balance", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				if s, ok := tc.respBody.(string); ok {
					io.WriteString(w, s)
					return
				}
				json.NewEncoder(w).Encode(tc.respBody)
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()

			c := newTestClient(t, srv, "live_key")
			_, err := c.GetBalance(context.Background())
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			e, ok := AsError(err)
			if !ok {
				t.Fatalf("AsError = false for %T", err)
			}
			if e.Status != tc.status {
				t.Errorf("status = %d, want %d", e.Status, tc.status)
			}
			if e.Code != tc.wantCode {
				t.Errorf("code = %q, want %q", e.Code, tc.wantCode)
			}
			if e.Message != tc.wantMsg {
				t.Errorf("message = %q, want %q", e.Message, tc.wantMsg)
			}
			if len(e.FieldErrors) != tc.wantFields {
				t.Errorf("field errors = %d, want %d", len(e.FieldErrors), tc.wantFields)
			}
		})
	}
}

func TestNetworkErrorAndEmptyKey(t *testing.T) {
	// Empty API key -> network_error on first use, without any HTTP call.
	c := New(Options{APIKey: ""})
	_, err := c.GetBalance(context.Background())
	e, ok := AsError(err)
	if !ok {
		t.Fatalf("AsError = false for %T", err)
	}
	if e.Status != 0 || e.Code != "network_error" {
		t.Fatalf("empty-key error = %+v, want status 0 / network_error", e)
	}
}

func TestListDefaultsPageOne(t *testing.T) {
	var tokenCalls int32
	var gotQuery string
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/token", tokenHandler(t, "live", &tokenCalls))
	mux.HandleFunc("/v1/sms/list", func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"meta": map[string]any{"total": 0}, "data": []any{}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv, "live_key")
	if _, err := c.List(context.Background(), ListParams{}); err != nil {
		t.Fatalf("List: %v", err)
	}
	if gotQuery != "page=1" {
		t.Fatalf("query = %q, want page=1", gotQuery)
	}
}

func TestGetSMSTypesUnwrapsData(t *testing.T) {
	var tokenCalls int32
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/auth/token", tokenHandler(t, "live", &tokenCalls))
	mux.HandleFunc("/v1/sms-types", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": []any{
			map[string]any{"id": 1, "name": "Transactional", "price": 0.08, "sale": nil},
		}})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	c := newTestClient(t, srv, "live_key")
	types, err := c.GetSMSTypes(context.Background())
	if err != nil {
		t.Fatalf("GetSMSTypes: %v", err)
	}
	if len(types) != 1 || types[0].ID != 1 || types[0].Name != "Transactional" {
		t.Fatalf("unexpected types: %+v", types)
	}
	if types[0].Sale != nil {
		t.Fatalf("Sale = %v, want nil", types[0].Sale)
	}
}

func TestVerifyWebhookSignatureGoldenVector(t *testing.T) {
	const (
		secret   = "whsec_3f8a9c2e1b6d4a70f5e2c9b8a1d7e0f4"
		rawBody  = `{"event":"sms.status","data":{"sendId":"7c3e1a90-2b4d-4f6a-8c1e-9d0f2a3b4c5d","phone":"5511999990000","status":"delivered"}}`
		expected = "sha256=986eb0c41355b1c94165c4cb275ce2cc9b175e5f93efe7e2ed4294ba58d330c3"
	)

	if !VerifyWebhookSignature([]byte(rawBody), expected, secret) {
		t.Fatal("golden vector: expected signature to verify true")
	}

	tampered := []struct {
		name string
		body []byte
		sig  string
		sec  string
	}{
		{"flipped byte in body", []byte(rawBody[:10] + "X" + rawBody[11:]), expected, secret},
		{"wrong secret", []byte(rawBody), expected, "whsec_wrong"},
		{"truncated signature", []byte(rawBody), expected[:len(expected)-2], secret},
		{"empty signature", []byte(rawBody), "", secret},
		{"empty secret", []byte(rawBody), expected, ""},
		{"garbage signature", []byte(rawBody), "not-a-signature", secret},
	}
	for _, tc := range tampered {
		t.Run(tc.name, func(t *testing.T) {
			if VerifyWebhookSignature(tc.body, tc.sig, tc.sec) {
				t.Fatalf("expected false for %s", tc.name)
			}
		})
	}
}
