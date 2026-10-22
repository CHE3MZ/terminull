package termcore

import (
	"runtime/debug"
	"strings"
)

// Version and Commit are set at link time via -ldflags for releases:
//
//	go build -ldflags "-X github.com/CHE3MZ/terminull/v2/termcore.Version=1.2.3 -X github.com/CHE3MZ/terminull/v2/termcore.Commit=abc123"
//
// When left as "unknown" (plain `go build` / `go install`), Version falls
// back to the module version from the build info, else stays "unknown".
var (
	Version = "unknown"
	Commit  = "unknown"
)

// FullVersion returns "Version-Commit" for --version output.
func FullVersion() string {
	return versionWithFallback(Version) + "-" + Commit
}

func versionWithFallback(v string) string {
	if v != "unknown" {
		return v
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	// Build info versions look like "v1.2.3"; strip the prefix for
	// consistency with -ldflags builds. "(devel)" means plain `go build`,
	// which carries no version, so keep "unknown" instead.
	trimmed := strings.TrimPrefix(info.Main.Version, "v")
	if trimmed == "" || trimmed == "(devel)" {
		return "unknown"
	}

	return trimmed
}
