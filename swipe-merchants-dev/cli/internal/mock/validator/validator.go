package validator

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/gorillamux"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/mock/problemdetails"
)

// Validator parses the embedded spec once and validates incoming requests
// against it. The result can be wired in as a chi middleware.
type Validator struct {
	router routers.Router
}

// New builds a Validator from raw OpenAPI YAML bytes (typically pulled from
// internal/spec.Spec.RawYAML()).
func New(rawYAML []byte) (*Validator, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	doc, err := loader.LoadFromData(rawYAML)
	if err != nil {
		return nil, fmt.Errorf("validator: load spec: %w", err)
	}
	// Replace the spec's `servers` with a permissive entry so requests
	// against localhost match a server URL during routing; otherwise
	// kin-openapi rejects them as "no matching server".
	doc.Servers = openapi3.Servers{{URL: "/"}}
	if err := doc.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("validator: spec validation: %w", err)
	}
	r, err := gorillamux.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("validator: build router: %w", err)
	}
	return &Validator{router: r}, nil
}

// Middleware returns an http middleware that validates every request whose
// path matches a spec operation. Paths that do not match a spec operation
// (e.g. /_admin/*, /.well-known/*, /oauth2/*) are passed through unchanged.
//
// Validation failures emit ProblemDetails 400 with VALIDATION_ERROR and a
// field-level errors list when the cause is a specific parameter or body.
func (v *Validator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !shouldValidate(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}
		route, pathParams, err := v.router.FindRoute(r)
		if err != nil {
			// No matching operation. Let downstream chi return 404 if the
			// path isn't routable in the mock either.
			next.ServeHTTP(w, r)
			return
		}
		input := &openapi3filter.RequestValidationInput{
			Request:    r,
			PathParams: pathParams,
			Route:      route,
			Options: &openapi3filter.Options{
				AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
			},
		}
		if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
			problemdetails.Write(w, http.StatusBadRequest, toProblemDetails(err))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// shouldValidate decides whether kin-openapi should attempt to validate the
// request. Admin and well-known paths are intentionally out of spec scope.
func shouldValidate(path string) bool {
	switch {
	case strings.HasPrefix(path, "/_admin"):
		return false
	case strings.HasPrefix(path, "/.well-known"):
		return false
	case strings.HasPrefix(path, "/oauth2"):
		return false
	default:
		return true
	}
}

// toProblemDetails builds a VALIDATION_ERROR problem from a kin-openapi
// validation error. The library returns rich structured errors but the
// problem schema only supports {name, reason} pairs; we summarize.
func toProblemDetails(err error) problemdetails.ProblemDetails {
	p := problemdetails.New(problemdetails.TypeValidationError, "request validation failed")
	if verr, ok := err.(*openapi3filter.RequestError); ok {
		var field string
		if verr.Parameter != nil {
			field = verr.Parameter.Name
		}
		if field == "" && verr.RequestBody != nil {
			field = "body"
		}
		reason := verr.Reason
		if reason == "" && verr.Err != nil {
			reason = verr.Err.Error()
		}
		if field == "" {
			field = "request"
		}
		p = p.WithFieldError(field, reason)
		return p
	}
	p = p.WithFieldError("request", err.Error())
	return p
}
