package termapp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// UpdateOptions tunes CheckUpdate. Repo is "owner/name"; CurrentVersion
// is the running version, with or without a leading v.
type UpdateOptions struct {
	AppName        string
	Repo           string
	CurrentVersion string
	// StateFile overrides the default cache path. Empty uses the user
	// cache directory and AppName.
	StateFile string
	// Interval is the minimum time between network checks. Zero uses
	// 24 hours.
	Interval time.Duration
	// Client and BaseURL are test seams; nil and empty use net/http and
	// the public GitHub API.
	Client  *http.Client
	BaseURL string
}

// UpdateStatus describes a completed or cached update check.
type UpdateStatus struct {
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

// CheckUpdate asks GitHub for the latest release, throttled by a state
// file. A cached result inside Interval returns without a network call.
// The state file is best-effort: a write failure does not hide a
// successful check.
func CheckUpdate(ctx context.Context, opts UpdateOptions) (UpdateStatus, error) {
	app := opts.AppName
	if app == "" {
		app = "terminull"
	}
	status := UpdateStatus{AppName: app, Current: normalizeVersion(opts.CurrentVersion)}
	if opts.Repo == "" {
		return status, fmt.Errorf("update: Repo is required")
	}
	interval := opts.Interval
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	stateFile := opts.StateFile
	if stateFile == "" {
		stateFile = defaultUpdateStateFile(app)
	}

	if state, ok := readUpdateState(stateFile); ok {
		status.CheckedAt = state.CheckedAt
		status.Latest = normalizeVersion(state.Latest)
		status.URL = state.URL
		status.Available = newerVersion(status.Latest, status.Current)
		if time.Since(state.CheckedAt) < interval {
			status.FromCache = true
			return status, nil
		}
	}

	state, err := fetchLatestRelease(ctx, opts)
	if err != nil {
		return status, err
	}
	status.CheckedAt = state.CheckedAt
	status.Latest = normalizeVersion(state.Latest)
	status.URL = state.URL
	status.Available = newerVersion(status.Latest, status.Current)
	_ = writeUpdateState(stateFile, state)
	return status, nil
}

// NotifyUpdate writes a short notice when a newer release exists.
func NotifyUpdate(w io.Writer, status UpdateStatus) error {
	if !status.Available {
		return nil
	}
	_, err := fmt.Fprintf(w, "%s %s is available (running %s): %s\n",
		status.AppName, status.Latest, status.Current, status.URL)
	return err
}

func defaultUpdateStateFile(app string) string {
	dir, err := os.UserCacheDir()
	if err != nil || dir == "" {
		return ""
	}
	return filepath.Join(dir, app, "update-check.json")
}

func readUpdateState(path string) (updateState, bool) {
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

func writeUpdateState(path string, state updateState) error {
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

func fetchLatestRelease(ctx context.Context, opts UpdateOptions) (updateState, error) {
	base := strings.TrimRight(opts.BaseURL, "/")
	if base == "" {
		base = "https://api.github.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/repos/"+opts.Repo+"/releases/latest", nil)
	if err != nil {
		return updateState{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "terminull-update-check")

	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return updateState{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return updateState{}, fmt.Errorf("update: GitHub returned %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&release); err != nil {
		return updateState{}, err
	}
	if release.TagName == "" {
		return updateState{}, fmt.Errorf("update: latest release has no tag")
	}
	return updateState{
		CheckedAt: time.Now().UTC(),
		Latest:    release.TagName,
		URL:       release.HTMLURL,
	}, nil
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
