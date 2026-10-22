package termstatic

import (
	"fmt"
	"io"

	"github.com/CHE3MZ/terminull/v2/termstyle"
)

// Banner writes the app banner line: the name painted green when color
// is true, followed by the plain version. It writes synchronously and
// returns the first write error, if any.
func Banner(w io.Writer, name, version string, colorEnabled bool) error {
	line := fmt.Sprintf("%s %s\n", termstyle.NewScheme(colorEnabled).Green(name), version)
	_, err := io.WriteString(w, line)
	return err
}
