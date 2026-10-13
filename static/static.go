package static

import (
	"fmt"
	"io"

	"github.com/CHE3MZ/terminull/style"
	"github.com/fatih/color"
)

// VersionLine returns the stable one-line version string,
// e.g. "demo version 1.2.3-abc123".
func VersionLine(name, version string) string {
	return fmt.Sprintf("%s version %s", name, version)
}

// Banner writes the app banner line: the name painted green when color
// is true, followed by the plain version. It writes synchronously and
// returns the first write error, if any.
func Banner(w io.Writer, name, version string, colorEnabled bool) error {
	line := fmt.Sprintf("%s %s\n", style.Paint(colorEnabled, color.FgGreen, name), version)
	_, err := io.WriteString(w, line)
	return err
}
