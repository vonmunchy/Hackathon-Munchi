package version

// Build-time variables injected via -ldflags. See the project Makefile for the
// exact -X assignments. Defaults make `go run` and unit tests work without
// the build wrapper.
var (
	// Version is the binary version (e.g. "v0.1.0" or "dev").
	Version = "dev"
	// Commit is the short git SHA of the build, or "unknown" outside of git.
	Commit = "unknown"
	// BuildDate is the UTC build timestamp in RFC 3339, or "unknown".
	BuildDate = "unknown"
	// SpecVersion is the OpenAPI spec version embedded into this binary.
	// Authoritative truth comes from the embedded spec at runtime; this value
	// is the build-time recorded copy and is exposed for diagnostic purposes.
	SpecVersion = "0.0.0"
)

// Info bundles the build-time identification fields for a single binary.
type Info struct {
	Version     string `json:"version" yaml:"version"`
	Commit      string `json:"commit" yaml:"commit"`
	BuildDate   string `json:"build_date" yaml:"build_date"`
	SpecVersion string `json:"spec_version" yaml:"spec_version"`
}

// Get returns the build-time Info bundle for the running binary.
func Get() Info {
	return Info{
		Version:     Version,
		Commit:      Commit,
		BuildDate:   BuildDate,
		SpecVersion: SpecVersion,
	}
}
