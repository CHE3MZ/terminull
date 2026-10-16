package style

import (
	"strings"

	"github.com/fatih/color"
)

// Scheme paints named styles, returning text unchanged when disabled.
// Build one per command from terminal detection:
//
//	scheme := style.NewScheme(style.ColorFor(out, noColor))
//
// Every method is safe on dumb terminals, pipes, and legacy consoles
// when disabled, and degrades to plain text rather than raw escapes.
type Scheme struct {
	Enabled bool
}

// NewScheme builds a Scheme with explicit enablement.
func NewScheme(enabled bool) Scheme {
	return Scheme{Enabled: enabled}
}

// Paint wraps s in attr when enabled, mirroring package Paint.
func (s Scheme) Paint(attr color.Attribute, str string) string {
	return Paint(s.Enabled, attr, str)
}

// Bold bolds text.
func (s Scheme) Bold(str string) string {
	return s.Paint(color.Bold, str)
}

// Muted dims secondary text in the current color via SGR faint, which is
// the only way to dim without knowing the terminal foreground. For a
// deterministic dim of a named color use Faint, for any RGB use FaintRGB.
func (s Scheme) Muted(str string) string {
	return s.Paint(color.Faint, str)
}

// Red paints errors.
func (s Scheme) Red(str string) string {
	return s.Paint(color.FgRed, str)
}

// Yellow paints warnings.
func (s Scheme) Yellow(str string) string {
	return s.Paint(color.FgYellow, str)
}

// Green paints success.
func (s Scheme) Green(str string) string {
	return s.Paint(color.FgGreen, str)
}

// Cyan paints highlights.
func (s Scheme) Cyan(str string) string {
	return s.Paint(color.FgCyan, str)
}

// SuccessIcon is a green check, plain otherwise.
func (s Scheme) SuccessIcon() string {
	return s.Paint(color.FgGreen, "✓")
}

// FailureIcon is a red X, plain otherwise.
func (s Scheme) FailureIcon() string {
	return s.Paint(color.FgRed, "X")
}

// WarningIcon is a yellow bang, plain otherwise.
func (s Scheme) WarningIcon() string {
	return s.Paint(color.FgYellow, "!")
}

// TableHeader uppercases and underlines a table header when enabled,
// returning it verbatim otherwise so piped output stays script-friendly.
func (s Scheme) TableHeader(str string) string {
	if !s.Enabled {
		return str
	}
	return s.Paint(color.Underline, strings.ToUpper(str))
}
