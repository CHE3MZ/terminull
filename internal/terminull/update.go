package terminull

import (
	"fmt"

	"github.com/CHE3MZ/terminull/v2/termapp"
	"github.com/CHE3MZ/terminull/v2/termcore"
	"github.com/CHE3MZ/terminull/v2/termupdate"
)

func updateCommand() termapp.Command {
	return termapp.Command{
		Use:    "update-check",
		Short:  "check for a newer terminull release",
		Long:   "Check the latest terminull release. Results are cached for a day.",
		Hidden: true,
		Args:   termapp.NoArgs,
		Flags: func(fs *termapp.FlagSet) {
			fs.String("repo", "CHE3MZ/terminull", "GitHub repository owner/name")
			fs.String("state", "", "update state file (default: user cache dir)")
		},
		Run: func(ctx *termapp.Context) error {
			repo, err := ctx.String("repo")
			if err != nil {
				return err
			}
			state, err := ctx.String("state")
			if err != nil {
				return err
			}
			source, err := termupdate.NewGitHub(termupdate.GitHubOptions{Repo: repo})
			if err != nil {
				return err
			}
			status, err := termupdate.Check(ctx, termupdate.Options{
				AppName:        "terminull",
				CurrentVersion: termcore.Version,
				StateFile:      state,
				Source:         source,
			})
			if err != nil {
				return err
			}
			if status.Available {
				return termupdate.Notify(ctx.Out, status)
			}
			if status.Current == "" || status.Current == "unknown" {
				_, err = fmt.Fprintln(ctx.Out, "terminull version unknown; update check skipped")
				return err
			}
			_, err = fmt.Fprintf(ctx.Out, "terminull %s is up to date\n", status.Current)
			return err
		},
	}
}
