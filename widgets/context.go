package widgets

import (
	"fmt"

	"github.com/CS-5/statusline/render"
	"github.com/CS-5/statusline/ui"
)

// ContextWidget displays context usage information
type ContextWidget struct{}

// NewContext creates a new context widget
func NewContext() Widget {
	return &ContextWidget{}
}

// Render displays context information
func (w *ContextWidget) Render(ctx *render.RenderContext) *ui.Segment {
	contextStr := w.formatContext(ctx)
	text := fmt.Sprintf(" %s %s ", ui.ContextIcon, contextStr)
	return ui.CreateSegment(text, ui.ColorContextBg, ui.ColorContextFg)
}

func (w *ContextWidget) formatContext(ctx *render.RenderContext) string {
	if ctx.Claude.TranscriptPath == "" {
		return "0 (100%)"
	}
	
	// Use actual token metrics if available
	if ctx.TokenMetrics != nil && ctx.TokenMetrics.ContextLength > 0 {
		return fmt.Sprintf("%d tokens", ctx.TokenMetrics.ContextLength)
	}
	
	return "ctx (--)"
}