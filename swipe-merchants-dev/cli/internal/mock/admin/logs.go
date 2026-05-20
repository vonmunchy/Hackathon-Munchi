package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// LogsHandler serves the /_admin/logs surface used by `swipe logs tail` /
// `swipe logs show`. It is mounted under the localhostOnly gate.
type LogsHandler struct {
	Store *store.Store
}

// Mount registers the handler's routes onto r at /_admin/logs.
func (h *LogsHandler) Mount(r chi.Router) {
	r.Route("/logs", func(lr chi.Router) {
		lr.Get("/", h.tail)
		lr.Get("/{id}", h.show)
	})
}

// DefaultTailLimit is the page size used by /_admin/logs when limit is
// omitted. Matches `swipe logs tail -n` default.
const DefaultTailLimit = 50

func (h *LogsHandler) tail(w http.ResponseWriter, r *http.Request) {
	limit := DefaultTailLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 0 {
			problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "limit must be a non-negative integer")
			return
		}
		if n > 0 {
			limit = n
		}
	}
	entries, err := h.Store.TailRequestLog(limit)
	if err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "tail logs: %v", err)
		return
	}
	if entries == nil {
		entries = []store.RequestLogEntry{}
	}
	writeJSON(w, http.StatusOK, entries)
}

func (h *LogsHandler) show(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, err := h.Store.GetRequestLog(id)
	switch {
	case errors.Is(err, store.ErrNotFound):
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "log entry %s not found", id)
		return
	case err != nil:
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "load log: %v", err)
		return
	}
	writeJSON(w, http.StatusOK, entry)
}
