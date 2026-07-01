package smsgo

import (
	"context"
	"net/http"
	"net/url"
)

// doFunc is the shared transport (auth/refresh/errors) used by resources.
type doFunc func(ctx context.Context, method, path string, payload, out any) error

// ContactsResource is the contacts namespace, reachable via [Client.Contacts].
type ContactsResource struct {
	do doFunc
}

// List lists contacts (paginated; Page is required).
func (r *ContactsResource) List(ctx context.Context, params ContactsListParams) (*Paginated[map[string]any], error) {
	q := buildQuery(map[string]string{
		"page":    intOrEmpty(params.Page),
		"perPage": intOrEmpty(params.PerPage),
		"search":  params.Search,
		"title":   params.Title,
	})
	var out Paginated[map[string]any]
	if err := r.do(ctx, http.MethodGet, "/v1/contacts/list"+q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create creates (or upserts by phone) a contact and returns its UUID.
func (r *ContactsResource) Create(ctx context.Context, input ContactInput) (string, error) {
	var id string
	if err := r.do(ctx, http.MethodPost, "/v1/contacts/store", input, &id); err != nil {
		return "", err
	}
	return id, nil
}

// Get details a contact by its UUID.
func (r *ContactsResource) Get(ctx context.Context, id string) (*ContactDetail, error) {
	var out ContactDetail
	path := "/v1/contacts/" + url.PathEscape(id) + "/show"
	if err := r.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update updates a contact and returns its UUID.
func (r *ContactsResource) Update(ctx context.Context, id string, input ContactInput) (string, error) {
	var newID string
	path := "/v1/contacts/" + url.PathEscape(id) + "/update"
	if err := r.do(ctx, http.MethodPut, path, input, &newID); err != nil {
		return "", err
	}
	return newID, nil
}

// Delete deletes a contact.
func (r *ContactsResource) Delete(ctx context.Context, id string) (*MessageResult, error) {
	var out MessageResult
	path := "/v1/contacts/" + url.PathEscape(id) + "/delete"
	if err := r.do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
