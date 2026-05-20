// Package problemdetails implements the RFC 9457 ProblemDetails body used by
// every error response the mock emits (D-015). The Type field is constrained
// to the catalog declared by D-016 so producers cannot drift into ad-hoc
// strings.
//
// The package also provides Write helpers that serialize a ProblemDetails to
// an http.ResponseWriter at a given status code with the correct content
// type, so handlers and middleware can stay terse.
package problemdetails
