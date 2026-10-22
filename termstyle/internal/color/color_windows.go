//go:build windows

package color

import "github.com/CHE3MZ/terminull/v2/termcore"

func init() {
	_ = termcore.EnsureConsole()
}
