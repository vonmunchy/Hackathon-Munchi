package admin

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/scenarios"
	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/store"
)

// ScenariosHandler serves /_admin/scenarios. Localhost-gated upstream.
type ScenariosHandler struct {
	Store *store.Store
	// OnEnable is invoked after a successful enable so callers can run
	// any side-effects (e.g. reset webhook_delivery_fail counter,
	// schedule client_revoked_after timer).
	OnEnable func(name string, args map[string]string)
	// OnDisable mirrors OnEnable.
	OnDisable func(name string)
}

// Mount registers the routes on r at /_admin/scenarios.
func (h *ScenariosHandler) Mount(r chi.Router) {
	r.Route("/scenarios", func(sr chi.Router) {
		sr.Get("/", h.list)
		sr.Get("/{name}", h.show)
		sr.Post("/{name}/enable", h.enable)
		sr.Post("/{name}/disable", h.disable)
	})
}

// ScenarioView is the wire shape returned by list/show. It blends static
// registry metadata with the per-instance enabled state.
type ScenarioView struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Args        []scenarios.ArgDef `json:"args,omitempty"`
	Enabled     bool               `json:"enabled"`
	CurrentArgs map[string]string  `json:"current_args,omitempty"`
}

func (h *ScenariosHandler) list(w http.ResponseWriter, _ *http.Request) {
	out := make([]ScenarioView, 0, len(scenarios.List()))
	for _, d := range scenarios.List() {
		view := ScenarioView{
			Name:        d.Name,
			Description: d.Description,
			Args:        d.Args,
		}
		if st, ok, _ := h.Store.GetScenario(d.Name); ok {
			view.Enabled = st.Enabled
			view.CurrentArgs = st.Args
		}
		out = append(out, view)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ScenariosHandler) show(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	d, ok := scenarios.Get(name)
	if !ok {
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "scenario %s not found", name)
		return
	}
	view := ScenarioView{Name: d.Name, Description: d.Description, Args: d.Args}
	if st, ok, _ := h.Store.GetScenario(name); ok {
		view.Enabled = st.Enabled
		view.CurrentArgs = st.Args
	}
	writeJSON(w, http.StatusOK, view)
}

// enableRequest is the body of POST /_admin/scenarios/{name}/enable.
// Empty body is allowed — defaults from the registry apply.
type enableRequest struct {
	Args map[string]string `json:"args,omitempty"`
}

func (h *ScenariosHandler) enable(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if _, ok := scenarios.Get(name); !ok {
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "scenario %s not found", name)
		return
	}
	var req enableRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			problemdetails.WriteNew(w, http.StatusBadRequest, problemdetails.TypeValidationError, "decode body: %v", err)
			return
		}
	}
	if err := h.Store.EnableScenario(name, req.Args); err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "enable: %v", err)
		return
	}
	if h.OnEnable != nil {
		h.OnEnable(name, req.Args)
	}
	st, _, _ := h.Store.GetScenario(name)
	writeJSON(w, http.StatusOK, st)
}

func (h *ScenariosHandler) disable(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if _, ok := scenarios.Get(name); !ok {
		problemdetails.WriteNew(w, http.StatusNotFound, problemdetails.TypeNotFound, "scenario %s not found", name)
		return
	}
	if err := h.Store.DisableScenario(name); err != nil {
		problemdetails.WriteNew(w, http.StatusInternalServerError, problemdetails.TypeValidationError, "disable: %v", err)
		return
	}
	if h.OnDisable != nil {
		h.OnDisable(name)
	}
	w.WriteHeader(http.StatusNoContent)
}
