package widgets

import (
	"fmt"

	"github.com/CS-5/statusline/render"
	"github.com/CS-5/statusline/ui"
)

// ProjectWidget displays the project name
type ProjectWidget struct{}

// NewProject creates a new project widget
func NewProject() Widget {
	return &ProjectWidget{}
}

// Render displays the project name with project icon
func (w *ProjectWidget) Render(ctx *render.RenderContext) *ui.Segment {
	if ctx.ProjectName == "" {
		return nil
	}

	text := fmt.Sprintf(" %s %s ", ui.ProjectIcon, ctx.ProjectName)
	return ui.CreateSegment(text, ui.ColorProjectBg, ui.ColorProjectFg)
}