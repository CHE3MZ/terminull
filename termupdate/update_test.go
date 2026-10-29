package termupdate

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCheckAvailableAndCached(t *testing.T) {
	t.Parallel()

	calls := 0
	source := SourceFunc(func(context.Context) (Release, error) {
		calls++
		return Release{Version: "v1.2.4", URL: "https://example.test/releases/v1.2.4"}, nil
	})
	state := filepath.Join(t.TempDir(), "update.json")
	opts := Options{
		AppName:        "demo",
		CurrentVersion: "1.2.3",
		StateFile:      state,
		Interval:       time.Hour,
		Source:         source,
	}

	status, err := Check(context.Background(), opts)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !status.Available || status.Latest != "1.2.4" || status.FromCache {
		t.Fatalf("status = %+v, want available 1.2.4 live", status)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}

	status, err = Check(context.Background(), opts)
	if err != nil {
		t.Fatalf("cached Check() error = %v", err)
	}
	if !status.Available || !status.FromCache {
		t.Fatalf("cached status = %+v, want available from cache", status)
	}
	if calls != 1 {
		t.Fatalf("calls = %d, want 1 after cache hit", calls)
	}

	stateData, err := os.ReadFile(state) // #nosec G304 -- test-owned state path.
	if err != nil {
		t.Fatalf("ReadFile(state) error = %v", err)
	}
	if !strings.Contains(string(stateData), "1.2.4") {
		t.Fatalf("state = %q, want latest version", stateData)
	}
}

func TestCheckNoNewerRelease(t *testing.T) {
	t.Parallel()

	status, err := Check(context.Background(), Options{
		CurrentVersion: "1.2.3",
		StateFile:      filepath.Join(t.TempDir(), "update.json"),
		Source: SourceFunc(func(context.Context) (Release, error) {
			return Release{Version: "v1.2.3"}, nil
		}),
	})
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if status.Available {
		t.Fatalf("status = %+v, want no update", status)
	}
}

func TestCheckSourceErrorKeepsExpiredCache(t *testing.T) {
	t.Parallel()

	state := filepath.Join(t.TempDir(), "update.json")
	if err := writeState(state, updateState{
		CheckedAt: time.Now().Add(-2 * time.Hour).UTC(),
		Latest:    "v2.0.0",
		URL:       "https://example.test/releases/v2.0.0",
	}); err != nil {
		t.Fatalf("writeState() error = %v", err)
	}

	status, err := Check(context.Background(), Options{
		AppName:        "demo",
		CurrentVersion: "1.0.0",
		StateFile:      state,
		Interval:       time.Hour,
		Source: SourceFunc(func(context.Context) (Release, error) {
			return Release{}, errors.New("offline")
		}),
	})
	if err == nil || !strings.Contains(err.Error(), "offline") {
		t.Fatalf("Check() error = %v, want offline", err)
	}
	if !status.Available || status.Latest != "2.0.0" || status.FromCache {
		t.Fatalf("status = %+v, want stale cached release", status)
	}
}

func TestCheckRequiresSource(t *testing.T) {
	t.Parallel()

	status, err := Check(context.Background(), Options{CurrentVersion: "v1.0.0"})
	if err == nil || !strings.Contains(err.Error(), "source") {
		t.Fatalf("Check() error = %v, want source error", err)
	}
	if status.AppName != "terminull" || status.Current != "1.0.0" {
		t.Fatalf("status = %+v, want defaults", status)
	}
}

func TestCheckRejectsEmptyVersion(t *testing.T) {
	t.Parallel()

	_, err := Check(context.Background(), Options{
		StateFile: filepath.Join(t.TempDir(), "update.json"),
		Source: SourceFunc(func(context.Context) (Release, error) {
			return Release{}, nil
		}),
	})
	if err == nil || !strings.Contains(err.Error(), "empty version") {
		t.Fatalf("Check() error = %v, want empty version", err)
	}
}

func TestSourceFuncNil(t *testing.T) {
	t.Parallel()

	var source SourceFunc
	if _, err := source.Latest(context.Background()); err == nil {
		t.Fatal("Latest() error = nil, want error")
	}
}

func TestNotify(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	if err := Notify(&b, Status{AppName: "demo", Current: "1.0.0", Latest: "1.1.0", URL: "https://example.test", Available: true}); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if got := b.String(); !strings.Contains(got, "demo 1.1.0") || !strings.Contains(got, "https://example.test") {
		t.Fatalf("notice = %q", got)
	}
}

func TestNotifyNoUpdate(t *testing.T) {
	t.Parallel()

	var b strings.Builder
	if err := Notify(&b, Status{AppName: "demo", Current: "1.1.0", Latest: "1.1.0"}); err != nil {
		t.Fatalf("Notify() error = %v", err)
	}
	if b.Len() != 0 {
		t.Fatalf("notice = %q, want empty", b.String())
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

func TestDefaultStateFileUsesCacheDir(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("LOCALAPPDATA", dir)

	got := DefaultStateFile("demo")
	if got == "" || !strings.Contains(got, "demo") || !strings.HasSuffix(got, "update-check.json") {
		t.Fatalf("DefaultStateFile() = %q", got)
	}
	if _, err := os.Stat(filepath.Dir(got)); err == nil {
		t.Fatal("cache directory should not be created until a check succeeds")
	}
}
