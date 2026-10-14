package renderers

import (
	"bufio"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/CHE3MZ/terminull/render"
	"github.com/CHE3MZ/terminull/render/renderers/config"
	"github.com/CHE3MZ/terminull/render/renderers/internal/console"
	"github.com/CHE3MZ/terminull/render/renderers/internal/node"
	"github.com/CHE3MZ/terminull/render/terminal"
)

const disableAutoWrap = "\u001B[?7l"
const enableAutoWrap = "\u001B[?7h"
const defaultFrameBufSize = 38400 // 80 by 120 of 4 bytes UTF-8 characters

// InteractiveRenderer draws a live hierarchical tree with incremental
// redraws. Run StartDrawing in a goroutine and defer StopDrawing:
//
//	renderer := renderers.NewInteractiveRenderer(os.Stdout, nil)
//	go renderer.StartDrawing()
//	defer renderer.StopDrawing()
//
// Foreign output (prompts, spinners) must pause the loop first with
// PauseDrawing and resume after with ResumeDrawing, or both writers
// garble the terminal.
type InteractiveRenderer struct {
	out               *bufio.Writer
	rootNode          *node.EchelonNode
	config            *config.InteractiveRendererConfig
	currentFrameLines []string
	drawLock          sync.Mutex
	paused            atomic.Bool
	terminalHeight    int

	StubRenderer
}

// NewInteractiveRenderer builds an InteractiveRenderer for a terminal file.
// A nil config selects emoji on Unix and symbols-only on Windows.
func NewInteractiveRenderer(out *os.File, rendererConfig *config.InteractiveRendererConfig) *InteractiveRenderer {
	if rendererConfig == nil {
		rendererConfig = config.NewDefaultRenderingConfig()
	}
	return &InteractiveRenderer{
		out:            bufio.NewWriterSize(out, defaultFrameBufSize),
		rootNode:       node.NewEchelonNode("root", rendererConfig),
		config:         rendererConfig,
		terminalHeight: console.TerminalHeight(out),
	}
}

func findScopedNode(scopes []string, r *InteractiveRenderer) *node.EchelonNode {
	result := r.rootNode
	for _, scope := range scopes {
		result = result.FindOrCreateChild(scope)
	}
	return result
}

// RenderScopeStarted opens (or reuses) the scope node and starts it.
func (r *InteractiveRenderer) RenderScopeStarted(entry *render.LogScopeStarted) {
	findScopedNode(entry.GetScopes(), r).Start()
}

// RenderScopeFinished completes the scope node, keeping failure output
// visible per config and collapsing successes.
func (r *InteractiveRenderer) RenderScopeFinished(entry *render.LogScopeFinished) {
	n := findScopedNode(entry.GetScopes(), r)

	switch entry.FinishType() {
	case render.FinishTypeSucceeded:
		if n != r.rootNode {
			n.ClearAllChildren()
			n.ClearDescription()
		}
		n.CompleteWithColor(r.config.SuccessStatus, r.config.Colors.SuccessColor)
	case render.FinishTypeFailed:
		n.SetVisibleDescriptionLines(r.config.DescriptionLinesWhenFailed)
		n.CompleteWithColor(r.config.FailureStatus, r.config.Colors.FailureColor)
	case render.FinishTypeSkipped:
		if r.config.DescriptionLinesWhenSkipped != 0 {
			n.SetVisibleDescriptionLines(r.config.DescriptionLinesWhenSkipped)
		} else if n != r.rootNode {
			n.ClearAllChildren()
			n.ClearDescription()
		}
		n.CompleteWithColor(r.config.SkippedStatus, r.config.Colors.NeutralColor)
	}
}

// RenderMessage appends the message to the scope node's description.
func (r *InteractiveRenderer) RenderMessage(entry *render.LogEntryMessage) {
	findScopedNode(entry.GetScopes(), r).AppendDescription(entry.GetMessage())
}

// StartDrawing prepares the terminal and redraws until the root node
// completes. Run it in a goroutine; setup itself is synchronous so
// prompts started afterwards see a stable terminal.
func (r *InteractiveRenderer) StartDrawing() {
	r.Setup()
	for !r.rootNode.HasCompleted() {
		r.DrawFrame()
		time.Sleep(r.config.RefreshRate)
	}
}

// Setup prepares the terminal once: VT processing plus no-autowrap,
// flushed immediately. Idempotent: StartDrawing calls it too. Call it
// synchronously before any prompt runs so terminal configuration never
// races interactive input.
func (r *InteractiveRenderer) Setup() {
	_ = console.PrepareTerminalEnvironment()
	r.drawLock.Lock()
	defer r.drawLock.Unlock()
	_, _ = r.out.WriteString(disableAutoWrap)
	_ = r.out.Flush()
}

// StopDrawing completes the tree, draws the final frame, restores
// autowrap, and flushes.
func (r *InteractiveRenderer) StopDrawing() {
	r.rootNode.Complete()
	// one last redraw
	r.paused.Store(false)
	r.DrawFrame()
	// don't leave autowrap disabled in the terminal
	_, _ = r.out.WriteString(enableAutoWrap)
	_ = r.out.Flush()
}

// PauseDrawing freezes redraws for foreign output (prompts, spinners,
// bars) sharing the terminal. Tree state is untouched; ResumeDrawing
// continues where it left off. Safe to call without StartDrawing.
func (r *InteractiveRenderer) PauseDrawing() {
	r.drawLock.Lock()
	defer r.drawLock.Unlock()
	r.paused.Store(true)
	_ = r.out.Flush()
}

// ResumeDrawing unfreezes redraws after PauseDrawing.
func (r *InteractiveRenderer) ResumeDrawing() {
	r.paused.Store(false)
}

// DrawFrame renders one frame with an incremental terminal update,
// unless paused.
func (r *InteractiveRenderer) DrawFrame() {
	if r.paused.Load() {
		return
	}
	r.drawLock.Lock()
	defer r.drawLock.Unlock()
	var newFrameLines []string
	for _, n := range r.rootNode.GetChildren() {
		newFrameLines = append(newFrameLines, n.Render()...)
	}
	if r.terminalHeight > 0 {
		terminal.CalculateIncrementalUpdateMaxLines(r.out, r.currentFrameLines, newFrameLines, r.terminalHeight)
	} else {
		terminal.CalculateIncrementalUpdate(r.out, r.currentFrameLines, newFrameLines)
	}
	r.currentFrameLines = newFrameLines
}
