package rodaine_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	rodaine "github.com/CHE3MZ/terminull/table/rodaine"
	"github.com/mattn/go-runewidth"
)

type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }

func TestPrintGolden(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("ID", "NAME", "COST ($)").WithWriter(&buf)
	tb.AddRow("1", "Foobar", "1.23")
	tb.AddRow("2", "Fizzbuzz", "4.56")
	tb.AddRow("3", "Gizmo", "78.90")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	want := "" +
		"ID  NAME      COST ($)  \n" +
		"1   Foobar    1.23      \n" +
		"2   Fizzbuzz  4.56      \n" +
		"3   Gizmo     78.90     \n"

	if buf.String() != want {
		t.Fatalf("Print() = %q, want %q", buf.String(), want)
	}
}

func TestPrintNoHeaders(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("A", "B").WithWriter(&buf).WithPrintHeaders(false)
	tb.AddRow("1", "2")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if got, want := buf.String(), "1  2  \n"; got != want {
		t.Fatalf("Print() = %q, want %q", got, want)
	}
}

func TestPaddingZero(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("A", "B").WithWriter(&buf).WithPadding(0)
	tb.AddRow("1", "2")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if got, want := buf.String(), "AB\n12\n"; got != want {
		t.Fatalf("Print() = %q, want %q", got, want)
	}
}

func TestNegativePaddingClamped(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("A", "B").WithWriter(&buf).WithPadding(-3)
	tb.AddRow("1", "2")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if got, want := buf.String(), "AB\n12\n"; got != want {
		t.Fatalf("Print() = %q, want %q", got, want)
	}
}

func TestCJKAlignment(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("NAME", "QTY").WithWriter(&buf)
	tb.AddRow("日本", "3")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	want := "" +
		"NAME  QTY  \n" +
		"日本  3    \n"

	if buf.String() != want {
		t.Fatalf("Print() = %q, want %q", buf.String(), want)
	}

	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	if runewidth.StringWidth(lines[0]) != runewidth.StringWidth(lines[1]) {
		t.Fatalf("CJK rows misaligned: %q vs %q", lines[0], lines[1])
	}
}

func TestMultilineRow(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("H1", "H2").WithWriter(&buf)
	tb.AddRow("a\nb", "c")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	want := "" +
		"H1  H2  \n" +
		"a   c   \n" +
		"b       \n"

	if buf.String() != want {
		t.Fatalf("Print() = %q, want %q", buf.String(), want)
	}
}

func TestSetRowsTruncatesAndPads(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("A", "B").WithWriter(&buf)
	tb.SetRows([][]string{{"1"}, {"1", "2", "3"}})

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	want := "" +
		"A  B  \n" +
		"1     \n" +
		"1  2  \n"

	if buf.String() != want {
		t.Fatalf("Print() = %q, want %q", buf.String(), want)
	}
}

func TestHeaderSeparatorRow(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("ID", "NAME").WithWriter(&buf).WithHeaderSeparatorRow('-')
	tb.AddRow("1", "x")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	want := "" +
		"ID  NAME  \n" +
		"--  ----  \n" +
		"1   x     \n"

	if buf.String() != want {
		t.Fatalf("Print() = %q, want %q", buf.String(), want)
	}
}

func TestPrintErrorPropagates(t *testing.T) {
	t.Parallel()

	tb := rodaine.New("A").WithWriter(errWriter{})
	tb.AddRow("x")

	if err := tb.Print(); err == nil {
		t.Fatal("Print() = nil, want write error")
	}
}

func TestAddRowTruncatesExtraCells(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("A").WithWriter(&buf)
	tb.AddRow("x", "y")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	if got, want := buf.String(), "A  \nx  \n"; got != want {
		t.Fatalf("Print() = %q, want %q", got, want)
	}
}

func TestHeaderFormatterWithSeparator(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("name").WithWriter(&buf).WithHeaderSeparatorRow('-')
	tb.WithHeaderFormatter(func(f string, v ...interface{}) string {
		return strings.ToUpper(fmt.Sprintf(f, v...))
	})
	tb.AddRow("x")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	want := "" +
		"NAME  \n" +
		"----  \n" +
		"x     \n"

	if buf.String() != want {
		t.Fatalf("Print() = %q, want %q", buf.String(), want)
	}
}

func TestFirstColumnFormatter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	tb := rodaine.New("id", "name").WithWriter(&buf)
	tb.WithFirstColumnFormatter(func(f string, v ...interface{}) string {
		return strings.ToUpper(fmt.Sprintf(f, v...))
	})
	tb.AddRow("a", "b")

	if err := tb.Print(); err != nil {
		t.Fatalf("Print() error = %v", err)
	}

	want := "" +
		"id  name  \n" +
		"A   b     \n"

	if buf.String() != want {
		t.Fatalf("Print() = %q, want %q", buf.String(), want)
	}
}
