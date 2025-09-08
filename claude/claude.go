package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
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

// SessionDuration represents the session duration in milliseconds (5 hours)
const sessionDurationMs = int64(5 * 60 * 60 * 1000)

func parseMetrics(transcriptPath string) (*ClaudeTokenMetrics, *ClaudeBlockMetrics, error) {
	// Parses JSONL transcript file to extract token usage and session metrics

	if transcriptPath == "" {
		return nil, nil, nil
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		// Return nil metrics instead of failing - transcript may not exist yet
		if !os.IsNotExist(err) {
			log.Printf("Warning: failed to open transcript file %s: %v", transcriptPath, err)
		}
		return nil, nil, nil
	}
	defer file.Close()

	var inputTokens, outputTokens, cachedTokens, contextLength int64
	var mostRecentMainChainEntry *TranscriptEntry
	var mostRecentTimestamp time.Time

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry TranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			// Log parsing errors for debugging, but continue processing
			log.Printf("Warning: failed to parse transcript line: %v", err)
			continue
		}

		// Parse token usage data
		if entry.Message != nil && entry.Message.Usage != nil {
			usage := entry.Message.Usage
			inputTokens += usage.InputTokens
			outputTokens += usage.OutputTokens
			cachedTokens += usage.CacheReadInputTokens + usage.CacheCreationInputTokens

			// Track the most recent main chain entry for context length
			// Main chain entries have isSidechain = false or undefined (defaults to main chain)
			if !entry.IsSidechain && entry.Timestamp != "" {
				if entryTime, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
					if mostRecentTimestamp.IsZero() || entryTime.After(mostRecentTimestamp) {
						mostRecentTimestamp = entryTime
						mostRecentMainChainEntry = &entry
					}
				} else {
					log.Printf("Warning: failed to parse timestamp %s: %v", entry.Timestamp, err)
				}
			}
		}
	}

	// Calculate context length from the most recent main chain message
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

	// Parse block metrics from the same file to avoid duplicate I/O
	blockMetrics, err := parseBlockMetricsFromFile(file)
	if err != nil {
		log.Printf("Warning: failed to parse block metrics: %v", err)
		blockMetrics = nil
	}

	return tokenMetrics, blockMetrics, nil
}

// parseBlockMetricsFromFile parses block metrics from an already open file
func parseBlockMetricsFromFile(file *os.File) (*ClaudeBlockMetrics, error) {
	// Reset file pointer to beginning
	if _, err := file.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("failed to seek to start of file: %v", err)
	}

	timestamps, err := extractTimestamps(file)
	if err != nil {
		return nil, err
	}

	if len(timestamps) == 0 {
		return nil, nil
	}

	return calculateBlockMetrics(timestamps), nil
}

// extractTimestamps efficiently extracts and sorts timestamps from transcript
func extractTimestamps(file *os.File) ([]time.Time, error) {
	var timestamps []time.Time
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

		if entry.Timestamp != "" {
			if t, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
				timestamps = append(timestamps, t)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning file: %v", err)
	}

	// Use efficient sort instead of bubble sort
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	return timestamps, nil
}

// calculateBlockMetrics computes block metrics from sorted timestamps
func calculateBlockMetrics(timestamps []time.Time) *ClaudeBlockMetrics {
	if len(timestamps) == 0 {
		return nil
	}

	now := time.Now()
	mostRecentTimestamp := timestamps[len(timestamps)-1]

	// Check if the most recent activity is within the current session period
	if now.Sub(mostRecentTimestamp).Milliseconds() > sessionDurationMs {
		return nil // No recent activity
	}

	// Find the start of the current continuous work period
	continuousWorkStart := findContinuousWorkStart(timestamps, sessionDurationMs)

	// Floor to the hour
	flooredWorkStart := floorToHour(continuousWorkStart)

	// Calculate current block within the work period
	blockStart := calculateBlockStart(now, flooredWorkStart, sessionDurationMs)

	return &ClaudeBlockMetrics{
		StartTime:    blockStart,
		LastActivity: mostRecentTimestamp,
	}
}

// findContinuousWorkStart finds the start of continuous work period
func findContinuousWorkStart(timestamps []time.Time, sessionDurationMs int64) time.Time {
	if len(timestamps) == 0 {
		return time.Time{}
	}

	continuousWorkStart := timestamps[len(timestamps)-1]
	for i := len(timestamps) - 2; i >= 0; i-- {
		gap := timestamps[i+1].Sub(timestamps[i]).Milliseconds()
		if gap >= sessionDurationMs {
			break // Found a session boundary
		}
		continuousWorkStart = timestamps[i]
	}
	return continuousWorkStart
}

// floorToHour floors a timestamp to the hour boundary
func floorToHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

// calculateBlockStart determines the start time of the current block
func calculateBlockStart(now, flooredWorkStart time.Time, sessionDurationMs int64) time.Time {
	totalWorkTime := now.Sub(flooredWorkStart).Milliseconds()
	if totalWorkTime > sessionDurationMs {
		completedBlocks := totalWorkTime / sessionDurationMs
		blockStartMs := flooredWorkStart.UnixMilli() + (completedBlocks * sessionDurationMs)
		return time.UnixMilli(blockStartMs)
	}
	return flooredWorkStart
}

// getBlockMetrics maintains backward compatibility by wrapping the new implementation
func getBlockMetrics(transcriptPath string) *ClaudeBlockMetrics {
	if transcriptPath == "" {
		return nil
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("Warning: failed to open transcript file %s: %v", transcriptPath, err)
		}
		return nil
	}
	defer file.Close()

	blockMetrics, err := parseBlockMetricsFromFile(file)
	if err != nil {
		log.Printf("Warning: failed to parse block metrics from %s: %v", transcriptPath, err)
		return nil
	}

	return blockMetrics
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
