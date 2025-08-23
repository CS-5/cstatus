package widgets

import (
	"testing"

	"github.com/CS-5/statusline/render"
	"github.com/CS-5/statusline/ui"
)

func TestProjectWidget_Render(t *testing.T) {
	widget := NewProject()

	tests := []struct {
		name        string
		projectName string
		expectNil   bool
	}{
		{"with project name", "test-project", false},
		{"empty project name", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &render.RenderContext{
				ProjectName: tt.projectName,
			}

			segment := widget.Render(ctx)

			if tt.expectNil {
				if segment != nil {
					t.Error("Expected nil segment for empty project name")
				}
				return
			}

			if segment == nil {
				t.Fatal("Expected segment, got nil")
			}

			expectedText := " 📁 " + tt.projectName + " "
			if segment.Text != expectedText {
				t.Errorf("Text = %v, want %v", segment.Text, expectedText)
			}
			if segment.BgHex != ui.ColorProjectBg {
				t.Errorf("BgHex = %v, want %v", segment.BgHex, ui.ColorProjectBg)
			}
			if segment.FgHex != ui.ColorProjectFg {
				t.Errorf("FgHex = %v, want %v", segment.FgHex, ui.ColorProjectFg)
			}
		})
	}
}