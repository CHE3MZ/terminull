package termapp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func updateServer(t *testing.T, tag string, requests *int) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*requests++
		if r.URL.Path != "/repos/acme/tool/releases/latest" {
			t.Errorf("path = %q, want release endpoint", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"tag_name": tag,
			"html_url": "https://example.test/releases/" + tag,
		})
	}))
}

func TestCheckUpdateAvailableAndCached(t *testing.T) {
	t.Parallel()

	requests := 0
	srv := updateServer(t, "v1.2.4", &requests)
	defer srv.Close()

	state := filepath.Join(t.TempDir(), "update.json")
	opts := UpdateOptions{
		AppName:        "demo",
		Repo:           "acme/tool",
		CurrentVersion: "1.2.3",
		StateFile:      state,
		Interval:       time.Hour,
		BaseURL:        srv.URL,
	}

	status, err := CheckUpdate(context.Background(), opts)
	if err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}
	if !status.Available || status.Latest != "1.2.4" || status.FromCache {
		t.Fatalf("status = %+v, want available 1.2.4 live", status)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1", requests)
	}

	status, err = CheckUpdate(context.Background(), opts)
	if err != nil {
		t.Fatalf("cached CheckUpdate() error = %v", err)
	}
	if !status.Available || !status.FromCache {
		t.Fatalf("cached status = %+v, want available from cache", status)
	}
	if requests != 1 {
		t.Fatalf("requests = %d, want 1 after cache hit", requests)
	}
}

func TestCheckUpdateNoNewerRelease(t *testing.T) {
	t.Parallel()

	requests := 0
	srv := updateServer(t, "v1.2.3", &requests)
	defer srv.Close()

	status, err := CheckUpdate(context.Background(), UpdateOptions{
		Repo:           "acme/tool",
		CurrentVersion: "1.2.3",
		StateFile:      filepath.Join(t.TempDir(), "update.json"),
		BaseURL:        srv.URL,
	})
	if err != nil {
		t.Fatalf("CheckUpdate() error = %v", err)
	}
	if status.Available {
		t.Fatalf("status = %+v, want no update", status)
	}
}

func TestCheckUpdateHTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := CheckUpdate(context.Background(), UpdateOptions{
		Repo:           "acme/tool",
		CurrentVersion: "1.2.3",
		StateFile:      filepath.Join(t.TempDir(), "update.json"),
		BaseURL:        srv.URL,
	})
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("CheckUpdate() error = %v, want 403", err)
	}
}

func TestNotifyUpdate(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	if err := NotifyUpdate(&b, UpdateStatus{AppName: "demo", Current: "1.0.0", Latest: "1.1.0", URL: "https://example.test", Available: true}); err != nil {
		t.Fatalf("NotifyUpdate() error = %v", err)
	}
	if got := b.String(); !strings.Contains(got, "demo 1.1.0") || !strings.Contains(got, "https://example.test") {
		t.Fatalf("notice = %q", got)
	}
}

func TestVersionComparison(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		latest, current string
		want            bool
	}{
		{"1.2.4", "1.2.3", true},
		{"v2.0.0", "1.9.9", true},
		{"1.2.3", "1.2.3", false},
		{"1.2.2", "1.2.3", false},
		{"unknown", "1.0.0", false},
		{"1.0", "0.9.9", true},
	} {
		if got := newerVersion(tc.latest, tc.current); got != tc.want {
			t.Errorf("newerVersion(%q, %q) = %v, want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}

func TestDefaultUpdateStateFileUsesCacheDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("LOCALAPPDATA", dir)

	got := defaultUpdateStateFile("demo")
	if got == "" || !strings.Contains(got, "demo") || !strings.HasSuffix(got, "update-check.json") {
		t.Fatalf("defaultUpdateStateFile() = %q", got)
	}
	if _, err := os.Stat(filepath.Dir(got)); err == nil {
		t.Fatal("cache directory should not be created until a check succeeds")
	}
}
