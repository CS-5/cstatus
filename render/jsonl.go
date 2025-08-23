package render

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// TokenMetrics represents parsed token usage from transcript
type TokenMetrics struct {
	InputTokens   int64 `json:"inputTokens"`
	OutputTokens  int64 `json:"outputTokens"`
	CachedTokens  int64 `json:"cachedTokens"`
	TotalTokens   int64 `json:"totalTokens"`
	ContextLength int64 `json:"contextLength"`
}

// TranscriptEntry represents a single line in the transcript JSONL
type TranscriptEntry struct {
	Timestamp   string    `json:"timestamp,omitempty"`
	IsSidechain bool      `json:"isSidechain,omitempty"`
	Message     *Message  `json:"message,omitempty"`
}

// Message represents the message structure in transcript entries
type Message struct {
	Usage *Usage `json:"usage,omitempty"`
}

// Usage represents token usage data
type Usage struct {
	InputTokens              int64 `json:"input_tokens,omitempty"`
	OutputTokens             int64 `json:"output_tokens,omitempty"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens,omitempty"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens,omitempty"`
}

// BlockMetrics represents 5-hour block timing information
type BlockMetrics struct {
	StartTime    time.Time `json:"startTime"`
	LastActivity time.Time `json:"lastActivity"`
}

// ParseTokenMetrics parses a JSONL transcript file and extracts token metrics
func ParseTokenMetrics(transcriptPath string) *TokenMetrics {
	if transcriptPath == "" {
		return &TokenMetrics{}
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return &TokenMetrics{}
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
			continue // Skip invalid JSON lines
		}

		if entry.Message != nil && entry.Message.Usage != nil {
			usage := entry.Message.Usage
			inputTokens += usage.InputTokens
			outputTokens += usage.OutputTokens
			cachedTokens += usage.CacheReadInputTokens + usage.CacheCreationInputTokens

			// Track the most recent main chain entry
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

	return &TokenMetrics{
		InputTokens:   inputTokens,
		OutputTokens:  outputTokens,
		CachedTokens:  cachedTokens,
		TotalTokens:   totalTokens,
		ContextLength: contextLength,
	}
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