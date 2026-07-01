package smsgo

import (
	"context"
	"net/http"
	"net/url"
)

// ListsResource is the lists namespace, reachable via [Client.Lists].
type ListsResource struct {
	do doFunc
}

// List lists the account's lists (paginated; Page is required).
func (r *ListsResource) List(ctx context.Context, params ListsListParams) (*Paginated[map[string]any], error) {
	q := buildQuery(map[string]string{
		"page":    intOrEmpty(params.Page),
		"perPage": intOrEmpty(params.PerPage),
		"title":   params.Title,
	})
	var out Paginated[map[string]any]
	if err := r.do(ctx, http.MethodGet, "/v1/lists/list"+q, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Create creates a list.
func (r *ListsResource) Create(ctx context.Context, input ListInput) (*ListResult, error) {
	var out ListResult
	if err := r.do(ctx, http.MethodPost, "/v1/lists/store", input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Get details a list by its UUID.
func (r *ListsResource) Get(ctx context.Context, id string) (*ListResult, error) {
	var out ListResult
	path := "/v1/lists/" + url.PathEscape(id) + "/show"
	if err := r.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Update updates a list.
func (r *ListsResource) Update(ctx context.Context, id string, input ListInput) (*ListResult, error) {
	var out ListResult
	path := "/v1/lists/" + url.PathEscape(id) + "/update"
	if err := r.do(ctx, http.MethodPut, path, input, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Delete deletes a list.
func (r *ListsResource) Delete(ctx context.Context, id string) (*MessageResult, error) {
	var out MessageResult
	path := "/v1/lists/" + url.PathEscape(id) + "/delete"
	if err := r.do(ctx, http.MethodDelete, path, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
