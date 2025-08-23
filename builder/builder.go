package builder

import (
	"strings"

	"github.com/CS-5/statusline/render"
	"github.com/CS-5/statusline/ui"
	"github.com/CS-5/statusline/widgets"
)

// WidgetRegistry maps widget names to their constructors
var WidgetRegistry = map[string]func() widgets.Widget{
	"project": widgets.NewProject,
	"git":     widgets.NewGit,
	"model":   widgets.NewModel,
	"session": widgets.NewSession,
	"context": widgets.NewContext,
}

// StatuslineBuilder builds and renders statuslines
type StatuslineBuilder struct {
	widgetNames []string
}

// New creates a new statusline builder with default widget order
func New() *StatuslineBuilder {
	return &StatuslineBuilder{
		widgetNames: []string{"project", "git", "model", "session", "context"},
	}
}

// WithWidgets sets the widget order for the builder
func (b *StatuslineBuilder) WithWidgets(widgetNames []string) *StatuslineBuilder {
	b.widgetNames = widgetNames
	return b
}

// Build creates widgets and renders the complete statusline
func (b *StatuslineBuilder) Build(ctx *render.RenderContext) string {
	var segments []*ui.Segment

	for _, widgetName := range b.widgetNames {
		constructor, exists := WidgetRegistry[widgetName]
		if !exists {
			continue
		}

		widget := constructor()
		segment := widget.Render(ctx)
		if segment != nil && !segment.IsEmpty() {
			segments = append(segments, segment)
		}
	}

	return b.renderSegments(segments)
}

func (b *StatuslineBuilder) renderSegments(segments []*ui.Segment) string {
	if len(segments) == 0 {
		return ""
	}

	var result strings.Builder
	for i, segment := range segments {
		bgColor := ui.HexToAnsi(segment.BgHex, true)
		fgColor := ui.HexToAnsi(segment.FgHex, false)

		result.WriteString(bgColor)
		result.WriteString(fgColor)
		result.WriteString(segment.Text)

		if i < len(segments)-1 {
			nextBg := ui.HexToAnsi(segments[i+1].BgHex, true)
			currentBgAsFg := ui.HexToAnsi(segment.BgHex, false)
			result.WriteString("\x1b[0m")
			result.WriteString(nextBg)
			result.WriteString(currentBgAsFg)
			result.WriteString(ui.SeparatorRight)
		} else {
			currentBgAsFg := ui.HexToAnsi(segment.BgHex, false)
			result.WriteString("\x1b[0m")
			result.WriteString(currentBgAsFg)
			result.WriteString(ui.SeparatorRight)
		}
	}

	result.WriteString("\x1b[0m")
	return result.String()
}
