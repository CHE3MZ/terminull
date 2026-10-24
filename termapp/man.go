package termapp

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/CHE3MZ/terminull/v2/termapp/internal/cli"
	"github.com/CHE3MZ/terminull/v2/termapp/internal/cli/pflag"
)

// ManPages renders one roff man page per visible command, keyed by the
// dash-joined command path (for example "terminull-new"). Hidden
// commands are omitted. Global and inherited flags are included on
// every page.
func (a *App) ManPages() (map[string][]byte, error) {
	pages := make(map[string][]byte)
	var walk func(cmd *cli.Command) error
	walk = func(cmd *cli.Command) error {
		if !cmd.Hidden {
			data, err := a.manPage(cmd)
			if err != nil {
				return err
			}
			name := strings.ReplaceAll(cmd.CommandPath(), " ", "-")
			pages[name] = data
		}
		for _, child := range cmd.Commands() {
			if err := walk(child); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(a.cmd); err != nil {
		return nil, err
	}
	return pages, nil
}

func (a *App) manPage(cmd *cli.Command) ([]byte, error) {
	var b bytes.Buffer
	title := strings.ToUpper(strings.ReplaceAll(cmd.CommandPath(), " ", "-"))
	date := time.Now().Format("January 2006")
	fmt.Fprintf(&b, ".TH %q 1 %q %q %q\n", title, date, a.version, a.name)
	b.WriteString(".SH NAME\n")
	fmt.Fprintf(&b, "%s \\- %s\n", roffText(cmd.CommandPath()), roffText(cmd.Short))
	b.WriteString(".SH SYNOPSIS\n")
	fmt.Fprintf(&b, ".B %s\n", roffText(cmd.CommandPath()))
	if synopsis := commandSynopsis(cmd); synopsis != "" {
		fmt.Fprintf(&b, "%s\n", roffText(synopsis))
	}

	b.WriteString(".SH DESCRIPTION\n")
	desc := cmd.Long
	if desc == "" {
		desc = cmd.Short
	}
	writeRoffParagraph(&b, desc)

	if children := visibleChildren(cmd); len(children) > 0 {
		b.WriteString(".SH COMMANDS\n")
		for _, child := range children {
			b.WriteString(".TP\n")
			fmt.Fprintf(&b, ".B %s\n", roffText(child.Name()))
			fmt.Fprintf(&b, "%s\n", roffText(child.Short))
		}
	}

	if flags := pageFlags(cmd); len(flags) > 0 {
		b.WriteString(".SH OPTIONS\n")
		for _, flag := range flags {
			b.WriteString(".TP\n")
			fmt.Fprintf(&b, ".B %s\n", roffFlag(flag))
			fmt.Fprintf(&b, "%s", roffText(flag.Usage))
			if flag.DefValue != "" && flag.DefValue != "false" && flag.DefValue != "0" {
				fmt.Fprintf(&b, " (default: %s)", roffText(flag.DefValue))
			}
			b.WriteByte('\n')
		}
	}

	if cmd.Example != "" {
		b.WriteString(".SH EXAMPLES\n")
		writeRoffParagraph(&b, cmd.Example)
	}

	b.WriteString(".SH SEE ALSO\n")
	fmt.Fprintf(&b, ".B %s(1)\n", roffText(a.name))

	return b.Bytes(), nil
}

func commandSynopsis(cmd *cli.Command) string {
	line := cmd.UseLine()
	path := cmd.CommandPath()
	if line == path {
		return ""
	}
	if strings.HasPrefix(line, path) {
		return strings.TrimSpace(strings.TrimPrefix(line, path))
	}
	if strings.HasPrefix(line, cmd.Name()) {
		return strings.TrimSpace(strings.TrimPrefix(line, cmd.Name()))
	}
	return ""
}

func visibleChildren(cmd *cli.Command) []*cli.Command {
	var out []*cli.Command
	for _, child := range cmd.Commands() {
		if !child.Hidden && child.IsAvailableCommand() {
			out = append(out, child)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name() < out[j].Name() })
	return out
}

func pageFlags(cmd *cli.Command) []*pflag.Flag {
	seen := make(map[string]bool)
	var flags []*pflag.Flag
	collect := func(fs *pflag.FlagSet) {
		if fs == nil {
			return
		}
		fs.VisitAll(func(flag *pflag.Flag) {
			if flag.Hidden || seen[flag.Name] {
				return
			}
			seen[flag.Name] = true
			flags = append(flags, flag)
		})
	}
	collect(cmd.Flags())
	collect(cmd.InheritedFlags())
	collect(cmd.PersistentFlags())
	sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })
	return flags
}

func roffFlag(flag *pflag.Flag) string {
	name := "\\-\\-" + roffText(flag.Name)
	if flag.Shorthand != "" {
		name = "\\-" + roffText(flag.Shorthand) + ", " + name
	}
	if flag.Value.Type() != "bool" {
		name += " \\fI" + roffText(flag.Value.Type()) + "\\fR"
	}
	return name
}

func writeRoffParagraph(b *bytes.Buffer, text string) {
	for _, line := range strings.Split(strings.TrimSpace(text), "\n") {
		b.WriteString(roffText(line))
		b.WriteByte('\n')
	}
}

func roffText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "-", "\\-")
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, ".") || strings.HasPrefix(line, "'") {
			lines[i] = "\\&" + line
		}
	}
	return strings.Join(lines, "\n")
}
