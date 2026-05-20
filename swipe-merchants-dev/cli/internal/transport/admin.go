package transport

import (
	"context"
	"net/http"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/admin"
)

// AdminClient calls the out-of-spec /_admin/* endpoints exposed by the
// running mock. It is built on the same http.Client as Client; the
// distinction is which endpoint set it speaks to (D-009).
type AdminClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewAdminClient builds an AdminClient targeting baseURL.
func NewAdminClient(baseURL string) *AdminClient {
	return &AdminClient{
		BaseURL:    NewClient(baseURL).BaseURL,
		HTTPClient: NewClient(baseURL).HTTPClient,
	}
}

// Status calls GET /_admin/status and returns the decoded response.
func (c *AdminClient) Status(ctx context.Context) (admin.StatusResponse, error) {
	var out admin.StatusResponse
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	if err := cli.do(ctx, http.MethodGet, "/_admin/status", nil, &out); err != nil {
		return admin.StatusResponse{}, err
	}
	return out, nil
}
