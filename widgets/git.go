package widgets

import (
	"fmt"

	"github.com/CS-5/statusline/render"
	"github.com/CS-5/statusline/ui"
)

// GitWidget displays git branch and change status
type GitWidget struct{}

// NewGit creates a new git widget
func NewGit() Widget {
	return &GitWidget{}
}

// Render displays git branch with optional changes indicator
func (w *GitWidget) Render(ctx *render.RenderContext) *ui.Segment {
	if ctx.GitBranch == "" {
		return nil
	}

	var text string
	bgColor := ui.ColorGitBg
	fgColor := ui.ColorGitFg

	if ctx.GitHasChanges {
		text = fmt.Sprintf(" %s %s %s ", ui.Branch, ctx.GitBranch, ui.GitChangesIcon)
		bgColor = ui.ColorGitChangesBg
		fgColor = ui.ColorGitChangesFg
	} else {
		text = fmt.Sprintf(" %s %s ", ui.Branch, ctx.GitBranch)
	}

	return ui.CreateSegment(text, bgColor, fgColor)
}