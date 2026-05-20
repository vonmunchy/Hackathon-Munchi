package output_test

import (
	"errors"
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/cli/output"
)

// errWriter always fails. Used to drive the error branches of the output
// renderers.
type errWriter struct{}

func (errWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("simulated write failure")
}

func TestRaw_WriterError_Surfaces(t *testing.T) {
	t.Parallel()
	r := output.New(output.FormatYAML, errWriter{}, false)
	if err := r.Raw([]byte("anything")); err == nil {
		t.Errorf("expected write error; got nil")
	}
}

func TestObject_JSON_NonMarshalable_Errors(t *testing.T) {
	t.Parallel()
	var buf nullWriter
	r := output.New(output.FormatJSON, &buf, false)
	// `chan` is not encodable as JSON; encoding/json returns an error.
	ch := make(chan int)
	if err := r.Object(ch); err == nil {
		t.Errorf("expected encode error for chan; got nil")
	}
}

func TestObject_JSON_WriterFails_Errors(t *testing.T) {
	t.Parallel()
	r := output.New(output.FormatJSON, errWriter{}, false)
	if err := r.Object(map[string]string{"a": "b"}); err == nil {
		t.Errorf("expected write error from json encoder; got nil")
	}
}

func TestObject_YAML_WriterFails_Errors(t *testing.T) {
	t.Parallel()
	r := output.New(output.FormatYAML, errWriter{}, false)
	if err := r.Object(map[string]string{"a": "b"}); err == nil {
		t.Errorf("expected write error from yaml encoder; got nil")
	}
}

// nullWriter accepts everything but discards it; used when only the
// encoder error path matters.
type nullWriter struct{}

func (*nullWriter) Write(p []byte) (int, error) { return len(p), nil }
