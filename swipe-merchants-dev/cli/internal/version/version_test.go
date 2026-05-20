package version_test

import (
	"testing"

	"github.com/BML-Digital/swipe-merchants-dev/cli/internal/version"
)

func TestGet_ReturnsBuildTimeVars(t *testing.T) {
	t.Parallel()
	// Stash and restore the package vars so we can assert that Get reads them.
	origVersion, origCommit, origBuildDate, origSpec := version.Version, version.Commit, version.BuildDate, version.SpecVersion
	t.Cleanup(func() {
		version.Version, version.Commit, version.BuildDate, version.SpecVersion = origVersion, origCommit, origBuildDate, origSpec
	})

	version.Version = "v1.2.3"
	version.Commit = "abcdef0"
	version.BuildDate = "2026-01-01T00:00:00Z"
	version.SpecVersion = "9.9.9"

	got := version.Get()
	if got.Version != "v1.2.3" {
		t.Errorf("Version = %q", got.Version)
	}
	if got.Commit != "abcdef0" {
		t.Errorf("Commit = %q", got.Commit)
	}
	if got.BuildDate != "2026-01-01T00:00:00Z" {
		t.Errorf("BuildDate = %q", got.BuildDate)
	}
	if got.SpecVersion != "9.9.9" {
		t.Errorf("SpecVersion = %q", got.SpecVersion)
	}
}
