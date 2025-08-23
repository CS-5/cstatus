# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

statusline is a high-performance, extensible statusline formatter for Claude Code CLI that displays model info, git branch, token usage, and other contextual metrics. The application is designed with extreme performance, ease of development, and maintainability as core priorities.

## Development Commands

```bash
# Run the application
cat test.json | go run main.go

# Build the application
go build main.go

# Run with test data
cat test.json | go run main.go

# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover

# Build for production
go build -o statusline main.go
```

## Architecture

The project follows a clean, layered architecture organized into focused packages:

### Package Structure

- **main.go**: CLI entry point - reads JSON from stdin and renders statusline
- **render/**: Data parsing, context creation, and JSONL transcript processing
- **ui/**: UI primitives (colors, icons, segments, ANSI rendering)  
- **widgets/**: Individual widget implementations
- **builder/**: Widget composition and statusline rendering

### Core Flow

```
JSON Input → render.NewRenderContext() → builder.New() → widgets.Render() → ANSI Output
```

### Core Components

**render/data.go**: Context management and data parsing
```go
// Creates context from Claude Code JSON string
func NewRenderContext(claudeJSON string) (*RenderContext, error)

// Rich context with pre-computed shared data
type RenderContext struct {
    Claude        *ClaudeCodeInput  // Parsed Claude Code JSON
    TokenMetrics  *TokenMetrics     // Parsed token usage from transcript
    GitBranch     string           // Cached git branch
    GitHasChanges bool             // Cached git status
    WorkingDir    string           // Resolved working directory
    ProjectName   string           // Derived project name
}
```

**render/jsonl.go**: JSONL transcript parsing for token metrics
```go
// Parses JSONL transcript file for token usage
func ParseTokenMetrics(transcriptPath string) *TokenMetrics

// Calculates session duration from timestamps
func GetSessionDuration(transcriptPath string) string
```

**ui/ui.go**: UI primitives and rendering
```go
// Rendered output segment
type Segment struct {
    Text  string  // Display text with icons
    BgHex string  // Background color hex
    FgHex string  // Foreground color hex
}

// Constants for icons and colors
const (
    ProjectIcon = "📁"
    Branch = "⎇"
    ModelIcon = "⚡"
    // ... colors and other icons
)
```

**widgets/widget.go**: Widget interface
```go
// Simple widget interface - every widget implements this
type Widget interface {
    Render(ctx *render.RenderContext) *ui.Segment
}
```

**builder/builder.go**: Widget composition and rendering
```go
// Registry-based widget system
var WidgetRegistry = map[string]func() widgets.Widget{
    "project": widgets.NewProject,
    "git":     widgets.NewGit,
    // ...
}
```

### Performance Optimizations

1. **Pre-computed Context**: All expensive operations (git commands, file parsing) done once in `NewRenderContext()`
2. **Minimal Interface**: Single `Render()` method with no overhead
3. **Registry Pattern**: Widgets registered once at startup
4. **JSONL Parsing**: Efficient streaming parser for transcript files
5. **Conditional Rendering**: Widgets return `nil` when they shouldn't display

## Adding New Widgets

Adding a widget is straightforward:

1. **Create the widget file** in `widgets/`:

```go
package widgets

import (
    "fmt"
    "github.com/CS-5/statusline/render"
    "github.com/CS-5/statusline/ui"
)

type MyWidget struct{}

func NewMyWidget() Widget {
    return &MyWidget{}
}

func (w *MyWidget) Render(ctx *render.RenderContext) *ui.Segment {
    // Return nil if widget shouldn't display
    if someCondition {
        return nil
    }
    
    text := fmt.Sprintf(" %s %s ", ui.MyIcon, value)
    return ui.CreateSegment(text, ui.MyBgColor, ui.MyFgColor)
}
```

2. **Add constants** to `ui/ui.go`:

```go
const (
    MyIcon = "🔧"
    MyBgColor = "#123456"
    MyFgColor = "#ffffff"
)
```

3. **Register the widget** in `builder/builder.go`:

```go
var WidgetRegistry = map[string]func() widgets.Widget{
    "my-widget": widgets.NewMyWidget,
    // ... existing widgets
}
```

4. **Add to default order** in `builder.New()`:

```go
func New() *StatuslineBuilder {
    return &StatuslineBuilder{
        widgetNames: []string{"project", "git", "model", "my-widget", "session", "context"},
    }
}
```

5. **Write tests** in `widgets/my_widget_test.go`:

```go
func TestMyWidget_Render(t *testing.T) {
    widget := NewMyWidget()
    ctx := &render.RenderContext{/* test data */}
    
    segment := widget.Render(ctx)
    // assertions...
}
```

## Widget Development Patterns

### Conditional Rendering

```go
func (w *MyWidget) Render(ctx *render.RenderContext) *ui.Segment {
    if !shouldDisplay {
        return nil  // Widget won't appear in statusline
    }
    // ... render logic
}
```

### Data Access

```go
func (w *MyWidget) Render(ctx *render.RenderContext) *ui.Segment {
    // Access Claude data
    modelName := ctx.Claude.Model.DisplayName
    cost := ctx.Claude.Cost.TotalCostUSD
    
    // Access pre-computed context
    branch := ctx.GitBranch
    hasChanges := ctx.GitHasChanges
    
    // Access token metrics from transcript
    if ctx.TokenMetrics != nil {
        tokens := ctx.TokenMetrics.TotalTokens
        contextLength := ctx.TokenMetrics.ContextLength
    }
}
```

### Color and Icon Management

```go
// Define in ui/ui.go
const (
    MyIcon = "🔧"
    MyBgColor = "#123456" 
    MyFgColor = "#ffffff"
)

// Use in widget
text := fmt.Sprintf(" %s %s ", ui.MyIcon, value)
return ui.CreateSegment(text, ui.MyBgColor, ui.MyFgColor)
```

## JSONL Transcript Integration

The application can parse Claude Code transcript files to extract rich token usage data:

```go
// Token metrics extracted from transcript
type TokenMetrics struct {
    InputTokens   int64  // Total input tokens used
    OutputTokens  int64  // Total output tokens generated
    CachedTokens  int64  // Total cached tokens (read + creation)
    TotalTokens   int64  // Sum of all token types
    ContextLength int64  // Current context window usage
}

// Usage in widgets
func (w *ContextWidget) Render(ctx *render.RenderContext) *ui.Segment {
    if ctx.TokenMetrics != nil && ctx.TokenMetrics.ContextLength > 0 {
        text := fmt.Sprintf(" %s %d tokens ", ui.ContextIcon, ctx.TokenMetrics.ContextLength)
        return ui.CreateSegment(text, ui.ColorContextBg, ui.ColorContextFg)
    }
    return nil
}
```

## Testing Guidelines

- **Unit test every widget** with various input conditions
- **Test conditional rendering** (nil returns)
- **Test the builder** with different widget combinations
- **Test context creation** with various JSON inputs
- **Test JSONL parsing** with mock transcript files

### Test Structure

```go
func TestMyWidget_Render(t *testing.T) {
    tests := []struct {
        name      string
        context   *render.RenderContext
        expectNil bool
        expectText string
    }{
        {"normal case", &render.RenderContext{...}, false, " 🔧 value "},
        {"hidden case", &render.RenderContext{...}, true, ""},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            widget := NewMyWidget()
            segment := widget.Render(tt.context)
            
            if tt.expectNil {
                if segment != nil {
                    t.Error("Expected nil segment")
                }
                return
            }
            
            // Assert segment properties...
        })
    }
}
```

## Configuration

Currently, configuration happens at compile-time through:

1. **Widget order** in `builder.New()`
2. **Colors and icons** in `ui/ui.go` constants
3. **Widget registry** in `builder/builder.go`

Future: Runtime configuration could be added by reading from config files or environment variables.

## Performance Guidelines

- **Avoid I/O in widgets**: Use pre-computed data from `RenderContext`
- **Cache expensive operations**: Add to `NewRenderContext()` if needed by multiple widgets
- **Return early**: Return `nil` immediately for widgets that shouldn't display
- **Minimize allocations**: Reuse strings and avoid unnecessary formatting
- **JSONL streaming**: The transcript parser streams data efficiently

## Debugging

Test context creation:
```go
jsonData := `{"model":{"display_name":"Test"},"transcript_path":"/path/to/transcript.jsonl"}`
ctx, err := render.NewRenderContext(jsonData)
```

Test widgets in isolation:
```go
widget := widgets.NewMyWidget()
segment := widget.Render(ctx)
```

Test full statusline:
```go
builder := builder.New()
output := builder.Build(ctx)
```

## Common Patterns

### Value Formatting
```go
func (w *MyWidget) formatValue(value float64) string {
    if value > 1000 {
        return fmt.Sprintf("%.1fK", value/1000)
    }
    return fmt.Sprintf("%.0f", value)
}
```

### Icon Selection
```go
func (w *MyWidget) getIcon(status string) string {
    switch status {
    case "good":
        return "✅"
    case "warning":
        return "⚠️"
    default:
        return "❓"
    }
}
```

### Color Selection
```go
func (w *MyWidget) getColors(status string) (bg, fg string) {
    switch status {
    case "error":
        return "#ff0000", "#ffffff"
    case "warning":
        return "#ffaa00", "#000000"
    default:
        return "#00ff00", "#000000"
    }
}
```

## Extension Points

The architecture supports easy extension:

1. **Custom data sources**: Add fields to `RenderContext` and parsing to `NewRenderContext()`
2. **JSONL data**: Extend transcript parsing for additional metrics
3. **Custom rendering**: Implement complex rendering in `StatuslineBuilder`
4. **Multiple layouts**: Create different builders for different use cases
5. **Plugin system**: Could be added via Go plugins or configuration

## Best Practices

- **Keep widgets focused**: One responsibility per widget
- **Test thoroughly**: Include edge cases and error conditions
- **Document behavior**: Clear comments for complex logic
- **Follow patterns**: Match existing code style and structure
- **Performance first**: Always consider the impact of changes
- **Package boundaries**: Keep UI, data, and business logic separate