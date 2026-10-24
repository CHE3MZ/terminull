package terminull

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/CHE3MZ/terminull/v2/termapp"
)

func manCommand(app *termapp.App) termapp.Command {
	return termapp.Command{
		Use:    "man",
		Short:  "write man pages",
		Long:   "Write one roff man page per visible command into DIR.",
		Hidden: true,
		Args:   termapp.NoArgs,
		Flags: func(fs *termapp.FlagSet) {
			fs.String("dir", "man", "output directory")
		},
		Run: func(ctx *termapp.Context) error {
			ctx.StopDrawing()
			dir, err := ctx.String("dir")
			if err != nil {
				return err
			}
			pages, err := app.ManPages()
			if err != nil {
				return err
			}
			if err := os.MkdirAll(dir, 0o750); err != nil {
				return err
			}
			names := make([]string, 0, len(pages))
			for name := range pages {
				names = append(names, name)
			}
			sort.Strings(names)
			for _, name := range names {
				path := filepath.Join(dir, name+".1")
				if err := os.WriteFile(path, pages[name], 0o600); err != nil {
					return err
				}
			}
			_, err = fmt.Fprintf(ctx.Out, "wrote %d man pages to %s\n", len(names), dir)
			return err
		},
	}
}
