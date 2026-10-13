package node

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/CHE3MZ/terminull/render/renderers/config"
	"github.com/CHE3MZ/terminull/render/terminal"
	"github.com/CHE3MZ/terminull/render/utils"
	"golang.org/x/text/width"
)

// EchelonNode is one tree node: title, spinner/status, duration,
// description lines, and children. All methods are goroutine-safe.
type EchelonNode struct {
	lock                    sync.RWMutex
	done                    sync.WaitGroup
	status                  string
	title                   string
	titleColor              int
	description             []string
	visibleDescriptionLines int
	config                  *config.InteractiveRendererConfig
	startTime               time.Time
	endTime                 time.Time
	children                []*EchelonNode
}

// StartNewEchelonNode creates a node and starts it immediately.
func StartNewEchelonNode(title string, config *config.InteractiveRendererConfig) *EchelonNode {
	result := NewEchelonNode(title, config)
	result.Start()
	return result
}

// NewEchelonNode creates a paused node with the neutral title color.
func NewEchelonNode(title string, config *config.InteractiveRendererConfig) *EchelonNode {
	zeroTime := time.Time{}
	result := &EchelonNode{
		status:                  "⏸",
		title:                   title,
		titleColor:              config.Colors.NeutralColor,
		description:             make([]string, 0),
		visibleDescriptionLines: config.VisibleDescriptionLines,
		config:                  config,
		startTime:               zeroTime,
		endTime:                 zeroTime,
		children:                make([]*EchelonNode, 0),
	}
	result.done.Add(1)
	return result
}

// GetChildren returns the child nodes.
func (node *EchelonNode) GetChildren() []*EchelonNode {
	node.lock.RLock()
	defer node.lock.RUnlock()
	return node.children
}

// UpdateTitle replaces the node title.
func (node *EchelonNode) UpdateTitle(text string) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.title = text
}

// UpdateConfig replaces the node config.
func (node *EchelonNode) UpdateConfig(config *config.InteractiveRendererConfig) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.config = config
}

// ClearAllChildren drops all child nodes.
func (node *EchelonNode) ClearAllChildren() {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.children = make([]*EchelonNode, 0)
}

// ClearDescription drops all description lines.
func (node *EchelonNode) ClearDescription() {
	node.SetDescription(make([]string, 0))
}

// SetDescription replaces the description lines.
func (node *EchelonNode) SetDescription(description []string) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.description = description
}

// SetVisibleDescriptionLines caps rendered description lines.
func (node *EchelonNode) SetVisibleDescriptionLines(count int) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.visibleDescriptionLines = count
}

// DescriptionLength returns the description line count.
func (node *EchelonNode) DescriptionLength() int {
	node.lock.RLock()
	defer node.lock.RUnlock()
	return len(node.description)
}

// Render returns the node lines: fancy title, children, then the
// description tail (truncated with "..." when over the visible cap).
func (node *EchelonNode) Render() []string {
	node.lock.RLock()
	defer node.lock.RUnlock()
	title := node.fancyTitle()
	tail := node.renderChildren()
	if len(node.description) > node.visibleDescriptionLines && node.visibleDescriptionLines >= 0 {
		tail = append(tail, "...")
		tail = append(tail, node.description[(len(node.description)-node.visibleDescriptionLines):]...)
	} else {
		tail = append(tail, node.description...)
	}
	indent := "  " // two spaces by default
	props, _ := width.LookupString(title)
	if props.Kind() == width.EastAsianWide || props.Kind() == width.EastAsianFullwidth {
		indent = "   " // three spaces since title start with a wide emoji
	}
	result := []string{title}
	for _, descriptionLine := range tail {
		result = append(result, indent+descriptionLine)
	}

	return result
}

func (node *EchelonNode) renderChildren() []string {
	var result []string
	for _, child := range node.children {
		result = append(result, child.Render()...)
	}
	return result
}

func (node *EchelonNode) fancyTitle() string {
	duration := utils.FormatDuration(node.executionDuration(), len(node.children) == 0)
	isRunning := node.isRunning()
	prefix := node.status
	if isRunning {
		prefix = node.config.CurrentProgressIndicatorFrame()
	}
	coloredTitle := node.title
	if node.titleColor >= 0 {
		coloredTitle = terminal.GetColoredText(node.titleColor, node.title)
	}
	return fmt.Sprintf("%s %s %s", prefix, coloredTitle, duration)
}

// ExecutionDuration returns how long the node ran (or has run so far).
func (node *EchelonNode) ExecutionDuration() time.Duration {
	node.lock.RLock()
	defer node.lock.RUnlock()
	return node.executionDuration()
}

func (node *EchelonNode) executionDuration() time.Duration {
	if !node.startTime.IsZero() && node.endTime.IsZero() {
		return time.Since(node.startTime)
	}
	return node.endTime.Sub(node.startTime)
}

// HasStarted reports whether the node started.
func (node *EchelonNode) HasStarted() bool {
	node.lock.RLock()
	defer node.lock.RUnlock()
	return !node.startTime.IsZero()
}

// HasCompleted reports whether the node completed.
func (node *EchelonNode) HasCompleted() bool {
	node.lock.RLock()
	defer node.lock.RUnlock()
	return !node.endTime.IsZero()
}

// IsRunning reports whether the node started but not completed.
func (node *EchelonNode) IsRunning() bool {
	node.lock.RLock()
	defer node.lock.RUnlock()
	return node.isRunning()
}

func (node *EchelonNode) isRunning() bool {
	return !node.startTime.IsZero() && node.endTime.IsZero()
}

// StartNewChild creates, starts, and attaches a child node.
func (node *EchelonNode) StartNewChild(childName string) *EchelonNode {
	child := StartNewEchelonNode(childName, node.config)
	node.AddNewChild(child)
	return child
}

// FindOrCreateChild returns the last child with the title, creating it first.
func (node *EchelonNode) FindOrCreateChild(childTitle string) *EchelonNode {
	node.lock.Lock()
	defer node.lock.Unlock()
	// look from the end since this is a common pattern to get the last child
	for i := len(node.children) - 1; i >= 0; i-- {
		child := node.children[i]
		if child.title == childTitle {
			return child
		}
	}
	child := NewEchelonNode(childTitle, node.config)
	node.children = append(node.children, child)
	return child
}

// AddNewChild attaches an existing child node.
func (node *EchelonNode) AddNewChild(child *EchelonNode) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.children = append(node.children, child)
}

// Start marks the node started (once).
func (node *EchelonNode) Start() {
	node.lock.Lock()
	defer node.lock.Unlock()
	if node.startTime.IsZero() {
		node.startTime = time.Now()
	}
}

// CompleteWithColor completes the node with a status glyph and title color.
func (node *EchelonNode) CompleteWithColor(status string, titleColor int) {
	if !node.endTime.IsZero() {
		return
	}
	node.lock.Lock()
	defer node.lock.Unlock()
	node.endTime = time.Now()
	if node.startTime.IsZero() {
		node.startTime = node.endTime
	}
	node.status = status
	node.titleColor = titleColor
	node.done.Done()
}

// Complete completes the node keeping its status glyph.
func (node *EchelonNode) Complete() {
	if !node.endTime.IsZero() {
		return
	}
	node.lock.Lock()
	defer node.lock.Unlock()
	node.endTime = time.Now()
	if node.startTime.IsZero() {
		node.startTime = node.endTime
	}
	node.done.Done()
}

// SetTitleColor replaces the ANSI title color.
func (node *EchelonNode) SetTitleColor(ansiColor int) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.titleColor = ansiColor
}

// SetStatus replaces the status glyph.
func (node *EchelonNode) SetStatus(text string) {
	node.lock.Lock()
	defer node.lock.Unlock()
	node.status = text
}

// WaitCompletion blocks until the node completes.
func (node *EchelonNode) WaitCompletion() {
	node.done.Wait()
}

// AppendDescription merges text into the description, joining the first
// new line onto the last existing one (streaming writes rarely end cleanly).
func (node *EchelonNode) AppendDescription(text string) {
	if node.HasCompleted() {
		return
	}
	node.lock.Lock()
	defer node.lock.Unlock()
	linesToAppend := strings.Split(text, "\n")
	if len(linesToAppend) == 0 {
		return
	}
	if len(node.description) == 0 {
		node.description = linesToAppend
		return
	}
	// append first new line to the last one
	node.description[len(node.description)-1] = node.description[len(node.description)-1] + linesToAppend[0]
	if len(linesToAppend) > 1 {
		node.description = append(node.description, linesToAppend[1:]...)
	}
}
