package ui

import (
	"fmt"
	"strconv"
	"strings"
)

// Powerline symbols and icons
const (
	SeparatorRight    = "\uE0B0"
	SeparatorLeft     = "\uE0B2"
	Branch            = "⎇"
	ModelIcon         = "⚡"
	SessionIcon       = "§"
	ContextIcon       = "◔"
	GitChangesIcon    = "●"
	ProjectIcon       = "📁"
	ClockIcon         = "🕐"
	CostIcon          = "💰"
)

// Color constants - dark theme
const (
	ColorProjectBg    = "#8b4513"
	ColorProjectFg    = "#ffffff"
	ColorGitBg        = "#404040"
	ColorGitFg        = "#ffffff"
	ColorGitChangesBg = "#ff6b6b"
	ColorGitChangesFg = "#ffffff"
	ColorModelBg      = "#2d2d2d"
	ColorModelFg      = "#ffffff"
	ColorSessionBg    = "#202020"
	ColorSessionFg    = "#00ffff"
	ColorContextBg    = "#4a5568"
	ColorContextFg    = "#cbd5e0"
)

// Segment represents a rendered piece of the statusline
type Segment struct {
	Text  string
	BgHex string
	FgHex string
}

// IsEmpty returns true if the segment has no text
func (s *Segment) IsEmpty() bool {
	return s == nil || s.Text == ""
}

// CreateSegment is a helper function to create a segment with colors
func CreateSegment(text, bgColor, fgColor string) *Segment {
	return &Segment{
		Text:  text,
		BgHex: bgColor,
		FgHex: fgColor,
	}
}

// HexToAnsi converts hex color to ANSI escape sequence
func HexToAnsi(hex string, background bool) string {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return ""
	}

	r, _ := strconv.ParseInt(hex[0:2], 16, 0)
	g, _ := strconv.ParseInt(hex[2:4], 16, 0)
	b, _ := strconv.ParseInt(hex[4:6], 16, 0)

	escapeCode := "38"
	if background {
		escapeCode = "48"
	}
	return fmt.Sprintf("\x1b[%s;2;%d;%d;%dm", escapeCode, r, g, b)
}
