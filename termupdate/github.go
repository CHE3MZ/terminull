package termupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubOptions tunes a GitHub source. Repo is "owner/name"; BaseURL and
// Client are optional overrides for GitHub-compatible endpoints and tests.
type GitHubOptions struct {
	Repo    string
	BaseURL string
	Client  *http.Client
}

// GitHub checks a GitHub-compatible releases/latest endpoint.
type GitHub struct {
	repo    string
	baseURL string
	client  *http.Client
}

var _ Source = (*GitHub)(nil)

// NewGitHub validates opts and returns a GitHub source.
func NewGitHub(opts GitHubOptions) (*GitHub, error) {
	repo := strings.TrimSpace(opts.Repo)
	owner, name, ok := strings.Cut(repo, "/")
	if !ok || owner == "" || name == "" || strings.Contains(name, "/") ||
		strings.ContainsAny(repo, " \t\r\n?#") {
		return nil, fmt.Errorf("termupdate: repo must be owner/name")
	}

	client := opts.Client
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &GitHub{
		repo:    repo,
		baseURL: strings.TrimRight(opts.BaseURL, "/"),
		client:  client,
	}, nil
}

// Latest asks the GitHub-compatible API for the latest release.
func (g *GitHub) Latest(ctx context.Context) (Release, error) {
	base := g.baseURL
	if base == "" {
		base = "https://api.github.com"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/repos/"+g.repo+"/releases/latest", nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "terminull-update-check")

	resp, err := g.client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return Release{}, fmt.Errorf("termupdate: GitHub returned %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&release); err != nil {
		return Release{}, err
	}
	if release.TagName == "" {
		return Release{}, fmt.Errorf("termupdate: latest release has no tag")
	}
	return Release{Version: release.TagName, URL: release.HTMLURL}, nil
}
