package renderers

import (
	"bufio"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/CHE3MZ/terminull/termrender"
	"github.com/CHE3MZ/terminull/termrender/renderers/config"
	"github.com/CHE3MZ/terminull/termrender/renderers/internal/console"
	"github.com/CHE3MZ/terminull/termrender/renderers/internal/node"
	"github.com/CHE3MZ/terminull/termrender/terminal"
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
	// loopStarted guards single loop ownership; loopWG tracks the loop so
	// StopDrawing joins it. WaitGroup orderings are safe here because
	// exactly one Add ever happens (CAS winner) before any Wait can block.
	// stopping, set first in StopDrawing, tells late starters to stay out
	// so no setup bytes land after final output.
	loopStarted atomic.Bool
	loopWG      sync.WaitGroup
	stopping    atomic.Bool

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
func (r *InteractiveRenderer) RenderScopeStarted(entry *termrender.LogScopeStarted) {
	findScopedNode(entry.GetScopes(), r).Start()
}

// RenderScopeFinished completes the scope node, keeping failure output
// visible per config and collapsing successes.
func (r *InteractiveRenderer) RenderScopeFinished(entry *termrender.LogScopeFinished) {
	n := findScopedNode(entry.GetScopes(), r)

	switch entry.FinishType() {
	case termrender.FinishTypeSucceeded:
		if n != r.rootNode {
			n.ClearAllChildren()
			n.ClearDescription()
		}
		n.CompleteWithColor(r.config.SuccessStatus, r.config.Colors.SuccessColor)
	case termrender.FinishTypeFailed:
		n.SetVisibleDescriptionLines(r.config.DescriptionLinesWhenFailed)
		n.CompleteWithColor(r.config.FailureStatus, r.config.Colors.FailureColor)
	case termrender.FinishTypeSkipped:
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
func (r *InteractiveRenderer) RenderMessage(entry *termrender.LogEntryMessage) {
	findScopedNode(entry.GetScopes(), r).AppendDescription(entry.GetMessage())
}

// StartDrawing prepares the terminal and redraws until the root node
// completes. Run it in a goroutine; setup itself is synchronous so
// prompts started afterwards see a stable terminal. Only the first call
// owns the loop; calls after stopping started return silently, so a
// delayed start can never write past finalized output.
func (r *InteractiveRenderer) StartDrawing() {
	_ = console.PrepareTerminalEnvironment()
	r.drawLock.Lock()
	if r.stopping.Load() || r.rootNode.HasCompleted() || !r.loopStarted.CompareAndSwap(false, true) {
		r.drawLock.Unlock()
		return
	}
	r.writeNoWrapLocked()
	r.loopWG.Add(1)
	r.drawLock.Unlock()
	defer r.loopWG.Done()
	r.drawLoop()
}

// drawLoop redraws until completion, skipping paused frames.
func (r *InteractiveRenderer) drawLoop() {
	for !r.rootNode.HasCompleted() {
		r.DrawFrame()
		time.Sleep(r.config.RefreshRate)
	}
}

// Setup prepares the terminal once: VT processing plus no-autowrap,
// flushed immediately. Idempotent; StartDrawing calls it too. Call it
// synchronously before any prompt runs so terminal configuration never
// races interactive input. All writer access holds the draw lock, so
// Setup is safe beside a running draw loop.
func (r *InteractiveRenderer) Setup() {
	_ = console.PrepareTerminalEnvironment()
	r.drawLock.Lock()
	defer r.drawLock.Unlock()
	r.writeNoWrapLocked()
}

// writeNoWrapLocked disables autowrap and flushes; the caller holds drawLock.
func (r *InteractiveRenderer) writeNoWrapLocked() {
	_, _ = r.out.WriteString(disableAutoWrap)
	_ = r.out.Flush()
}

// StopDrawing completes the tree, draws the final frame, waits for the
// background loop to observe completion and exit, then restores autowrap
// and flushes. Only after it returns is the terminal free for other
// output: that join is what keeps tables and subsequent writes intact.
// Safe to call twice and without StartDrawing.
func (r *InteractiveRenderer) StopDrawing() {
	// First: no late starter may write past this point.
	r.stopping.Store(true)
	r.rootNode.Complete()
	// one last redraw
	r.paused.Store(false)
	r.DrawFrame()
	// Join a running loop, if any. No loop (never started or already
	// exited) means a zero counter and an immediate return; the single
	// Add rule keeps concurrent Add/Wait sound in every ordering.
	r.loopWG.Wait()
	// don't leave autowrap disabled in the terminal
	r.drawLock.Lock()
	defer r.drawLock.Unlock()
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
