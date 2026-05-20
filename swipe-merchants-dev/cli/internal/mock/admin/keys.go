package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// KeysHandler serves /_admin/keys CRUD + rotate/revoke. It is mounted under
// the localhostOnly gate by the mock server, so it carries no per-request
// auth of its own (scaffold §5.8, D-009).
type KeysHandler struct {
	Store *store.Store
	// DefaultMerchantID is used when a create request omits merchant_id.
	DefaultMerchantID string
}

// ClientView is the wire shape returned by every keys endpoint. It omits
// the bcrypt hash by construction; secrets only appear in CreateResponse /
// RotateResponse on the immediate create or rotate call.
type ClientView struct {
	ID         string    `json:"id"`
	MerchantID string    `json:"merchant_id"`
	Name       string    `json:"name"`
	Scopes     []string  `json:"scopes"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateRequest is the body of POST /_admin/keys.
type CreateRequest struct {
	Name       string   `json:"name"`
	MerchantID string   `json:"merchant_id,omitempty"`
	Scopes     []string `json:"scopes,omitempty"`
}

// CreateResponse is the body of POST /_admin/keys. The plaintext client
// secret is returned exactly once on this response (D-010, D-023).
type CreateResponse struct {
	Client ClientView `json:"client"`
	Secret string     `json:"client_secret"`
}

// UpdateRequest is the body of PATCH /_admin/keys/{id}. Each field is a
// pointer so the caller can distinguish "unset" from "set to zero value".
type UpdateRequest struct {
	Name    *string   `json:"name,omitempty"`
	Scopes  *[]string `json:"scopes,omitempty"`
	Enabled *bool     `json:"enabled,omitempty"`
}

// RotateResponse is the body of POST /_admin/keys/{id}/rotate.
type RotateResponse struct {
	Client ClientView `json:"client"`
	Secret string     `json:"client_secret"`
}

// Mount registers the handler's routes onto r at /_admin/keys.
func (h *KeysHandler) Mount(r chi.Router) {
	r.Route("/keys", func(kr chi.Router) {
		kr.Post("/", h.create)
		kr.Get("/", h.list)
		kr.Get("/{id}", h.show)
		kr.Patch("/{id}", h.update)
		kr.Post("/{id}/rotate", h.rotate)
		kr.Post("/{id}/revoke", h.revoke)
		kr.Delete("/{id}", h.delete)
	})
}

func (h *KeysHandler) create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "decode body: %v", err)
		return
	}
	if req.Name == "" {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "name is required")
		return
	}
	merchant := req.MerchantID
	if merchant == "" {
		merchant = h.DefaultMerchantID
	}
	c, secret, err := h.Store.CreateClient(merchant, req.Name, req.Scopes)
	switch {
	case errors.Is(err, store.ErrNotFound):
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "merchant %s not found", merchant)
		return
	case err != nil:
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "create client: %v", err)
		return
	}
	writeJSON(w, http.StatusCreated, CreateResponse{Client: toView(c), Secret: secret})
}

func (h *KeysHandler) list(w http.ResponseWriter, _ *http.Request) {
	clients, err := h.Store.ListClients()
	if err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "list clients: %v", err)
		return
	}
	out := make([]ClientView, 0, len(clients))
	for _, c := range clients {
		out = append(out, toView(c))
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *KeysHandler) show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	c, err := h.Store.GetClient(id)
	if writeStoreError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, toView(c))
}

func (h *KeysHandler) update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "decode body: %v", err)
		return
	}
	updated, err := h.Store.UpdateClient(id, func(c store.Client) store.Client {
		if req.Name != nil {
			c.Name = *req.Name
		}
		if req.Scopes != nil {
			c.Scopes = *req.Scopes
		}
		if req.Enabled != nil {
			c.Enabled = *req.Enabled
		}
		return c
	})
	if writeStoreError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, toView(updated))
}

func (h *KeysHandler) rotate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	updated, secret, err := h.Store.RotateClientSecret(id)
	if writeStoreError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, RotateResponse{Client: toView(updated), Secret: secret})
}

func (h *KeysHandler) revoke(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	updated, err := h.Store.UpdateClient(id, func(c store.Client) store.Client {
		c.Enabled = false
		return c
	})
	if writeStoreError(w, err) {
		return
	}
	writeJSON(w, http.StatusOK, toView(updated))
}

func (h *KeysHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.Store.DeleteClient(id); err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "delete client: %v", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// writeStoreError handles the common store error mapping for the keys
// endpoints. Returns true when an error has been written (callers should
// short-circuit), false when err is nil.
func writeStoreError(w http.ResponseWriter, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, store.ErrNotFound):
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "client not found")
	default:
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "%v", err)
	}
	return true
}

func toView(c store.Client) ClientView {
	return ClientView{
		ID:         c.ID,
		MerchantID: c.MerchantID,
		Name:       c.Name,
		Scopes:     append([]string(nil), c.Scopes...),
		Enabled:    c.Enabled,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
