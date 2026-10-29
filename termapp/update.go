package termapp

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/CHE3MZ/terminull/v2/termupdate"
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
type UpdateStatus = termupdate.Status

// CheckUpdate checks GitHub for the latest release, throttled by a state
// file. It is a compatibility wrapper around termupdate.Check. Nothing
// calls it automatically.
func CheckUpdate(ctx context.Context, opts UpdateOptions) (UpdateStatus, error) {
	source, err := termupdate.NewGitHub(termupdate.GitHubOptions{
		Repo:    opts.Repo,
		BaseURL: opts.BaseURL,
		Client:  opts.Client,
	})
	if err != nil {
		app := opts.AppName
		if app == "" {
			app = "terminull"
		}
		return UpdateStatus{AppName: app, Current: opts.CurrentVersion}, err
	}
	return termupdate.Check(ctx, termupdate.Options{
		AppName:        opts.AppName,
		CurrentVersion: opts.CurrentVersion,
		StateFile:      opts.StateFile,
		Interval:       opts.Interval,
		Source:         source,
	})
}

// NotifyUpdate writes a short notice when a newer release exists.
func NotifyUpdate(w io.Writer, status UpdateStatus) error {
	return termupdate.Notify(w, status)
}
