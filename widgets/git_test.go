package widgets

import (
	"testing"

	"github.com/CS-5/statusline/render"
	"github.com/CS-5/statusline/ui"
)

func TestGitWidget_Render(t *testing.T) {
	widget := NewGit()

	tests := []struct {
		name       string
		branch     string
		hasChanges bool
		expectNil  bool
		expectText string
		expectBg   string
		expectFg   string
	}{
		{"no branch", "", false, true, "", "", ""},
		{"clean branch", "main", false, false, " ⎇ main ", ui.ColorGitBg, ui.ColorGitFg},
		{"dirty branch", "feature", true, false, " ⎇ feature ● ", ui.ColorGitChangesBg, ui.ColorGitChangesFg},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &render.RenderContext{
				GitBranch:     tt.branch,
				GitHasChanges: tt.hasChanges,
			}

			segment := widget.Render(ctx)

			if tt.expectNil {
				if segment != nil {
					t.Error("Expected nil segment for empty branch")
				}
				return
			}

			if segment == nil {
				t.Fatal("Expected segment, got nil")
			}

			if segment.Text != tt.expectText {
				t.Errorf("Text = %v, want %v", segment.Text, tt.expectText)
			}
			if segment.BgHex != tt.expectBg {
				t.Errorf("BgHex = %v, want %v", segment.BgHex, tt.expectBg)
			}
			if segment.FgHex != tt.expectFg {
				t.Errorf("FgHex = %v, want %v", segment.FgHex, tt.expectFg)
			}
		})
	}
}