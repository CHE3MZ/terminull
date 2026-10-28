package termapp

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/v2/termcore"
	"github.com/CHE3MZ/terminull/v2/termrender"
	"github.com/CHE3MZ/terminull/v2/termrender/renderers"
	"github.com/CHE3MZ/terminull/v2/termrender/terminal"
)

func newLoggingApp(out, errOut *bytes.Buffer, run func(*Context) error, opts ...Option) *App {
	options := append([]Option{WithVersion("1-test")}, opts...)
	app := New("demo", options...)
	app.SetOut(out)
	app.SetErr(errOut)
	app.Add(Command{Use: "log", Run: run})
	return app
}

func TestLogsStayOffStdout(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	app := newLoggingApp(&out, &errOut, func(ctx *Context) error {
		ctx.Logger.Infof("starting")
		_, err := fmt.Fprintln(ctx.Out, `{"ok":true}`)
		return err
	})

	if err := app.Execute(context.Background(), []string{"--output", "json", "log"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if got, want := out.String(), "{\"ok\":true}\n"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
	if got := errOut.String(); !strings.Contains(got, "starting") {
		t.Fatalf("stderr = %q, want log line", got)
	}
}

func TestWithLogOutput(t *testing.T) {
	t.Parallel()

	var out, errOut, logOut bytes.Buffer
	app := newLoggingApp(&out, &errOut, func(ctx *Context) error {
		ctx.Logger.Infof("redirected")
		return nil
	}, WithLogOutput(&logOut))

	if err := app.Execute(context.Background(), []string{"log"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if out.Len() != 0 || errOut.Len() != 0 {
		t.Fatalf("stdout=%q stderr=%q, want logs only in WithLogOutput", out.String(), errOut.String())
	}
	if got := logOut.String(); !strings.Contains(got, "redirected") {
		t.Fatalf("log output = %q, want redirected message", got)
	}
}

func TestWithLogRendererFactory(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	factoryCalled := false
	stopCalled := false
	app := newLoggingApp(&out, &errOut, func(ctx *Context) error {
		task := ctx.Logger.Scoped("build")
		task.Infof("hello")
		task.Finish(true)
		return nil
	},
		WithLogStyle(termrender.LogStylePlain),
		WithLogRendererFactory(func(opts LogRendererOptions) (termrender.LogRendered, func()) {
			factoryCalled = true
			if opts.Out != &errOut || opts.File != nil {
				t.Errorf("factory output = %T/%v, want stderr buffer", opts.Out, opts.File)
			}
			if opts.Style != termrender.LogStylePlain {
				t.Errorf("factory style = %v, want plain", opts.Style)
			}
			simple := renderers.NewSimpleRenderer(opts.Out, terminal.NoColorSchema())
			return renderers.NewGithubActionsLogsRenderer(simple), func() { stopCalled = true }
		}),
	)

	if err := app.Execute(context.Background(), []string{"log"}); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !factoryCalled {
		t.Fatal("custom renderer factory was not called")
	}
	if !stopCalled {
		t.Fatal("custom renderer stop was not called")
	}
	got := errOut.String()
	for _, want := range []string{"##[group]build", "hello", "##[endgroup]", "succeeded"} {
		if !strings.Contains(got, want) {
			t.Fatalf("custom output %q does not contain %q", got, want)
		}
	}
}

func TestLogRendererFactoryNilReturnsError(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	ran := false
	app := newLoggingApp(&out, &errOut, func(*Context) error {
		ran = true
		return nil
	}, WithLogRendererFactory(func(LogRendererOptions) (termrender.LogRendered, func()) {
		return nil, nil
	}))

	err := app.Execute(context.Background(), []string{"log"})
	if err == nil {
		t.Fatal("Execute() = nil, want nil renderer error")
	}
	if ran {
		t.Fatal("handler ran with a nil renderer")
	}
	if got := termcore.ExitCodeOf(err); got != 2 {
		t.Fatalf("ExitCodeOf = %d, want 2", got)
	}
}

func TestLogLevelFlag(t *testing.T) {
	t.Parallel()

	logAll := func(ctx *Context) error {
		ctx.Logger.Errorf("error-line")
		ctx.Logger.Warnf("warn-line")
		ctx.Logger.Infof("info-line")
		ctx.Logger.Debugf("debug-line")
		ctx.Logger.Tracef("trace-line")
		return nil
	}

	tests := []struct {
		name string
		args []string
		opts []Option
		want []string
		drop []string
	}{
		{"default info", nil, nil, []string{"error-line", "warn-line", "info-line"}, []string{"debug-line", "trace-line"}},
		{"debug flag", []string{"--debug"}, nil, []string{"error-line", "warn-line", "info-line", "debug-line"}, []string{"trace-line"}},
		{"verbose flag", []string{"--verbose"}, nil, []string{"error-line", "warn-line", "info-line", "debug-line"}, []string{"trace-line"}},
		{"trace level", []string{"--log-level", "trace"}, nil, []string{"error-line", "warn-line", "info-line", "debug-line", "trace-line"}, nil},
		{"error level", []string{"--log-level", "error"}, nil, []string{"error-line"}, []string{"warn-line", "info-line", "debug-line", "trace-line"}},
		{"explicit wins", []string{"--debug", "--log-level", "error"}, nil, []string{"error-line"}, []string{"debug-line"}},
		{"option default", nil, []Option{WithLogLevel(termrender.ErrorLevel)}, []string{"error-line"}, []string{"warn-line", "info-line"}},
		{"flag beats option", []string{"--log-level", "warn"}, []Option{WithLogLevel(termrender.ErrorLevel)}, []string{"error-line", "warn-line"}, []string{"info-line"}},
		{"debug keeps trace", []string{"--debug"}, []Option{WithLogLevel(termrender.TraceLevel)}, []string{"error-line", "info-line", "debug-line", "trace-line"}, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out, errOut bytes.Buffer
			app := newLoggingApp(&out, &errOut, logAll, tt.opts...)
			args := append(append([]string{}, tt.args...), "log")
			if err := app.Execute(context.Background(), args); err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			got := errOut.String()
			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("output %q does not contain %q", got, want)
				}
			}
			for _, drop := range tt.drop {
				if strings.Contains(got, drop) {
					t.Errorf("output %q contains dropped %q", got, drop)
				}
			}
		})
	}
}

func TestInvalidLogLevelExits2(t *testing.T) {
	t.Parallel()

	var out, errOut bytes.Buffer
	ran := false
	app := newLoggingApp(&out, &errOut, func(*Context) error {
		ran = true
		return nil
	})

	err := app.Execute(context.Background(), []string{"--log-level", "loud", "log"})
	if err == nil {
		t.Fatal("Execute() = nil, want invalid level error")
	}
	if ran {
		t.Fatal("handler ran with an invalid log level")
	}
	if got := termcore.ExitCodeOf(err); got != 2 {
		t.Fatalf("ExitCodeOf = %d, want 2", got)
	}
}
