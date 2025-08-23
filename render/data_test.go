package render

import (
	"testing"
)

func TestNewRenderContext(t *testing.T) {
	jsonData := `{
		"cwd": "/test/path",
		"model": {
			"display_name": "Test Model"
		},
		"workspace": {
			"current_dir": "/test/current",
			"project_dir": "/test/project"
		},
		"cost": {
			"total_cost_usd": 1.23
		}
	}`

	ctx, err := NewRenderContext(jsonData)
	if err != nil {
		t.Fatalf("NewRenderContext failed: %v", err)
	}

	if ctx.Claude == nil {
		t.Error("Claude not set correctly")
	}
	if ctx.WorkingDir != "/test/current" {
		t.Errorf("WorkingDir = %v, want %v", ctx.WorkingDir, "/test/current")
	}
	if ctx.ProjectName != "project" {
		t.Errorf("ProjectName = %v, want %v", ctx.ProjectName, "project")
	}
	if ctx.TokenMetrics == nil {
		t.Error("TokenMetrics should not be nil")
	}
}