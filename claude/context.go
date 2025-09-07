package claude

import (
	"fmt"
	"io"
)

type Context struct {
	Code         *ClaudeCode
	TokenMetrics *ClaudeTokenMetrics
	BlockMetrics *ClaudeBlockMetrics
	WorkingDir   string
	ProjectName  string
}

func NewContextFromReader(r io.Reader) (*Context, error) {
	jsonData, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	// Handle empty input gracefully
	if len(jsonData) == 0 {
		return nil, fmt.Errorf("no input received")
	}

	code, err := unmarshalClaudeCodeInput(jsonData)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	var tokenMetrics *ClaudeTokenMetrics
	var blockMetrics *ClaudeBlockMetrics
	if code.TranscriptPath != "" {
		tokenMetrics, blockMetrics, err = parseMetrics(code.TranscriptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse metrics: %w", err)
		}
	}

	return &Context{
		Code:         code,
		TokenMetrics: tokenMetrics,
		BlockMetrics: blockMetrics,
		WorkingDir:   code.getWorkingDir(),
		ProjectName:  code.getProjectName(),
	}, nil
}
