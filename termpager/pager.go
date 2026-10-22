package termpager

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/CHE3MZ/terminull/v2/termcore"
)

type options struct {
	pager string
}

// Option tunes Start.
type Option func(*options)

// WithPager forces a pager command line, e.g. "less -R" or "more".
// Default: $PAGER, then less.
func WithPager(cmd string) Option {
	return func(o *options) {
		o.pager = cmd
	}
}

// Start launches the pager for w when appropriate and returns the writer
// commands should use plus a stop func (close the feed, wait for exit).
// Paging needs a terminal file, a non-dumb TERM, and a found pager
// binary; otherwise it returns w and a no-op stop with no error.
func Start(w io.Writer, opts ...Option) (io.Writer, func(), error) {
	noop := func() {}

	var cfg options
	for _, opt := range opts {
		opt(&cfg)
	}

	f, ok := w.(*os.File)
	if !ok || !termcore.IsTerminal(f) {
		return w, noop, nil
	}

	prog, args, ok := resolvePager(cfg.pager, os.Getenv("PAGER"), os.Getenv("TERM"), exec.LookPath)
	if !ok {
		return w, noop, nil
	}
	return spawn(prog, args, f, os.Environ())
}

// resolvePager picks the pager binary: explicit flag, $PAGER, then less.
// Dumb terminals, cat, unparseable lines, and missing binaries decline.
func resolvePager(pagerFlag, envPager, term string, lookPath func(string) (string, error)) (string, []string, bool) {
	if strings.EqualFold(term, "dumb") {
		return "", nil, false
	}

	want := pagerFlag
	if want == "" {
		want = envPager
	}

	var prog string
	var args []string
	if want == "" {
		prog = "less"
	} else {
		parts, err := termcore.ShellSplit(want)
		if err != nil || len(parts) == 0 {
			return "", nil, false
		}
		prog, args = parts[0], parts[1:]
	}

	if base := strings.ToLower(filepath.Base(prog)); base == "cat" || base == "cat.exe" {
		return "", nil, false
	}
	resolved, err := lookPath(prog)
	if err != nil {
		return "", nil, false
	}
	return resolved, args, true
}

// spawn runs the pager reading from a pipe we feed and writing to the
// terminal file. LESS defaults to FRX (quit-if-fits, raw, no init)
// unless the caller already sets it.
func spawn(prog string, args []string, stdout *os.File, baseEnv []string) (io.Writer, func(), error) {
	// #nosec G204,G702 -- pager binary is user-configured (PAGER flag/env),
	// the documented equivalent of typing `| $PAGER` in a shell.
	cmd := exec.Command(prog, args...)
	cmd.Stdout = stdout
	cmd.Stderr = os.Stderr
	cmd.Env = withLessDefault(baseEnv)

	feed, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}
	stop := func() {
		_ = feed.Close()
		_ = cmd.Wait()
	}
	return feed, stop, nil
}

func withLessDefault(env []string) []string {
	for _, kv := range env {
		if strings.HasPrefix(kv, "LESS=") {
			return env
		}
	}
	return append(env, "LESS=FRX")
}
