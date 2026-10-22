//go:build windows

package color

import "github.com/CHE3MZ/terminull/termcore"

func init() {
	_ = termcore.EnsureConsole()
}
