package smsgo

import (
	"context"
	"net/http"
)

// BillingResource is the billing namespace, reachable via [Client.Billing].
type BillingResource struct {
	do doFunc
}

// Plans returns the available recharge tiers.
func (r *BillingResource) Plans(ctx context.Context) ([]Plan, error) {
	var wrapper struct {
		Data []Plan `json:"data"`
	}
	if err := r.do(ctx, http.MethodGet, "/v1/billing/plans", nil, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Data, nil
}

// Cards returns the saved cards (last 4 digits only).
func (r *BillingResource) Cards(ctx context.Context) ([]Card, error) {
	var wrapper struct {
		Data []Card `json:"data"`
	}
	if err := r.do(ctx, http.MethodGet, "/v1/billing/cards", nil, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Data, nil
}

// Invoices returns the invoice/receipt history (paginated).
func (r *BillingResource) Invoices(ctx context.Context, params InvoicesParams) (*Paginated[InvoiceItem], error) {
	q := buildQuery(map[string]string{
		"page":    intOrEmpty(params.Page),
		"perPage": intOrEmpty(params.PerPage),
	})
	var out Paginated[InvoiceItem]
	if err := r.do(ctx, http.MethodGet, "/v1/billing/invoices"+q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Purchase buys credits by charging a saved card (off-session). Set Quantity or
// PlanID. Without CardID the default card is used.
//
// Idempotency: each call creates a new charge. On timeout, query [BillingResource.Invoices]
// before retrying — do NOT blindly retry.
func (r *BillingResource) Purchase(ctx context.Context, params PurchaseParams) (*PurchaseResult, error) {
	var out PurchaseResult
	if err := r.do(ctx, http.MethodPost, "/v1/billing/purchase", params, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
