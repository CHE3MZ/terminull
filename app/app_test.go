package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/CHE3MZ/terminull/core"
	"github.com/CHE3MZ/terminull/render"
	"github.com/CHE3MZ/terminull/render/renderers"
	rendercfg "github.com/CHE3MZ/terminull/render/renderers/config"
)

// testDeployRun records what the handler saw.
type testDeployRun struct {
	flagEnv   string
	configEnv string
	output    string
	ran       bool
}

func newTestApp(out, errOut *bytes.Buffer, run *testDeployRun) *App {
	app := New("demo", WithVersion("9.9.9-test"))
	app.SetOut(out)
	app.SetErr(errOut)
	app.Add(Command{
		Use:   "deploy",
		Short: "deploy things",
		Args:  cobra.MaximumNArgs(1),
		Flags: func(fs *pflag.FlagSet) {
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
	for _, want := range []string{"demo", "deploy", "deploy things"} {
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
	if got := core.ExitCodeOf(err); got != 2 {
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
	if got := core.ExitCodeOf(err); got != 2 {
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
	if got := core.ExitCodeOf(err); got != 2 {
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
	ctx := &Context{Logger: render.NewLogger(render.InfoLevel, renderer)}

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
	ctx.Logger.Scoped("b").Infof("y")
	renderer.DrawFrame()
	if got := size(); got != grown {
		t.Fatalf("paused DrawFrame changed %d to %d", grown, got)
	}

	ctx.ResumeDrawing()
	renderer.DrawFrame()
	if got := size(); got <= grown {
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
		Flags: func(fs *pflag.FlagSet) {
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
