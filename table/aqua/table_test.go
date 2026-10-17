package aqua_test

import (
	"bytes"
	"strings"
	"testing"

	aqua "github.com/CHE3MZ/terminull/table/aqua"
)

func draw(t *testing.T, build func(*aqua.Table)) string {
	t.Helper()

	var buf bytes.Buffer
	tb := aqua.New(&buf)
	build(tb)
	tb.Render()
	return buf.String()
}

func TestRenderBorderedGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("NAME", "STATUS")
		tb.AddRow("web", "ok")
	})

	want := "" +
		"┌──────┬────────┐\n" +
		"│ NAME │ STATUS │\n" +
		"├──────┼────────┤\n" +
		"│ web  │ ok     │\n" +
		"└──────┴────────┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderHeadersOnlyGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("NAME", "STATUS")
	})

	want := "" +
		"┌──────┬────────┐\n" +
		"│ NAME │ STATUS │\n" +
		"└──────┴────────┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderHeaderCenteredGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("N", "K")
		tb.AddRow("abc", "1")
	})

	want := "" +
		"┌─────┬───┐\n" +
		"│  N  │ K │\n" +
		"├─────┼───┤\n" +
		"│ abc │ 1 │\n" +
		"└─────┴───┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderRightAlignedGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("KEY")
		tb.SetAlignment(aqua.AlignRight)
		tb.AddRow("ab")
	})

	want := "" +
		"┌─────┐\n" +
		"│ KEY │\n" +
		"├─────┤\n" +
		"│  ab │\n" +
		"└─────┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderWrappedColumnGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("code")
		tb.AddRow("abcdefgh")
		tb.SetColumnMaxWidth(6)
		tb.SetAvailableWidth(1)
		tb.SetPadding(0)
	})

	want := "" +
		"┌──────┐\n" +
		"│ code │\n" +
		"├──────┤\n" +
		"│abcde-│\n" +
		"│fgh   │\n" +
		"└──────┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderCJKGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("名", "数量")
		tb.AddRow("日本", "3")
	})

	want := "" +
		"┌──────┬──────┐\n" +
		"│  名  │ 数量 │\n" +
		"├──────┼──────┤\n" +
		"│ 日本 │ 3    │\n" +
		"└──────┴──────┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderRowLinesDisabled(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("K", "V")
		tb.AddRow("a", "1")
		tb.AddRow("b", "2")
		tb.SetRowLines(false)
	})

	want := "" +
		"┌───┬───┐\n" +
		"│ K │ V │\n" +
		"├───┼───┤\n" +
		"│ a │ 1 │\n" +
		"│ b │ 2 │\n" +
		"└───┴───┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderAutoMerge(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("NAME", "TAG")
		tb.AddRow("a", "x")
		tb.AddRow("a", "y")
		tb.SetAutoMerge(true)
	})

	want := "" +
		"┌──────┬─────┐\n" +
		"│ NAME │ TAG │\n" +
		"├──────┼─────┤\n" +
		"│ a    │ x   │\n" +
		"│      ├─────┤\n" +
		"│      │ y   │\n" +
		"└──────┴─────┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderASCIIDividersGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("NAME", "STATUS")
		tb.AddRow("web", "ok")
		tb.SetDividers(aqua.ASCIIDividers)
	})

	want := "" +
		"+------+--------+\n" +
		"| NAME | STATUS |\n" +
		"+------+--------+\n" +
		"| web  | ok     |\n" +
		"+------+--------+\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderPreservesANSIContent(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("COLOR", "X")
		tb.AddRow("\x1b[31mred\x1b[0m", "x")
	})

	want := "" +
		"┌───────┬───┐\n" +
		"│ COLOR │ X │\n" +
		"├───────┼───┤\n" +
		"│ \x1b[31mred\x1b[0m   │ x │\n" +
		"└───────┴───┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
	if strings.Count(got, "\x1b[31m") != 1 {
		t.Fatalf("Render() re-emitted or dropped ANSI: %q", got)
	}
}

func TestRenderDefaultHasNoANSI(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("NAME", "STATUS")
		tb.AddRow("web", "ok")
	})

	if strings.Contains(got, "\x1b") {
		t.Fatalf("default Render() emitted ANSI: %q", got)
	}
}

func TestRenderLineStyleOptIn(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("NAME")
		tb.AddRow("web")
		tb.SetLineStyle(aqua.StyleYellow)
	})

	if !strings.Contains(got, "\x1b[33m┌") {
		t.Fatalf("Render() missing yellow line style: %q", got)
	}
	if !strings.Contains(got, "\x1b[0m") {
		t.Fatalf("Render() missing reset after line style: %q", got)
	}
}

func TestRenderHeaderColSpanGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("TITLE")
		tb.SetHeaderColSpans(0, 2)
		tb.AddRow("a", "b")
	})

	want := "" +
		"┌───────┐\n" +
		"│ TITLE │\n" +
		"├───┬───┤\n" +
		"│ a │ b │\n" +
		"└───┴───┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderFooterGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("NAME")
		tb.AddRow("a")
		tb.SetFooters("total")
	})

	want := "" +
		"┌───────┐\n" +
		"│ NAME  │\n" +
		"├───────┤\n" +
		"│ a     │\n" +
		"├───────┤\n" +
		"│ total │\n" +
		"└───────┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderBordersDisabled(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("NAME", "STATUS")
		tb.AddRow("web", "ok")
		tb.SetBorders(false)
	})

	want := "" +
		" NAME │ STATUS \n" +
		"──────┼────────\n" +
		" web  │ ok     \n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderFillWidthGolden(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("A", "B")
		tb.AddRows([]string{"1", "2"})
		tb.SetAvailableWidth(20)
		tb.SetFillWidth(true)
	})

	want := "" +
		"┌────────┬────────┐\n" +
		"│   A    │   B    │\n" +
		"├────────┼────────┤\n" +
		"│ 1      │ 2      │\n" +
		"└────────┴────────┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestRenderHeaderStyleOptIn(t *testing.T) {
	t.Parallel()

	got := draw(t, func(tb *aqua.Table) {
		tb.SetHeaders("NAME")
		tb.AddRow("web")
		tb.SetHeaderStyle(aqua.StyleBold)
	})

	if !strings.Contains(got, "\x1b[1mNAME\x1b[0m") {
		t.Fatalf("Render() missing bold header: %q", got)
	}
}

func TestRenderEmptyDrawsNothing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	aqua.New(&buf).Render()

	if buf.String() != "" {
		t.Fatalf("empty Render() = %q, want silence", buf.String())
	}
}

func TestLoadCSVGolden(t *testing.T) {
	t.Parallel()

	csvData := "id,name\n1,alpha\n2,beta\n"

	got := draw(t, func(tb *aqua.Table) {
		if err := tb.LoadCSV(strings.NewReader(csvData), true); err != nil {
			t.Fatalf("LoadCSV() error = %v", err)
		}
	})

	want := "" +
		"┌────┬───────┐\n" +
		"│ id │ name  │\n" +
		"├────┼───────┤\n" +
		"│ 1  │ alpha │\n" +
		"├────┼───────┤\n" +
		"│ 2  │ beta  │\n" +
		"└────┴───────┘\n"

	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}

func TestLoadCSVError(t *testing.T) {
	t.Parallel()

	tb := aqua.New(&bytes.Buffer{})
	err := tb.LoadCSV(strings.NewReader("a,\"b\n"), false)
	if err == nil {
		t.Fatal("LoadCSV() = nil, want parse error")
	}
}

func TestClearResetsRows(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := aqua.New(&buf)
	tb.SetHeaders("A")
	tb.AddRow("1")
	if tb.IsEmpty() || tb.RowCount() != 1 {
		t.Fatalf("IsEmpty/RowCount = %v/%d, want false/1", tb.IsEmpty(), tb.RowCount())
	}

	tb.Clear()
	if !tb.IsEmpty() || tb.RowCount() != 0 {
		t.Fatalf("after Clear IsEmpty/RowCount = %v/%d, want true/0", tb.IsEmpty(), tb.RowCount())
	}
}
