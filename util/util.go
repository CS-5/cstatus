package util

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CS-5/cstatus/claude"
)

type StatuslineBuilder struct {
	claudeContext *claude.Context
	segments      []*Segment
}

func NewStatusLineBuilder(claudeContext *claude.Context) *StatuslineBuilder {
	return &StatuslineBuilder{
		claudeContext: claudeContext,
		segments:      []*Segment{},
	}
}

func (b *StatuslineBuilder) Append(render func(claudeContext *claude.Context) *Segment) *StatuslineBuilder {
	if segment := render(b.claudeContext); segment != nil {
		b.segments = append(b.segments, segment)
	}
	return b
}

func (b *StatuslineBuilder) Render() string {
	if len(b.segments) == 0 {
		return ""
	}

	var result strings.Builder
	for i, segment := range b.segments {
		if segment == nil || segment.IsEmpty() {
			continue
		}

		result.WriteString(segment.String())

		var next *Segment
		if i < len(b.segments)-1 {
			next = b.segments[i+1]
		}
		result.WriteString(segment.Sep(next))
	}

	return result.String()
}

const (
	asciiSeparatorRight = "\uE0B0"
	asciiSeparatorLeft  = "\uE0B2"
	asciiColorReset     = "\x1b[0m"
)

type Segment struct {
	icon  string
	text  string
	bgHex string
	fgHex string
}

func (s *Segment) IsEmpty() bool {
	return s == nil || (s.text == "" && s.icon == "")
}

func NewSegment(icon, text, fgColor, bgColor string) *Segment {
	return &Segment{
		icon:  icon,
		text:  text,
		bgHex: bgColor,
		fgHex: fgColor,
	}
}

func (s *Segment) String() string {
	return fmt.Sprintf("%s%s%s %s %s", s.BgColor(), s.FgColor(), s.icon, s.text, asciiColorReset)
}

func (s *Segment) BgColor() string {
	return hexToAnsi(s.bgHex, true)
}

func (s *Segment) FgColor() string {
	return hexToAnsi(s.fgHex, false)
}

func (s *Segment) Sep(next *Segment) string {
	sep := ""
	if next != nil {
		sep = next.BgColor()
	}
	return sep + hexToAnsi(s.bgHex, false) + asciiSeparatorRight + asciiColorReset
}

func hexToAnsi(hex string, background bool) string {
	// https://gist.github.com/fnky/458719343aabd01cfb17a3a4f7296797

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

func FormatCost(cost float64) string {
	if cost < 0.01 {
		return fmt.Sprintf("%.1f¢", cost*100)
	}
	return fmt.Sprintf("$%.2f", cost)
}

func FormatTokens(cost float64) string {
	tokens := int64(cost * 333333)
	if tokens > 1000000 {
		return fmt.Sprintf("%.1fM", float64(tokens)/1000000)
	}
	if tokens > 1000 {
		return fmt.Sprintf("%.1fK", float64(tokens)/1000)
	}
	return fmt.Sprintf("%d", tokens)
}
