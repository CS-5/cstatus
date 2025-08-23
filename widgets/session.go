package widgets

import (
	"fmt"

	"github.com/CS-5/statusline/render"
	"github.com/CS-5/statusline/ui"
)

// SessionWidget displays session cost and token information
type SessionWidget struct{}

// NewSession creates a new session widget
func NewSession() Widget {
	return &SessionWidget{}
}

// Render displays cost and token information
func (w *SessionWidget) Render(ctx *render.RenderContext) *ui.Segment {
	cost := ctx.Claude.Cost.TotalCostUSD
	costStr := w.formatCost(cost)
	tokensStr := w.formatTokens(cost)

	text := fmt.Sprintf(" %s %s (%s) ", ui.SessionIcon, costStr, tokensStr)
	return ui.CreateSegment(text, ui.ColorSessionBg, ui.ColorSessionFg)
}

func (w *SessionWidget) formatCost(cost float64) string {
	if cost < 0.01 {
		return fmt.Sprintf("%.1f¢", cost*100)
	}
	return fmt.Sprintf("$%.2f", cost)
}

func (w *SessionWidget) formatTokens(cost float64) string {
	tokens := int64(cost * 333333)
	if tokens > 1000000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
	}
	if tokens > 1000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1000)
	}
	return fmt.Sprintf("%d", tokens)
}