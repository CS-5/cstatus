package render

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
)

// This package is responsible for parsing JSON, executing git commands, and providing render context.

// https://docs.anthropic.com/en/docs/claude-code/statusline
type ClaudeCodeInput struct {
	HookEventName  string      `json:"hook_event_name"`
	SessionID      string      `json:"session_id"`
	TranscriptPath string      `json:"transcript_path"`
	Cwd            string      `json:"cwd"`
	Model          Model       `json:"model"`
	Workspace      Workspace   `json:"workspace"`
	Version        string      `json:"version"`
	OutputStyle    OutputStyle `json:"output_style"`
	Cost           Cost        `json:"cost"`
}

type Model struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
}

type Workspace struct {
	CurrentDir string `json:"current_dir"`
	ProjectDir string `json:"project_dir"`
}

type OutputStyle struct {
	Name string `json:"name"`
}

type Cost struct {
	TotalCostUSD       float64 `json:"total_cost_usd"`
	TotalDurationMs    int64   `json:"total_duration_ms"`
	TotalAPIDurationMs int64   `json:"total_api_duration_ms"`
	TotalLinesAdded    int     `json:"total_lines_added"`
	TotalLinesRemoved  int     `json:"total_lines_removed"`
}

type RenderContext struct {
	Claude        *ClaudeCodeInput
	TokenMetrics  *TokenMetrics
	GitBranch     string
	GitHasChanges bool
	WorkingDir    string
	ProjectName   string
}

// NewRenderContext creates a new context from Claude Code JSON string
func NewRenderContext(claudeJSON string) (*RenderContext, error) {
	var claude ClaudeCodeInput
	if err := json.Unmarshal([]byte(claudeJSON), &claude); err != nil {
		return nil, err
	}

	ctx := &RenderContext{
		Claude: &claude,
	}

	// Compute working directory
	ctx.WorkingDir = ctx.getWorkingDir()

	// Compute project name
	ctx.ProjectName, _ = ctx.getProjectName()

	// Compute git info
	ctx.GitBranch, _ = ctx.getGitBranch()
	ctx.GitHasChanges = ctx.checkGitChanges()

	// Parse token metrics from transcript if available
	if claude.TranscriptPath != "" {
		ctx.TokenMetrics = ParseTokenMetrics(claude.TranscriptPath)
	} else {
		ctx.TokenMetrics = &TokenMetrics{}
	}

	return ctx, nil
}

func (ctx *RenderContext) getWorkingDir() string {
	if ctx.Claude.Workspace.CurrentDir != "" {
		return ctx.Claude.Workspace.CurrentDir
	}
	return ctx.Claude.Cwd
}

func (ctx *RenderContext) getProjectName() (string, bool) {
	dirs := []string{
		ctx.Claude.Workspace.ProjectDir,
		ctx.Claude.Workspace.CurrentDir,
		ctx.Claude.Cwd,
	}

	for _, dir := range dirs {
		if dir != "" {
			return filepath.Base(dir), true
		}
	}
	return "", false
}

func (ctx *RenderContext) getGitBranch() (string, bool) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = ctx.WorkingDir
	output, err := cmd.Output()
	if err != nil {
		return "", false
	}
	branch := strings.TrimSpace(string(output))
	return branch, branch != ""
}

func (ctx *RenderContext) checkGitChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = ctx.WorkingDir
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(output))) > 0
}
