package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/CS-5/statusline/claude"
	"github.com/CS-5/statusline/util"
)

func main() {
	claudeContext, err := claude.NewContextFromReader(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Claude context: %v\n", err)
		os.Exit(1)
	}

	lb := util.NewStatusLineBuilder(claudeContext).
		Append(projectWidget).
		Append(gitStatusWidget).
		Append(sessionWidget).
		Append(contextWidget).
		Append(blockTimerWidget)

	fmt.Print(lb.Render())
}

func projectWidget(claudeContext *claude.Context) *util.Segment {
	if claudeContext.ProjectName == "" {
		return nil
	}
	return util.NewSegment("", claudeContext.ProjectName, "#ffffff", "#8b4513")
}

func gitStatusWidget(claudeContext *claude.Context) *util.Segment {
	// Safely check working directory
	if claudeContext == nil || claudeContext.WorkingDir == "" {
		return nil
	}

	// Check if we're in a git repository
	gitDir := filepath.Join(claudeContext.WorkingDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil
	}

	// Get the current branch name using git CLI with timeout
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = claudeContext.WorkingDir
	
	// Set a reasonable timeout for git commands
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = claudeContext.WorkingDir
	
	output, err := cmd.Output()
	if err != nil {
		// Git command failed - repository might be corrupted or git not available
		return nil
	}

	branchName := strings.TrimSpace(string(output))
	if branchName == "" {
		return nil
	}

	return util.NewSegment("⎇", branchName, "#ffffff", "#ff6b6b")
}

func modelWidget(claudeContext *claude.Context) *util.Segment {
	if claudeContext.Code.Model.DisplayName == "" {
		return nil
	}
	return util.NewSegment("⚡", claudeContext.Code.Model.DisplayName, "#ffffff", "#2d2d2d")
}

func sessionWidget(claudeContext *claude.Context) *util.Segment {
	// Safely check context and code
	if claudeContext == nil || claudeContext.Code == nil {
		return util.NewSegment("§", "$0.00 (0)", "#00ffff", "#202020")
	}
	
	cost := claudeContext.Code.Cost.TotalCostUSD
	costStr := util.FormatCost(cost)
	tokensStr := util.FormatTokens(cost)

	return util.NewSegment("§", fmt.Sprintf("%s (%s)", costStr, tokensStr), "#00ffff", "#202020")
}

func contextWidget(claudeContext *claude.Context) *util.Segment {
	// Safely check context
	if claudeContext == nil || claudeContext.TokenMetrics == nil || claudeContext.TokenMetrics.ContextLength == 0 {
		return util.NewSegment("🧠", "0 ctx", "#ff00ff", "#202020")
	}

	// Show context length and percentage used
	ctxStr := util.FormatTokens(float64(claudeContext.TokenMetrics.ContextLength))

	// Estimate context window size based on model (default to 200k for Claude 3.5 Sonnet)
	var contextWindow int64 = 200000
	if claudeContext.Code != nil && claudeContext.Code.Model.ID != "" && strings.Contains(strings.ToLower(claudeContext.Code.Model.ID), "haiku") {
		contextWindow = 200000 // Claude 3 Haiku also has 200k context
	}

	percentage := float64(claudeContext.TokenMetrics.ContextLength) / float64(contextWindow) * 100

	return util.NewSegment("🧠", fmt.Sprintf("%s (%.0f%%)", ctxStr, percentage), "#ff00ff", "#202020")
}

func versionWidget(claudeContext *claude.Context) *util.Segment {
	if claudeContext.Code == nil || claudeContext.Code.Version == "" {
		return nil
	}
	return util.NewSegment("🔧", fmt.Sprintf("v%s", claudeContext.Code.Version), "#ffffff", "#666666")
}

func blockTimerWidget(claudeContext *claude.Context) *util.Segment {
	// Safely check context and block metrics
	if claudeContext == nil || claudeContext.BlockMetrics == nil || claudeContext.BlockMetrics.StartTime.IsZero() {
		return util.NewSegment("⏱️", "0hr 0m", "#ffff00", "#333333")
	}

	// Calculate elapsed time in 5-hour block
	elapsed := time.Since(claudeContext.BlockMetrics.StartTime)
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60

	var timeStr string
	if hours == 0 {
		timeStr = fmt.Sprintf("%dm", minutes)
	} else if minutes == 0 {
		timeStr = fmt.Sprintf("%dhr", hours)
	} else {
		timeStr = fmt.Sprintf("%dhr %dm", hours, minutes)
	}

	return util.NewSegment("⏱️", timeStr, "#ffff00", "#333333")
}
