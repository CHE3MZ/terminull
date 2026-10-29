package termupdate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubLatest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/tool/releases/latest" {
			t.Errorf("path = %q, want release endpoint", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("Accept = %q", r.Header.Get("Accept"))
		}
		if r.Header.Get("User-Agent") != "terminull-update-check" {
			t.Errorf("User-Agent = %q", r.Header.Get("User-Agent"))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"tag_name": "v1.2.4",
			"html_url": "https://example.test/releases/v1.2.4",
		})
	}))
	defer srv.Close()

	source, err := NewGitHub(GitHubOptions{
		Repo:    "acme/tool",
		BaseURL: srv.URL + "/",
		Client:  srv.Client(),
	})
	if err != nil {
		t.Fatalf("NewGitHub() error = %v", err)
	}
	release, err := source.Latest(context.Background())
	if err != nil {
		t.Fatalf("Latest() error = %v", err)
	}
	if release.Version != "v1.2.4" || release.URL != "https://example.test/releases/v1.2.4" {
		t.Fatalf("release = %+v", release)
	}
}

func TestGitHubHTTPError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusForbidden)
	}))
	defer srv.Close()

	source, err := NewGitHub(GitHubOptions{Repo: "acme/tool", BaseURL: srv.URL, Client: srv.Client()})
	if err != nil {
		t.Fatalf("NewGitHub() error = %v", err)
	}
	_, err = source.Latest(context.Background())
	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("Latest() error = %v, want 403", err)
	}
}

func TestGitHubMissingTag(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"html_url": "https://example.test"})
	}))
	defer srv.Close()

	source, err := NewGitHub(GitHubOptions{Repo: "acme/tool", BaseURL: srv.URL, Client: srv.Client()})
	if err != nil {
		t.Fatalf("NewGitHub() error = %v", err)
	}
	_, err = source.Latest(context.Background())
	if err == nil || !strings.Contains(err.Error(), "no tag") {
		t.Fatalf("Latest() error = %v, want no tag", err)
	}
}

func TestNewGitHubValidation(t *testing.T) {
	t.Parallel()

	for _, repo := range []string{"", "acme", "acme/", "/tool", "acme/tool/extra", "acme/t ool", "acme/tool?x"} {
		if _, err := NewGitHub(GitHubOptions{Repo: repo}); err == nil {
			t.Errorf("NewGitHub(%q) error = nil, want validation error", repo)
		}
	}

	if _, err := NewGitHub(GitHubOptions{Repo: " acme/tool "}); err != nil {
		t.Fatalf("NewGitHub(trimmed) error = %v", err)
	}
}
