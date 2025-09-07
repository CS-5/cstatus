package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ClaudeCode represents the input from Claude Code to the statusline application.
// Claude Code sends a JSON string via stdin whenever the statusline is expected to perform an update.
// https://docs.anthropic.com/en/docs/claude-code/statusline
type ClaudeCode struct {
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
	TotalLinesAdded    int64   `json:"total_lines_added"`
	TotalLinesRemoved  int64   `json:"total_lines_removed"`
}

func unmarshalClaudeCodeInput(data []byte) (*ClaudeCode, error) {
	var input ClaudeCode
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, err
	}

	// Set defaults for missing/empty critical fields
	if input.Model.DisplayName == "" {
		input.Model.DisplayName = "Unknown Model"
	}
	if input.Model.ID == "" {
		input.Model.ID = "unknown"
	}
	if input.Cwd == "" {
		if cwd, err := os.Getwd(); err == nil {
			input.Cwd = cwd
		}
	}
	if input.Version == "" {
		input.Version = "unknown"
	}
	if input.OutputStyle.Name == "" {
		input.OutputStyle.Name = "default"
	}

	return &input, nil
}

func (c *ClaudeCode) getWorkingDir() string {
	if c.Workspace.CurrentDir != "" {
		return c.Workspace.CurrentDir
	}
	return c.Cwd
}

func (c *ClaudeCode) getProjectName() string {
	if c.Workspace.ProjectDir != "" {
		return filepath.Base(c.Workspace.ProjectDir)
	}
	return ""
}

// TranscriptEntry represents a single entry in the session's transcript. Claude
// Code transscripts are stored in JSONL format (newline delimited distinct JSON objects).
type TranscriptEntry struct {
	Timestamp   string   `json:"timestamp,omitempty"`
	IsSidechain bool     `json:"isSidechain,omitempty"`
	Message     *Message `json:"message,omitempty"`
}

type Message struct {
	Usage *Usage `json:"usage,omitempty"`
}

type Usage struct {
	InputTokens              int64 `json:"input_tokens,omitempty"`
	OutputTokens             int64 `json:"output_tokens,omitempty"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens,omitempty"`
}

type ClaudeTokenMetrics struct {
	InputTokens   int64 `json:"inputTokens"`
	OutputTokens  int64 `json:"outputTokens"`
	CachedTokens  int64 `json:"cachedTokens"`
	TotalTokens   int64 `json:"totalTokens"`
	ContextLength int64 `json:"contextLength"`
}

type ClaudeBlockMetrics struct {
	StartTime    time.Time `json:"startTime"`
	LastActivity time.Time `json:"lastActivity"`
}

func parseMetrics(transcriptPath string) (*ClaudeTokenMetrics, *ClaudeBlockMetrics, error) {
	// This function is vibe coded - I take no responsibility

	if transcriptPath == "" {
		return nil, nil, nil
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		// Return nil metrics instead of failing - transcript may not exist yet
		return nil, nil, nil
	}
	defer file.Close()

	var inputTokens, outputTokens, cachedTokens, contextLength int64
	var mostRecentMainChainEntry *TranscriptEntry
	var mostRecentTimestamp time.Time
	var firstTimestamp time.Time

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid JSON lines
		}

		// Parse timestamp for block metrics (first valid timestamp)
		if entry.Timestamp != "" && firstTimestamp.IsZero() {
			if t, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
				firstTimestamp = t
			}
		}

		// Parse token usage data
		if entry.Message != nil && entry.Message.Usage != nil {
			usage := entry.Message.Usage
			inputTokens += usage.InputTokens
			outputTokens += usage.OutputTokens
			cachedTokens += usage.CacheReadInputTokens + usage.CacheCreationInputTokens

			// Track the most recent main chain entry for context length
			if !entry.IsSidechain && entry.Timestamp != "" {
				if entryTime, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
					if mostRecentTimestamp.IsZero() || entryTime.After(mostRecentTimestamp) {
						mostRecentTimestamp = entryTime
						mostRecentMainChainEntry = &entry
					}
				}
			}
		}
	}

	// Calculate context length from most recent main chain message
	if mostRecentMainChainEntry != nil && mostRecentMainChainEntry.Message != nil && mostRecentMainChainEntry.Message.Usage != nil {
		usage := mostRecentMainChainEntry.Message.Usage
		contextLength = usage.InputTokens + usage.CacheReadInputTokens + usage.CacheCreationInputTokens
	}

	totalTokens := inputTokens + outputTokens + cachedTokens

	tokenMetrics := &ClaudeTokenMetrics{
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		CachedTokens:  cachedTokens,
		TotalTokens:   totalTokens,
		ContextLength: contextLength,
	}

	var blockMetrics *ClaudeBlockMetrics
	if !firstTimestamp.IsZero() {
		blockMetrics = &ClaudeBlockMetrics{
			StartTime:    firstTimestamp,
			LastActivity: time.Now(),
		}
	}

	return tokenMetrics, blockMetrics, nil
}

// GetSessionDuration calculates session duration from transcript timestamps
func GetSessionDuration(transcriptPath string) string {
	if transcriptPath == "" {
		return ""
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	var firstTimestamp, lastTimestamp time.Time
	scanner := bufio.NewScanner(file)

	// Find first valid timestamp
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
				firstTimestamp = t
				break
			}
		}
	}

	// Find last valid timestamp by reading the entire file
	file.Seek(0, 0) // Reset file pointer
	scanner = bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		if entry.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
				lastTimestamp = t
			}
		}
	}

	if firstTimestamp.IsZero() || lastTimestamp.IsZero() {
		return ""
	}

	duration := lastTimestamp.Sub(firstTimestamp)
	totalMinutes := int(duration.Minutes())

	if totalMinutes < 1 {
		return "<1m"
	}

	hours := totalMinutes / 60
	minutes := totalMinutes % 60

	if hours == 0 {
		return formatDuration(0, minutes)
	} else if minutes == 0 {
		return formatDuration(hours, 0)
	} else {
		return formatDuration(hours, minutes)
	}
}

func formatDuration(hours, minutes int) string {
	if hours == 0 {
		return fmt.Sprintf("%dm", minutes)
	} else if minutes == 0 {
		return fmt.Sprintf("%dhr", hours)
	} else {
		return fmt.Sprintf("%dhr %dm", hours, minutes)
	}
}
