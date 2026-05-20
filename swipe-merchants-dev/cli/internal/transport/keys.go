package transport

import (
	"context"
	"net/http"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
)

// CreateClient calls POST /_admin/keys and returns the created client view
// plus the plaintext secret (returned exactly once per D-010).
func (c *AdminClient) CreateClient(ctx context.Context, req admin.CreateRequest) (admin.CreateResponse, error) {
	var out admin.CreateResponse
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	if err := cli.do(ctx, http.MethodPost, "/_admin/keys", req, &out); err != nil {
		return admin.CreateResponse{}, err
	}
	return out, nil
}

// ListClients calls GET /_admin/keys.
func (c *AdminClient) ListClients(ctx context.Context) ([]admin.ClientView, error) {
	var out []admin.ClientView
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	if err := cli.do(ctx, http.MethodGet, "/_admin/keys", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetClient calls GET /_admin/keys/{id}.
func (c *AdminClient) GetClient(ctx context.Context, id string) (admin.ClientView, error) {
	var out admin.ClientView
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	if err := cli.do(ctx, http.MethodGet, "/_admin/keys/"+id, nil, &out); err != nil {
		return admin.ClientView{}, err
	}
	return out, nil
}

// UpdateClient calls PATCH /_admin/keys/{id}.
func (c *AdminClient) UpdateClient(ctx context.Context, id string, req admin.UpdateRequest) (admin.ClientView, error) {
	var out admin.ClientView
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	if err := cli.do(ctx, http.MethodPatch, "/_admin/keys/"+id, req, &out); err != nil {
		return admin.ClientView{}, err
	}
	return out, nil
}

// RotateClientSecret calls POST /_admin/keys/{id}/rotate and returns the
// fresh plaintext secret.
func (c *AdminClient) RotateClientSecret(ctx context.Context, id string) (admin.RotateResponse, error) {
	var out admin.RotateResponse
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	if err := cli.do(ctx, http.MethodPost, "/_admin/keys/"+id+"/rotate", nil, &out); err != nil {
		return admin.RotateResponse{}, err
	}
	return out, nil
}

// RevokeClient calls POST /_admin/keys/{id}/revoke (disables but keeps the
// record so token verifications fail cleanly per scaffold §5.4).
func (c *AdminClient) RevokeClient(ctx context.Context, id string) (admin.ClientView, error) {
	var out admin.ClientView
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	if err := cli.do(ctx, http.MethodPost, "/_admin/keys/"+id+"/revoke", nil, &out); err != nil {
		return admin.ClientView{}, err
	}
	return out, nil
}

// DeleteClient calls DELETE /_admin/keys/{id}.
func (c *AdminClient) DeleteClient(ctx context.Context, id string) error {
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	return cli.do(ctx, http.MethodDelete, "/_admin/keys/"+id, nil, nil)
}
