package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/CS-5/cstatus/claude"
	"github.com/CS-5/cstatus/util"
)

func projectWidget(claudeContext *claude.Context) *util.Segment {
	if claudeContext.ProjectName == "" {
		return nil
	}
	return util.NewSegment("", claudeContext.ProjectName, "#ffffff", "#8b4513")
}

func gitStatusWidget(claudeContext *claude.Context) *util.Segment {
	if claudeContext == nil || claudeContext.WorkingDir == "" {
		return nil
	}

	gitDir := filepath.Join(claudeContext.WorkingDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil
	}

	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = claudeContext.WorkingDir

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = claudeContext.WorkingDir

	output, err := cmd.Output()
	if err != nil {
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
	if claudeContext == nil || claudeContext.Code == nil {
		return nil
	}

	cost := claudeContext.Code.Cost.TotalCostUSD
	costStr := util.FormatCost(cost)
	tokensStr := util.FormatTokens(cost)

	return util.NewSegment("§", fmt.Sprintf("%s (%s)", costStr, tokensStr), "#00ffff", "#202020")
}


func contextWidget(claudeContext *claude.Context) *util.Segment {
	if claudeContext == nil || claudeContext.TokenMetrics == nil || claudeContext.TokenMetrics.ContextLength == 0 {
		return nil
	}

	ctxStr := util.FormatTokens(float64(claudeContext.TokenMetrics.ContextLength))

	// All current Claude models have 200k context window according to reference implementation
	contextWindow := int64(200000)
	percentage := float64(claudeContext.TokenMetrics.ContextLength) / float64(contextWindow) * 100

	return util.NewSegment("🧠", fmt.Sprintf("%s (%.1f%%)", ctxStr, percentage), "#ff00ff", "#202020")
}

func versionWidget(claudeContext *claude.Context) *util.Segment {
	if claudeContext.Code == nil || claudeContext.Code.Version == "" {
		return nil
	}
	return util.NewSegment("🔧", fmt.Sprintf("v%s", claudeContext.Code.Version), "#ffffff", "#666666")
}

func blockTimerWidget(claudeContext *claude.Context) *util.Segment {
	// Return nil when no active block - similar to reference implementation
	if claudeContext == nil || claudeContext.BlockMetrics == nil || claudeContext.BlockMetrics.StartTime.IsZero() {
		return nil
	}

	elapsed := time.Since(claudeContext.BlockMetrics.StartTime)
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60

	var timeStr string
	if hours == 0 && minutes == 0 {
		timeStr = "0m"
	} else if hours == 0 {
		timeStr = fmt.Sprintf("%dm", minutes)
	} else if minutes == 0 {
		timeStr = fmt.Sprintf("%dhr", hours)
	} else {
		timeStr = fmt.Sprintf("%dhr %dm", hours, minutes)
	}

	return util.NewSegment("⏱️", timeStr, "#ffff00", "#333333")
}
