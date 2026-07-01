package smsgo

// AuthMode is the authentication mode of the current API key.
type AuthMode string

const (
	// ModeLive is a production key.
	ModeLive AuthMode = "live"
	// ModeTest is a sandbox (test_) key.
	ModeTest AuthMode = "test"
)

/* -------------------------------------------------------------------------- */
/* Sending                                                                    */
/* -------------------------------------------------------------------------- */

// SendParams are the parameters for [Client.Send].
type SendParams struct {
	// Phone in international E.164 format, e.g. +5511999990000.
	Phone string `json:"phone"`
	// Message text (1–1600 characters; the real limit depends on the provider).
	Message string `json:"message"`
	// Schedule is an optional ISO-8601 timestamp.
	Schedule string `json:"schedule,omitempty"`
	// Reference is your own identifier, echoed back in webhooks (optional).
	Reference string `json:"reference,omitempty"`
	// From is the sender, as supported by the provider (optional).
	From string `json:"from,omitempty"`
	// SMSTypeID selects a pricing tier (optional). See [Client.GetSMSTypes].
	SMSTypeID int `json:"sms_type_id,omitempty"`
}

// BulkMessage is a single message inside [SendBulkParams].
type BulkMessage struct {
	Phone     string `json:"phone"`
	Message   string `json:"message"`
	Schedule  string `json:"schedule,omitempty"`
	Reference string `json:"reference,omitempty"`
	From      string `json:"from,omitempty"`
}

// SendBulkParams are the parameters for [Client.SendBulk].
type SendBulkParams struct {
	// Messages holds up to 5000 messages per request.
	Messages []BulkMessage `json:"messages"`
	// URLCallback receives delivery-status callbacks (optional).
	URLCallback string `json:"urlCallback,omitempty"`
	// FlashSms sends as a flash SMS if the provider supports it (optional). It is
	// a pointer so an explicit false is still transmitted (nil is stripped),
	// mirroring the Node SDK's stripUndefined semantics.
	FlashSms *bool `json:"flashSms,omitempty"`
	// SMSTypeID selects a pricing tier (optional).
	SMSTypeID int `json:"sms_type_id,omitempty"`
}

// SendResult is returned by [Client.Send] and [Client.SendBulk].
type SendResult struct {
	// ID is the send UUID.
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
	// Status is "scheduled" when scheduled, otherwise "queued".
	Status string `json:"status"`
	// Test is true only in sandbox mode.
	Test bool `json:"test,omitempty"`
}

/* -------------------------------------------------------------------------- */
/* Querying sends                                                             */
/* -------------------------------------------------------------------------- */

// PaginationMeta describes a paginated result page.
type PaginationMeta struct {
	Total           int     `json:"total"`
	PerPage         int     `json:"perPage"`
	CurrentPage     int     `json:"currentPage"`
	LastPage        int     `json:"lastPage"`
	FirstPage       int     `json:"firstPage"`
	FirstPageURL    string  `json:"firstPageUrl"`
	LastPageURL     string  `json:"lastPageUrl"`
	NextPageURL     *string `json:"nextPageUrl"`
	PreviousPageURL *string `json:"previousPageUrl"`
}

// Paginated is a generic paginated response.
type Paginated[T any] struct {
	Meta PaginationMeta `json:"meta"`
	Data []T            `json:"data"`
}

// ListParams are the parameters for [Client.List].
type ListParams struct {
	// Page number (defaults to 1 when zero).
	Page int
}

// SendListItem is one row of [Client.List].
type SendListItem struct {
	ID        string  `json:"id"`
	Number    *int    `json:"number"`
	Date      *string `json:"date"`
	Quantity  int     `json:"quantity"`
	FullName  string  `json:"full_name"`
	CreatedAt string  `json:"created_at"`
	Status    string  `json:"status"`
	Type      string  `json:"type"`
}

// SendSummary holds status-bucket counts for a send.
type SendSummary struct {
	Total      int `json:"total"`
	Delivered  int `json:"delivered"`
	Failed     int `json:"failed"`
	InProgress int `json:"inProgress"`
	// Done is true when no number is still in progress.
	Done bool `json:"done"`
}

// SendNumberDetail is a per-number entry inside [SendDetail].
type SendNumberDetail struct {
	ID         string  `json:"id"`
	Characters int     `json:"characters"`
	Code       *string `json:"code"`
	Cost       float64 `json:"cost"`
	Message    string  `json:"message"`
	Phone      string  `json:"phone"`
	Status     string  `json:"status"`
	Template   *string `json:"template"`
	CreatedAt  string  `json:"created_at"`
}

// SendDetail is returned by [Client.Get].
type SendDetail struct {
	ID         string             `json:"id"`
	Quantity   int                `json:"quantity"`
	Characters int                `json:"characters"`
	Date       *string            `json:"date"`
	Total      float64            `json:"total"`
	Cost       float64            `json:"cost"`
	User       string             `json:"user"`
	Status     string             `json:"status"`
	Type       string             `json:"type"`
	Summary    SendSummary        `json:"summary"`
	Phones     []SendNumberDetail `json:"phones"`
}

// SendNumberItem is one row of [Client.GetNumbers].
type SendNumberItem struct {
	ID        string  `json:"id"`
	Phone     string  `json:"phone"`
	Code      *string `json:"code"`
	Status    string  `json:"status"`
	CreatedAt string  `json:"created_at"`
}

// NumbersParams are the parameters for [Client.GetNumbers].
type NumbersParams struct {
	// Status filters by bucket: "delivered", "failed" or "in_progress".
	Status string
	Page   int
}

/* -------------------------------------------------------------------------- */
/* Account and catalog                                                        */
/* -------------------------------------------------------------------------- */

// Company holds basic account owner data.
type Company struct {
	Name     string  `json:"name"`
	Document *string `json:"document"`
}

// Balance is returned by [Client.GetBalance].
type Balance struct {
	// Balance available, in BRL.
	Balance  float64 `json:"balance"`
	Currency string  `json:"currency"`
	Company  Company `json:"company"`
}

// SMSTypeItem is one row of [Client.GetSMSTypes].
type SMSTypeItem struct {
	// ID is the value to pass as SMSTypeID.
	ID   int    `json:"id"`
	Name string `json:"name"`
	// Price is the unit price (BRL).
	Price float64 `json:"price"`
	// Sale is the promotional unit price (BRL), if any.
	Sale *float64 `json:"sale"`
}

// AutoRechargeConfig is returned by [Client.GetAutoRecharge] and [Client.SetAutoRecharge].
type AutoRechargeConfig struct {
	Enabled bool `json:"enabled"`
	// Threshold at which a recharge triggers (BRL).
	Threshold float64 `json:"threshold"`
	// PlanQuantity of credits bought on each recharge.
	PlanQuantity int     `json:"planQuantity"`
	CardID       *string `json:"cardId"`
	AlertEnabled bool    `json:"alertEnabled"`
	// AlertThreshold for the low-balance e-mail alert (BRL).
	AlertThreshold float64 `json:"alertThreshold"`
}

// AutoRechargeUpdate are the parameters for [Client.SetAutoRecharge]. Use
// pointer fields so unset values are stripped from the request body.
type AutoRechargeUpdate struct {
	Enabled *bool `json:"enabled,omitempty"`
	// Threshold recharges when the balance is <= this value (BRL).
	Threshold *float64 `json:"threshold,omitempty"`
	// PlanQuantity of credits bought on each recharge.
	PlanQuantity *int    `json:"plan_quantity,omitempty"`
	CardID       *string `json:"card_id,omitempty"`
	AlertEnabled *bool   `json:"alert_enabled,omitempty"`
	// AlertThreshold e-mails you when the balance is <= this value (BRL).
	AlertThreshold *float64 `json:"alert_threshold,omitempty"`
}

// WebhookConfig is returned by [Client.GetWebhook] and [Client.SetWebhook].
type WebhookConfig struct {
	// URL configured (nil = disabled).
	URL *string `json:"url"`
	// Secret is the HMAC secret. Sign the raw body to validate X-SMSGo-Signature.
	Secret *string `json:"secret"`
}

// WebhookUpdate are the parameters for [Client.SetWebhook].
type WebhookUpdate struct {
	// URL is your HTTPS endpoint. An empty string disables the webhook.
	URL *string `json:"url,omitempty"`
	// RotateSecret generates a new signing secret.
	RotateSecret *bool `json:"rotate_secret,omitempty"`
}

/* -------------------------------------------------------------------------- */
/* Billing                                                                    */
/* -------------------------------------------------------------------------- */

// Plan is a recharge tier from [BillingResource.Plans].
type Plan struct {
	ID       string  `json:"id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Sale     float64 `json:"sale"`
	// Unit is the effective unit price (BRL).
	Unit float64 `json:"unit"`
	// Total of the package (BRL).
	Total   float64 `json:"total"`
	Popular bool    `json:"popular"`
}

// Card is a saved card from [BillingResource.Cards].
type Card struct {
	ID string `json:"id"`
	// Number is the last 4 digits.
	Number string  `json:"number"`
	Name   string  `json:"name"`
	Alias  *string `json:"alias"`
	// Validate is the expiry, MM/YY.
	Validate string `json:"validate"`
	Flag     string `json:"flag"`
	Default  bool   `json:"default"`
}

// InvoiceStatus is the status object of an [InvoiceItem].
type InvoiceStatus struct {
	Code  string  `json:"code"`
	Name  string  `json:"name"`
	Icon  *string `json:"icon"`
	Color *string `json:"color"`
}

// InvoiceCard is the card object of an [InvoiceItem].
type InvoiceCard struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// InvoiceItem is one row of [BillingResource.Invoices].
type InvoiceItem struct {
	UUID      string         `json:"uuid"`
	Total     float64        `json:"total"`
	Date      string         `json:"date"`
	Expiry    string         `json:"expiry"`
	DisplayID int            `json:"displayId"`
	Status    *InvoiceStatus `json:"status"`
	Card      *InvoiceCard   `json:"card"`
}

// InvoicesParams are the parameters for [BillingResource.Invoices].
type InvoicesParams struct {
	Page    int
	PerPage int
}

// PurchaseParams are the parameters for [BillingResource.Purchase].
type PurchaseParams struct {
	// Quantity of credits (250–1,000,000). Ignored when PlanID is set.
	Quantity int `json:"quantity,omitempty"`
	// PlanID is a package UUID (tier). Takes priority over Quantity.
	PlanID string `json:"plan_id,omitempty"`
	// CardID is a saved-card UUID (optional; uses the default card when empty).
	CardID string `json:"card_id,omitempty"`
	// Coupon code (optional).
	Coupon string `json:"coupon,omitempty"`
}

// PurchaseResult is returned by [BillingResource.Purchase].
type PurchaseResult struct {
	// Status "succeeded" already credited the balance; "processing" confirms via webhook.
	Status      string `json:"status"`
	InvoiceUUID string `json:"invoiceUuid"`
	// Total charged (BRL).
	Total           float64 `json:"total"`
	Quantity        int     `json:"quantity"`
	PaymentIntentID string  `json:"paymentIntentId"`
}

/* -------------------------------------------------------------------------- */
/* Contacts and lists                                                         */
/* -------------------------------------------------------------------------- */

// ContactInput is the body for [ContactsResource.Create] and [ContactsResource.Update].
type ContactInput struct {
	FullName string `json:"full_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email,omitempty"`
	// Lists are the UUIDs of lists to associate the contact with.
	Lists []string `json:"lists,omitempty"`
}

// ContactDetail is returned by [ContactsResource.Get].
type ContactDetail struct {
	FullName string  `json:"fullName"`
	Email    *string `json:"email"`
	Phone    string  `json:"phone"`
}

// ContactsListParams are the parameters for [ContactsResource.List].
type ContactsListParams struct {
	// Page is required.
	Page    int
	PerPage int
	Search  string
	// Title filters contacts by list name.
	Title string
}

// ListInput is the body for [ListsResource.Create] and [ListsResource.Update].
type ListInput struct {
	// Name of the list (2–20 characters).
	Name string `json:"name"`
}

// ListResult is returned by list CRUD methods.
type ListResult struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// ListsListParams are the parameters for [ListsResource.List].
type ListsListParams struct {
	// Page is required.
	Page    int
	PerPage int
	Title   string
}

// MessageResult is the {message} response of delete endpoints.
type MessageResult struct {
	Message string `json:"message"`
}
