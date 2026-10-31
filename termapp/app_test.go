package termapp

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/CHE3MZ/terminull/v2/termcore"
	"github.com/CHE3MZ/terminull/v2/termrender"
	"github.com/CHE3MZ/terminull/v2/termrender/renderers"
	rendercfg "github.com/CHE3MZ/terminull/v2/termrender/renderers/config"
	"github.com/CHE3MZ/terminull/v2/termrender/terminal"
)

// testDeployRun records what the handler saw.
type testDeployRun struct {
	flagEnv   string
	configEnv string
	output    string
	ran       bool
}

func newTestApp(out, errOut *bytes.Buffer, run *testDeployRun, opts ...Option) *App {
	app := New("demo", append([]Option{WithVersion("9.9.9-test")}, opts...)...)
	app.SetOut(out)
	app.SetErr(errOut)
	app.Add(Command{
		Use:   "deploy",
		Short: "deploy things",
		Args:  MaximumNArgs(1),
		Flags: func(fs *FlagSet) {
			fs.String("env", "dev", "target environment")
		},
		Run: func(ctx *Context) error {
			run.ran = true
			flagEnv, err := ctx.String("env")
			if err != nil {
				return err
			}
			run.flagEnv = flagEnv
			run.configEnv = ctx.Config.String("env")
			run.output = ctx.Output
			return nil
		},
	})
	return app
}

func combined(out, errOut *bytes.Buffer) string {
	return out.String() + errOut.String()
}

func TestHelpIsOutOnly(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)

	if err := app.Execute(context.Background(), []string{"--help"}); err != nil {
		t.Fatalf("Execute(--help) error = %v", err)
	}
	if run.ran {
		t.Fatal("help must not run any command")
	}

	got := combined(&out, &errOut)
	for _, want := range []string{"demo", "deploy", "deploy things", "--log-level", "--log-style"} {
		if !strings.Contains(got, want) {
			t.Fatalf("help %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "\x1b[") {
		t.Fatalf("piped help %q contains ANSI escapes, want plain", got)
	}
}

func TestVersionIsOutOnly(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)

	if err := app.Execute(context.Background(), []string{"--version"}); err != nil {
		t.Fatalf("Execute(--version) error = %v", err)
	}

	if got := combined(&out, &errOut); !strings.Contains(got, "9.9.9-test") {
		t.Fatalf("version output %q does not contain version", got)
	}
}

func TestCompletionIsOutOnly(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)

	if err := app.Execute(context.Background(), []string{"completion", "bash"}); err != nil {
		t.Fatalf("Execute(completion bash) error = %v", err)
	}

	got := combined(&out, &errOut)
	if len(got) == 0 || !strings.Contains(got, "demo") {
		t.Fatalf("completion output %q is empty or missing app name", got)
	}
}

func TestCompletionCommandHiddenByDefault(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)

	if err := app.Execute(context.Background(), []string{"--help"}); err != nil {
		t.Fatalf("Execute(--help) error = %v", err)
	}
	got := combined(&out, &errOut)
	if strings.Contains(got, "completion") {
		t.Fatalf("help %q lists completion by default", got)
	}
	if !strings.Contains(got, "deploy") {
		t.Fatalf("help %q does not list deploy", got)
	}
}

func TestVisibleCompletionCommand(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run, WithVisibleCompletionCommand())

	if err := app.Execute(context.Background(), []string{"--help"}); err != nil {
		t.Fatalf("Execute(--help) error = %v", err)
	}
	if got := combined(&out, &errOut); !strings.Contains(got, "completion") {
		t.Fatalf("help %q does not list completion", got)
	}

	out.Reset()
	errOut.Reset()
	if err := app.Execute(context.Background(), []string{"completion", "bash"}); err != nil {
		t.Fatalf("Execute(completion bash) error = %v", err)
	}
	if got := combined(&out, &errOut); !strings.Contains(got, "demo") {
		t.Fatalf("completion output %q is empty or missing app name", got)
	}
}

func TestUnknownFlagExits2(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)

	err := app.Execute(context.Background(), []string{"deploy", "--bogus"})
	if err == nil {
		t.Fatal("Execute with unknown flag = nil, want error")
	}
	if run.ran {
		t.Fatal("handler must not run on parse failure")
	}
	if got := termcore.ExitCodeOf(err); got != 2 {
		t.Fatalf("ExitCodeOf = %d, want 2 for usage errors", got)
	}
}

func TestTooManyArgsExits2(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)

	err := app.Execute(context.Background(), []string{"deploy", "a", "b"})
	if err == nil {
		t.Fatal("Execute over arg limit = nil, want error")
	}
	if got := termcore.ExitCodeOf(err); got != 2 {
		t.Fatalf("ExitCodeOf = %d, want 2 for usage errors", got)
	}
}

func TestInvalidOutputExits2(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)

	err := app.Execute(context.Background(), []string{"--output", "xml", "deploy"})
	if err == nil {
		t.Fatal("Execute with bad --output = nil, want error")
	}
	if got := termcore.ExitCodeOf(err); got != 2 {
		t.Fatalf("ExitCodeOf = %d, want 2 for usage errors", got)
	}
}

func TestEnvBeatsDefault(t *testing.T) {
	t.Setenv("DEMO_ENV", "from-env")

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)

	if err := app.Execute(context.Background(), []string{"deploy"}); err != nil {
		t.Fatalf("Execute(deploy) error = %v", err)
	}
	if run.flagEnv != "dev" {
		t.Fatalf("flag env = %q, want default dev", run.flagEnv)
	}
	if run.configEnv != "from-env" {
		t.Fatalf("config env = %q, want env to beat default", run.configEnv)
	}
}

func TestFlagBeatsEnv(t *testing.T) {
	t.Setenv("DEMO_ENV", "from-env")

	var out, errOut bytes.Buffer
	var run testDeployRun
	app := newTestApp(&out, &errOut, &run)

	if err := app.Execute(context.Background(), []string{"deploy", "--env", "from-flag"}); err != nil {
		t.Fatalf("Execute(deploy --env) error = %v", err)
	}
	if run.flagEnv != "from-flag" || run.configEnv != "from-flag" {
		t.Fatalf("flag=%q config=%q, want from-flag to win", run.flagEnv, run.configEnv)
	}
}

func TestNestedChildren(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	ran := map[string]bool{}

	app := New("demo", WithVersion("1-test"))
	app.SetOut(&out)
	app.SetErr(&errOut)
	app.Add(Command{
		Use:   "worker",
		Short: "manage workers",
		Children: []Command{
			{
				Use:   "run",
				Short: "run a worker",
				Run: func(_ *Context) error {
					ran["run"] = true
					return nil
				},
			},
			{
				Use:   "pause",
				Short: "pause a worker",
				Run: func(_ *Context) error {
					ran["pause"] = true
					return nil
				},
			},
		},
	})

	if err := app.Execute(context.Background(), []string{"worker", "pause"}); err != nil {
		t.Fatalf("Execute(worker pause) error = %v", err)
	}
	if !ran["pause"] || ran["run"] {
		t.Fatalf("ran = %v, want only pause", ran)
	}

	// Bare group prints help.
	out.Reset()
	errOut.Reset()
	if err := app.Execute(context.Background(), []string{"worker"}); err != nil {
		t.Fatalf("Execute(worker) error = %v", err)
	}
	if got := combined(&out, &errOut); !strings.Contains(got, "run") || !strings.Contains(got, "pause") {
		t.Fatalf("group help %q does not list children", got)
	}
}

func TestPauseResumeSimpleNoop(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	app := New("demo", WithVersion("1-test"))
	app.SetOut(&out)
	app.SetErr(&errOut)
	app.Add(Command{
		Use:   "nop",
		Short: "pause on simple renderer",
		Run: func(ctx *Context) error {
			ctx.PauseDrawing()
			defer ctx.ResumeDrawing()
			return nil
		},
	})

	if err := app.Execute(context.Background(), []string{"nop"}); err != nil {
		t.Fatalf("Execute(nop) error = %v", err)
	}
}

func TestPauseResumeInteractive(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "frame.txt")
	// #nosec G304 -- test frame file under t.TempDir().
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	renderer := renderers.NewInteractiveRenderer(f, rendercfg.NewDefaultSymbolsOnlyRenderingConfig())
	ctx := &Context{Logger: termrender.NewLogger(termrender.InfoLevel, renderer)}

	size := func() int64 {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("Stat() error = %v", err)
		}
		return info.Size()
	}

	ctx.Logger.Scoped("a").Infof("x")

	// The logger drains on its own goroutine: poll for the first frame.
	var grown int64
	deadline := time.Now().Add(5 * time.Second)
	for {
		renderer.DrawFrame()
		if grown = size(); grown > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("DrawFrame wrote nothing with an open scope")
		}
		time.Sleep(5 * time.Millisecond)
	}

	ctx.PauseDrawing()
	paused := size()
	if paused <= grown {
		t.Fatal("PauseDrawing did not restore autowrap")
	}
	ctx.Logger.Scoped("b").Infof("y")
	renderer.DrawFrame()
	if got := size(); got != paused {
		t.Fatalf("paused DrawFrame changed %d to %d", paused, got)
	}

	ctx.ResumeDrawing()
	resumed := size()
	if resumed <= paused {
		t.Fatal("ResumeDrawing did not restore no-wrap drawing")
	}
	renderer.DrawFrame()
	if got := size(); got <= resumed {
		t.Fatal("resumed DrawFrame did not grow the frame")
	}
}

func TestChangedTracksExplicitFlags(t *testing.T) {
	t.Setenv("DEMO_ENV", "from-env")

	var out, errOut bytes.Buffer
	seen := map[string]bool{}
	app := New("demo", WithVersion("1-test"))
	app.SetOut(&out)
	app.SetErr(&errOut)
	app.Add(Command{
		Use: "deploy",
		Flags: func(fs *FlagSet) {
			fs.String("env", "dev", "target environment")
		},
		Run: func(ctx *Context) error {
			seen["env"] = ctx.Changed("env")
			seen["missing"] = ctx.Changed("missing")
			return nil
		},
	})

	if err := app.Execute(context.Background(), []string{"deploy", "--env", "x"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !seen["env"] || seen["missing"] {
		t.Fatalf("Changed = %v, want env true and missing false", seen)
	}
}

func TestStaticArgCompletion(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	app := New("demo", WithVersion("1-test"))
	app.SetOut(&out)
	app.SetErr(&errOut)
	app.Add(Command{
		Use:       "deploy",
		ValidArgs: []string{"prod", "staging"},
		Run:       func(*Context) error { return nil },
	})

	if err := app.Execute(context.Background(), []string{"__complete", "deploy", ""}); err != nil {
		t.Fatalf("Execute(__complete) error = %v", err)
	}
	got := combined(&out, &errOut)
	for _, want := range []string{"prod", "staging"} {
		if !strings.Contains(got, want) {
			t.Fatalf("completion %q does not contain %q", got, want)
		}
	}
}

func TestHiddenCommandAbsentFromHelp(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	app := New("demo", WithVersion("1-test"))
	app.SetOut(&out)
	app.SetErr(&errOut)
	app.Add(Command{Use: "visible", Short: "shown", Run: func(*Context) error { return nil }})
	app.Add(Command{Use: "secret", Short: "hidden", Hidden: true, Run: func(*Context) error { return nil }})

	if err := app.Execute(context.Background(), []string{"--help"}); err != nil {
		t.Fatalf("Execute(--help) error = %v", err)
	}
	got := combined(&out, &errOut)
	if !strings.Contains(got, "visible") {
		t.Fatalf("help %q misses visible command", got)
	}
	if strings.Contains(got, "secret") {
		t.Fatalf("help %q lists hidden command", got)
	}

	out.Reset()
	errOut.Reset()
	if err := app.Execute(context.Background(), []string{"secret"}); err != nil {
		t.Fatalf("Execute(secret) error = %v, hidden must still run", err)
	}
}

func TestDynamicArgCompletion(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	app := New("demo", WithVersion("1-test"))
	app.SetOut(&out)
	app.SetErr(&errOut)
	app.Add(Command{
		Use: "deploy",
		ValidArgsFunction: func(_ []string, toComplete string) ([]string, ShellCompDirective) {
			var matches []string
			for _, env := range []string{"prod", "staging", "dev"} {
				if strings.HasPrefix(env, toComplete) {
					matches = append(matches, env)
				}
			}
			return matches, ShellCompDirectiveNoFileComp
		},
		Run: func(*Context) error { return nil },
	})

	if err := app.Execute(context.Background(), []string{"__complete", "deploy", "p"}); err != nil {
		t.Fatalf("Execute(__complete) error = %v", err)
	}
	got := combined(&out, &errOut)
	if !strings.Contains(got, "prod") {
		t.Fatalf("completion %q does not contain filtered match", got)
	}
	if strings.Contains(got, "staging") || strings.Contains(got, "\ndev\n") {
		t.Fatalf("completion %q leaks unfiltered entries", got)
	}
}

func TestWithSymbolsOnlySetsField(t *testing.T) {
	t.Parallel()

	if got := New("demo", WithSymbolsOnly()); !got.symbolsOnly {
		t.Fatal("WithSymbolsOnly did not set the flag")
	}
	if got := New("demo"); got.symbolsOnly {
		t.Fatal("default app forces symbols")
	}
}

func TestContextEmojiHonorsSymbolsOnly(t *testing.T) {
	t.Setenv("TERM", "xterm")
	t.Setenv("CI", "")
	t.Setenv("WT_SESSION", "test")

	run := func(opts ...Option) bool {
		var got bool
		var out, errOut bytes.Buffer
		app := New("demo", append([]Option{WithVersion("1-test")}, opts...)...)
		app.SetOut(&out)
		app.SetErr(&errOut)
		app.Add(Command{
			Use: "emoji",
			Run: func(ctx *Context) error {
				got = ctx.Emoji
				return nil
			},
		})
		if err := app.Execute(context.Background(), []string{"emoji"}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		return got
	}

	if !run() {
		t.Fatal("default Context.Emoji is false on an emoji-capable terminal")
	}
	if run(WithSymbolsOnly()) {
		t.Fatal("WithSymbolsOnly did not disable Context.Emoji")
	}
}

func TestWithLogStyleColorsSetsField(t *testing.T) {
	t.Parallel()

	if got := New("demo"); got.styleColors == nil {
		t.Fatal("default app has no flat prefix colors")
	}

	colors := terminal.DefaultFlatStyleColors()
	colors.WarnColor = terminal.BlueColor
	if got := New("demo", WithLogStyleColors(colors)); got.styleColors != colors {
		t.Fatal("WithLogStyleColors did not use the supplied palette")
	}
	if got := New("demo", WithLogStyleColors(nil)); got.styleColors != nil {
		t.Fatal("WithLogStyleColors(nil) did not disable prefix colors")
	}
}

func TestSymbolsOnlyMatchesDefaultOffTerminal(t *testing.T) {
	t.Parallel()

	run := func(opts ...Option) string {
		var out, errOut bytes.Buffer
		app := New("demo", append([]Option{WithVersion("1-test")}, opts...)...)
		app.SetOut(&out)
		app.SetErr(&errOut)
		app.Add(Command{
			Use:   "deploy",
			Short: "deploy things",
			Run:   func(*Context) error { return nil },
		})
		if err := app.Execute(context.Background(), []string{"deploy"}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		return combined(&out, &errOut)
	}

	// Off-terminal both take the simple branch: identical bytes prove the
	// flag changes TTY glyphs only, never piped output.
	if plain, symbols := run(), run(WithSymbolsOnly()); plain != symbols {
		t.Fatalf("outputs differ off-terminal:\n%q\nvs\n%q", plain, symbols)
	}
}

func TestLogStyleFlag(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	app := New("demo", WithVersion("1-test"))
	app.SetOut(&out)
	app.SetErr(&errOut)
	app.Add(Command{
		Use: "log",
		Run: func(ctx *Context) error {
			ctx.Logger.Infof("starting")
			return nil
		},
	})

	if err := app.Execute(context.Background(), []string{"--log-style", "logrus", "log"}); err != nil {
		t.Fatalf("Execute(--log-style logrus) error = %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want logs on stderr", out.String())
	}
	got := errOut.String()
	if !strings.HasPrefix(got, "INFO[") || !strings.HasSuffix(got, "] starting\n") {
		t.Fatalf("logrus output = %q, want INFO[0000] starting", got)
	}
}

func TestWithLogStyleDefault(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	app := New("demo", WithVersion("1-test"), WithLogStyle(termrender.LogStyleTime))
	app.SetOut(&out)
	app.SetErr(&errOut)
	app.Add(Command{
		Use: "log",
		Run: func(ctx *Context) error {
			ctx.Logger.Infof("starting")
			return nil
		},
	})

	if err := app.Execute(context.Background(), []string{"log"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if out.Len() != 0 {
		t.Fatalf("stdout = %q, want logs on stderr", out.String())
	}
	got := errOut.String()
	if !strings.HasPrefix(got, "[") || !strings.HasSuffix(got, "] starting\n") {
		t.Fatalf("time output = %q, want [15:04:05] starting", got)
	}
}

func TestInvalidLogStyleExits2(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	ran := false
	app := New("demo", WithVersion("1-test"))
	app.SetOut(&out)
	app.SetErr(&errOut)
	app.Add(Command{
		Use: "log",
		Run: func(*Context) error {
			ran = true
			return nil
		},
	})

	err := app.Execute(context.Background(), []string{"--log-style", "rainbow", "log"})
	if err == nil {
		t.Fatal("Execute with bad --log-style = nil, want error")
	}
	if ran {
		t.Fatal("handler must not run on bad --log-style")
	}
	if got := termcore.ExitCodeOf(err); got != 2 {
		t.Fatalf("ExitCodeOf = %d, want 2 for usage errors", got)
	}
}
