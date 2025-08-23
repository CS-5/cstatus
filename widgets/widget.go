package widgets

import (
	"github.com/CS-5/statusline/render"
	"github.com/CS-5/statusline/ui"
)

// Widget interface that all widgets must implement
type Widget interface {
	Render(ctx *render.RenderContext) *ui.Segment
}