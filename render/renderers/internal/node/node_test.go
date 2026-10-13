package node

import (
	"strings"
	"testing"

	"github.com/CHE3MZ/terminull/render/renderers/config"
	"github.com/CHE3MZ/terminull/render/terminal"
)

func testConfig() *config.InteractiveRendererConfig {
	cfg := config.NewDefaultSymbolsOnlyRenderingConfig()
	cfg.Colors = terminal.NoColorSchema()
	return cfg
}

func TestFindOrCreateChildReusesTitle(t *testing.T) {
	t.Parallel()

	root := NewEchelonNode("root", testConfig())
	a := root.FindOrCreateChild("build")
	b := root.FindOrCreateChild("build")

	if a != b {
		t.Fatal("FindOrCreateChild returned different nodes for the same title")
	}
	if len(root.GetChildren()) != 1 {
		t.Fatalf("children = %d, want 1", len(root.GetChildren()))
	}
}

func TestRenderContainsTitle(t *testing.T) {
	t.Parallel()

	root := NewEchelonNode("root", testConfig())
	child := root.StartNewChild("compile")
	child.AppendDescription("gcc main.c")

	lines := root.Render()
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "compile") {
		t.Fatalf("Render() = %q, want to contain child title", joined)
	}
	if !strings.Contains(joined, "gcc main.c") {
		t.Fatalf("Render() = %q, want to contain description", joined)
	}
}

func TestCompletedNodeIgnoresDescription(t *testing.T) {
	t.Parallel()

	n := StartNewEchelonNode("done", testConfig())
	n.Complete()
	n.AppendDescription("late")

	if n.DescriptionLength() != 0 {
		t.Fatalf("DescriptionLength() = %d, want 0 after completion", n.DescriptionLength())
	}
	if !n.HasCompleted() || n.IsRunning() {
		t.Fatal("completed node should report HasCompleted and not IsRunning")
	}
}
