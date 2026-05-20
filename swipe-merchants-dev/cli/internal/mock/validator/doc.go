// Package validator wraps kin-openapi for runtime request validation
// against the embedded spec (scaffold §9 P3). The mock pipes every
// API request through Validate, which rejects malformed query
// parameters, headers, and request bodies with ProblemDetails 400.
package validator
