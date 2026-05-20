package transport

import (
	"context"
	"net/http"
	"strconv"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// TailLogs calls GET /_admin/logs[?limit=N] and returns the newest entries
// first. Pass limit=0 to use the server default.
func (c *AdminClient) TailLogs(ctx context.Context, limit int) ([]store.RequestLogEntry, error) {
	path := "/_admin/logs"
	if limit > 0 {
		path += "?limit=" + strconv.Itoa(limit)
	}
	var out []store.RequestLogEntry
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	if err := cli.do(ctx, http.MethodGet, path, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ShowLog calls GET /_admin/logs/{id}.
func (c *AdminClient) ShowLog(ctx context.Context, id string) (store.RequestLogEntry, error) {
	var out store.RequestLogEntry
	cli := &Client{BaseURL: c.BaseURL, HTTPClient: c.HTTPClient}
	if err := cli.do(ctx, http.MethodGet, "/_admin/logs/"+id, nil, &out); err != nil {
		return store.RequestLogEntry{}, err
	}
	return out, nil
}
