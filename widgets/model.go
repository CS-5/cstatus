package widgets

import (
	"fmt"

	"github.com/CS-5/statusline/render"
	"github.com/CS-5/statusline/ui"
)

// ModelWidget displays the Claude model name
type ModelWidget struct{}

// NewModel creates a new model widget
func NewModel() Widget {
	return &ModelWidget{}
}

// Render displays the model name with model icon
func (w *ModelWidget) Render(ctx *render.RenderContext) *ui.Segment {
	modelName := ctx.Claude.Model.DisplayName
	if modelName == "" {
		modelName = "Claude"
	}

	text := fmt.Sprintf(" %s %s ", ui.ModelIcon, modelName)
	return ui.CreateSegment(text, ui.ColorModelBg, ui.ColorModelFg)
}