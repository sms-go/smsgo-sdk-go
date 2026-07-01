package smsgo

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// defaultBaseURL is the public API base.
const defaultBaseURL = "https://api.smsgo.com.br"

// tokenTTL — the Bearer token is valid for 48h; renew early at 47h.
const tokenTTL = 47 * time.Hour

// Options configures a [Client].
type Options struct {
	// APIKey is the permanent account key (panel -> My account -> API). A key
	// starting with "test_" transparently selects sandbox mode. Required.
	APIKey string
	// BaseURL overrides the API base. Default: https://api.smsgo.com.br.
	BaseURL string
	// HTTPClient overrides the HTTP client. Default: http.DefaultClient.
	HTTPClient *http.Client
}

// Client is the SMSGo API client. Create one with [New]. It is safe for
// concurrent use by multiple goroutines.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client

	mu             sync.Mutex
	token          string
	tokenExpiresAt time.Time
	authMode       AuthMode
	modeKnown      bool

	// Contacts is the contacts namespace (CRUD).
	Contacts *ContactsResource
	// Lists is the lists namespace (CRUD).
	Lists *ListsResource
	// Billing is the billing namespace (plans, cards, invoices, purchase).
	Billing *BillingResource
}

// New creates a [Client]. It never panics; if APIKey is empty every network
// method returns an error ("apiKey is required") on first use.
func New(opts Options) *Client {
	base := opts.BaseURL
	if base == "" {
		base = defaultBaseURL
	}
	base = strings.TrimRight(base, "/")

	hc := opts.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}

	c := &Client{
		apiKey:     opts.APIKey,
		baseURL:    base,
		httpClient: hc,
	}
	c.Contacts = &ContactsResource{do: c.do}
	c.Lists = &ListsResource{do: c.do}
	c.Billing = &BillingResource{do: c.do}
	return c
}

// Mode returns the mode (live or test) of the current key, known after the
// first authenticated call. It returns an empty string before then; use
// [Client.ResolveMode] to force resolution.
func (c *Client) Mode() AuthMode {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.modeKnown {
		return ""
	}
	return c.authMode
}

// ResolveMode ensures a token exists and returns the key's mode (live or test).
func (c *Client) ResolveMode(ctx context.Context) (AuthMode, error) {
	if _, err := c.ensureToken(ctx, false); err != nil {
		return "", err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.modeKnown {
		return ModeLive, nil
	}
	return c.authMode, nil
}

/* --- SMS ---------------------------------------------------------------- */

// Send sends a single SMS.
func (c *Client) Send(ctx context.Context, params SendParams) (*SendResult, error) {
	var out SendResult
	if err := c.do(ctx, http.MethodPost, "/v1/sms/send/single", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SendBulk sends several messages in a single transaction (up to 5000).
func (c *Client) SendBulk(ctx context.Context, params SendBulkParams) (*SendResult, error) {
	var out SendResult
	if err := c.do(ctx, http.MethodPost, "/v1/sms/send/multiple", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// List lists the account's sends (paginated).
func (c *Client) List(ctx context.Context, params ListParams) (*Paginated[SendListItem], error) {
	page := params.Page
	if page == 0 {
		page = 1
	}
	q := buildQuery(map[string]string{"page": strconv.Itoa(page)})
	var out Paginated[SendListItem]
	if err := c.do(ctx, http.MethodGet, "/v1/sms/list"+q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get details a send by its UUID (includes a tracking summary).
func (c *Client) Get(ctx context.Context, id string) (*SendDetail, error) {
	var out SendDetail
	path := "/v1/sms/" + url.PathEscape(id) + "/show"
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetNumbers returns a send's numbers, paginated and filterable by status bucket.
func (c *Client) GetNumbers(ctx context.Context, id string, params NumbersParams) (*Paginated[SendNumberItem], error) {
	q := buildQuery(map[string]string{
		"status": params.Status,
		"page":   intOrEmpty(params.Page),
	})
	path := "/v1/sms/" + url.PathEscape(id) + "/numbers" + q
	var out Paginated[SendNumberItem]
	if err := c.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetSMSTypes returns the catalog of active SMS types (id = SMSTypeID).
func (c *Client) GetSMSTypes(ctx context.Context) ([]SMSTypeItem, error) {
	var wrapper struct {
		Data []SMSTypeItem `json:"data"`
	}
	if err := c.do(ctx, http.MethodGet, "/v1/sms-types", nil, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Data, nil
}

/* --- Account ------------------------------------------------------------ */

// GetBalance returns the monetary balance (BRL) plus basic account data.
func (c *Client) GetBalance(ctx context.Context) (*Balance, error) {
	var out Balance
	if err := c.do(ctx, http.MethodGet, "/v1/account/balance", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAutoRecharge reads the automatic-recharge + low-balance alert config.
func (c *Client) GetAutoRecharge(ctx context.Context) (*AutoRechargeConfig, error) {
	var out AutoRechargeConfig
	if err := c.do(ctx, http.MethodGet, "/v1/account/auto-recharge", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetAutoRecharge updates the automatic-recharge + alert config. To ENABLE the
// recharge, CardID and PlanQuantity are required.
func (c *Client) SetAutoRecharge(ctx context.Context, params AutoRechargeUpdate) (*AutoRechargeConfig, error) {
	var out AutoRechargeConfig
	if err := c.do(ctx, http.MethodPut, "/v1/account/auto-recharge", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetWebhook reads the outbound webhook URL and secret.
func (c *Client) GetWebhook(ctx context.Context) (*WebhookConfig, error) {
	var out WebhookConfig
	if err := c.do(ctx, http.MethodGet, "/v1/account/webhook", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// SetWebhook sets the outbound webhook (DLR + replies). An empty URL disables
// it; RotateSecret rotates the signing secret.
func (c *Client) SetWebhook(ctx context.Context, params WebhookUpdate) (*WebhookConfig, error) {
	var out WebhookConfig
	if err := c.do(ctx, http.MethodPut, "/v1/account/webhook", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

/* --- Internal auth ------------------------------------------------------ */

// ensureToken exchanges the SMSGo-key for a Bearer token, caching it until it
// expires. When forceRefresh is true the cached token is ignored.
func (c *Client) ensureToken(ctx context.Context, forceRefresh bool) (string, error) {
	if c.apiKey == "" {
		return "", &Error{Status: 0, Code: "network_error", Message: "apiKey is required"}
	}

	c.mu.Lock()
	if !forceRefresh && c.token != "" && time.Now().Before(c.tokenExpiresAt) {
		tok := c.token
		c.mu.Unlock()
		return tok, nil
	}
	c.mu.Unlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/v1/auth/token", nil)
	if err != nil {
		return "", &Error{Status: 0, Code: "network_error", Message: err.Error()}
	}
	req.Header.Set("SMSGo-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", &Error{Status: 0, Code: "network_error", Message: err.Error()}
	}
	defer res.Body.Close()

	body := parseBody(res.Body)

	obj, _ := body.(map[string]any)
	token, _ := obj["token"].(string)
	if res.StatusCode < 200 || res.StatusCode >= 300 || token == "" {
		return "", toError(res.StatusCode, body, "Falha ao autenticar a SMSGo-key.")
	}

	mode := ModeLive
	if m, _ := obj["mode"].(string); m == "test" {
		mode = ModeTest
	}

	c.mu.Lock()
	c.token = token
	c.authMode = mode
	c.modeKnown = true
	c.tokenExpiresAt = time.Now().Add(tokenTTL)
	c.mu.Unlock()

	return token, nil
}

// do performs an authenticated request. When out is non-nil the (JSON) response
// body is decoded into it. On HTTP 401 the token is refreshed once and the
// request retried once.
func (c *Client) do(ctx context.Context, method, path string, payload, out any) error {
	return c.request(ctx, method, path, payload, out, false)
}

func (c *Client) request(ctx context.Context, method, path string, payload, out any, isRetry bool) error {
	token, err := c.ensureToken(ctx, false)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	hasBody := payload != nil
	if hasBody {
		raw, err := json.Marshal(payload)
		if err != nil {
			return &Error{Status: 0, Code: "network_error", Message: err.Error()}
		}
		bodyReader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return &Error{Status: 0, Code: "network_error", Message: err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	if hasBody {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return &Error{Status: 0, Code: "network_error", Message: err.Error()}
	}
	defer res.Body.Close()

	// Expired/revoked token: refresh once and retry once.
	if res.StatusCode == http.StatusUnauthorized && !isRetry {
		io.Copy(io.Discard, res.Body)
		res.Body.Close()
		if _, err := c.ensureToken(ctx, true); err != nil {
			return err
		}
		return c.request(ctx, method, path, payload, out, true)
	}

	body := parseBody(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return toError(res.StatusCode, body, "Erro na requisição.")
	}

	if out == nil {
		return nil
	}
	// Re-marshal the already-parsed body into the typed output. On an empty
	// body (parsed as nil) there is nothing to decode.
	if body == nil {
		return nil
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return &Error{Status: 0, Code: "network_error", Message: err.Error()}
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return &Error{Status: 0, Code: "network_error", Message: err.Error()}
	}
	return nil
}

/* --- Helpers ------------------------------------------------------------ */

// parseBody reads the response body: empty -> nil, valid JSON -> decoded value,
// otherwise the raw string.
func parseBody(r io.Reader) any {
	raw, err := io.ReadAll(r)
	if err != nil || len(raw) == 0 {
		return nil
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	return v
}

// toError builds an [*Error] from a non-2xx response body.
func toError(status int, body any, fallback string) *Error {
	code := httpCodeName(status)
	message := fallback
	var fieldErrors []FieldError

	switch b := body.(type) {
	case map[string]any:
		switch cv := b["code"].(type) {
		case string:
			if cv != "" {
				code = cv
			}
		case float64:
			// Legacy controllers may return a numeric code (e.g. 1006). JSON
			// numbers decode to float64; match Node's String(body.code).
			code = strconv.FormatFloat(cv, 'f', -1, 64)
		}
		if m, ok := b["message"].(string); ok && m != "" {
			message = m
		}
		if arr, ok := b["errors"].([]any); ok {
			for _, item := range arr {
				if fe, ok := item.(map[string]any); ok {
					field, _ := fe["field"].(string)
					msg, _ := fe["message"].(string)
					fieldErrors = append(fieldErrors, FieldError{Field: field, Message: msg})
				}
			}
		}
	case string:
		if b != "" {
			message = b
		}
	}

	return &Error{
		Status:      status,
		Code:        code,
		Message:     message,
		Details:     body,
		FieldErrors: fieldErrors,
	}
}

// buildQuery builds a query string (?a=1&b=2), skipping empty values.
func buildQuery(params map[string]string) string {
	usp := url.Values{}
	for k, v := range params {
		if v != "" {
			usp.Set(k, v)
		}
	}
	qs := usp.Encode()
	if qs == "" {
		return ""
	}
	return "?" + qs
}

// intOrEmpty renders a non-zero int as a string, else "".
func intOrEmpty(n int) string {
	if n == 0 {
		return ""
	}
	return strconv.Itoa(n)
}
