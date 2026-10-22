// Command termtable-demo renders the same table in every termtable style.
package main

import (
	"fmt"
	"os"

	"github.com/CHE3MZ/terminull/v2/termtable"
)

func main() {
	header := []string{"NODE", "STATUS", "LATENCY"}
	rows := [][]string{
		{"web-1", "running", "18ms"},
		{"web-2", "running", "22ms"},
		{"db-1", "draining", "61ms"},
	}
	styles := []termtable.Style{
		termtable.StyleBordered,
		termtable.StyleRounded,
		termtable.StyleASCII,
		termtable.StyleMarkdown,
		termtable.StyleBorderless,
	}

	for i, style := range styles {
		if i > 0 {
			fmt.Println()
		}
		fmt.Printf("%s:\n", style)

		tb := termtable.New(os.Stdout, termtable.WithStyle(style))
		tb.SetHeaders(header...)
		tb.AddRows(rows...)
		if err := tb.Render(); err != nil {
			fmt.Fprintln(os.Stderr, "termtable demo:", err)
			os.Exit(1)
		}
	}
}
