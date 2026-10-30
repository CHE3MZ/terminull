// Package statuscard demonstrates a custom command and widget component
// built from public terminull packages.
package statuscard

import (
	"io"

	"github.com/CHE3MZ/terminull/v2/termapp"
	"github.com/CHE3MZ/terminull/v2/termbox"
	"github.com/CHE3MZ/terminull/v2/termoutput"
	"github.com/CHE3MZ/terminull/v2/termstyle"
)

type Card struct {
	Title   string
	Message string
	Width   int
}

func (c Card) Render(w io.Writer, scheme termstyle.Scheme) error {
	content := scheme.Bold(c.Title) + "\n" + scheme.Muted(c.Message)
	return termbox.Render(w, content, termbox.Options{
		Style:   termbox.StyleRounded,
		Title:   "custom component",
		Padding: 1,
		Width:   c.Width,
	}, scheme)
}

func Command() termapp.Command {
	return termapp.Command{
		Use:   "status",
		Short: "render a status card from a component package",
		Flags: func(fs *termapp.FlagSet) {
			fs.String("message", "all systems normal", "status message")
		},
		Run: func(ctx *termapp.Context) error {
			message, err := ctx.String("message")
			if err != nil {
				return err
			}

			card := Card{Title: "deployment", Message: message}
			ctx.StopDrawing()
			if ctx.Output != termoutput.FormatTable {
				return termoutput.Render(ctx.Out, ctx.Output, map[string]string{
					"title":   card.Title,
					"message": card.Message,
				})
			}

			scheme := termstyle.NewScheme(termstyle.ColorFor(ctx.Out, ctx.NoColor))
			return card.Render(ctx.Out, scheme)
		},
	}
}
