package termupdate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Release is the latest release reported by a Source.
type Release struct {
	Version string
	URL     string
}

// Source returns the latest release for an update check.
type Source interface {
	Latest(ctx context.Context) (Release, error)
}

// SourceFunc adapts a function to a Source.
type SourceFunc func(ctx context.Context) (Release, error)

// Latest calls f.
func (f SourceFunc) Latest(ctx context.Context) (Release, error) {
	if f == nil {
		return Release{}, errors.New("termupdate: nil source func")
	}
	return f(ctx)
}

// Options tunes Check. CurrentVersion is the running version, with or
// without a leading v.
type Options struct {
	AppName        string
	CurrentVersion string
	// StateFile overrides the default cache path. Empty uses the user
	// cache directory and AppName.
	StateFile string
	// Interval is the minimum time between network checks. Zero uses
	// 24 hours.
	Interval time.Duration
	// Source supplies the latest release. Check never calls it while the
	// cached result is fresh.
	Source Source
}

// Status describes a completed or cached update check.
type Status struct {
	AppName   string
	Current   string
	Latest    string
	URL       string
	Available bool
	CheckedAt time.Time
	FromCache bool
}

type updateState struct {
	CheckedAt time.Time `json:"checked_at"`
	Latest    string    `json:"latest"`
	URL       string    `json:"url"`
}

// Check asks Source for the latest release, throttled by a state file. A
// cached result inside Interval returns without a call to Source. The
// state file is best-effort: a write failure does not hide a successful
// check.
func Check(ctx context.Context, opts Options) (Status, error) {
	app := opts.AppName
	if app == "" {
		app = "terminull"
	}
	status := Status{AppName: app, Current: normalizeVersion(opts.CurrentVersion)}
	if opts.Source == nil {
		return status, errors.New("termupdate: source is required")
	}
	interval := opts.Interval
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	stateFile := opts.StateFile
	if stateFile == "" {
		stateFile = DefaultStateFile(app)
	}

	if state, ok := readState(stateFile); ok {
		status.CheckedAt = state.CheckedAt
		status.Latest = normalizeVersion(state.Latest)
		status.URL = state.URL
		status.Available = newerVersion(status.Latest, status.Current)
		if time.Since(state.CheckedAt) < interval {
			status.FromCache = true
			return status, nil
		}
	}

	release, err := opts.Source.Latest(ctx)
	if err != nil {
		return status, err
	}
	if strings.TrimSpace(release.Version) == "" {
		return status, errors.New("termupdate: source returned an empty version")
	}
	state := updateState{
		CheckedAt: time.Now().UTC(),
		Latest:    release.Version,
		URL:       release.URL,
	}
	status.CheckedAt = state.CheckedAt
	status.Latest = normalizeVersion(state.Latest)
	status.URL = state.URL
	status.Available = newerVersion(status.Latest, status.Current)
	_ = writeState(stateFile, state)
	return status, nil
}

// Notify writes a short notice when a newer release exists.
func Notify(w io.Writer, status Status) error {
	if !status.Available {
		return nil
	}
	_, err := fmt.Fprintf(w, "%s %s is available (running %s): %s\n",
		status.AppName, status.Latest, status.Current, status.URL)
	return err
}

// DefaultStateFile returns the default update cache path for app. It
// returns an empty string when the user cache directory is unavailable.
func DefaultStateFile(app string) string {
	dir, err := os.UserCacheDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, app, "update-check.json")
}

func readState(path string) (updateState, bool) {
	if path == "" {
		return updateState{}, false
	}
	data, err := os.ReadFile(path) // #nosec G304 -- state path is caller-selected cache or test path.
	if err != nil {
		return updateState{}, false
	}
	var state updateState
	if err := json.Unmarshal(data, &state); err != nil || state.CheckedAt.IsZero() || state.Latest == "" {
		return updateState{}, false
	}
	return state, true
}

func writeState(path string, state updateState) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "v")
	if i := strings.IndexAny(v, "-+"); i >= 0 {
		v = v[:i]
	}
	return v
}

func newerVersion(latest, current string) bool {
	latest = normalizeVersion(latest)
	current = normalizeVersion(current)
	if latest == "" || current == "" || current == "unknown" {
		return false
	}
	l, okL := parseVersion(latest)
	c, okC := parseVersion(current)
	if !okL || !okC {
		return false
	}
	for i := range l {
		if l[i] != c[i] {
			return l[i] > c[i]
		}
	}
	return false
}

func parseVersion(v string) ([3]int, bool) {
	var out [3]int
	parts := strings.Split(v, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return out, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
