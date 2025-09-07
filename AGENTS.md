# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

cstatus is a high-performance, extensible statusline formatter for Claude Code CLI that displays model info, git branch, token usage, and other contextual metrics. The application is designed with extreme performance, ease of development, and maintainability as core priorities. It's written in Go as a compiled binary alternative to JavaScript-based statusline tools.

## Installation

The application includes an `install` command that automatically configures Claude Code to use cstatus as the statusline:

```bash
# Install the binary
go install github.com/CS-5/cstatus@latest

# Configure Claude Code to use cstatus
cstatus install
```

The install command:
- Creates `~/.claude/settings.json` if it doesn't exist
- Adds the statusLine configuration to use cstatus
- Preserves existing settings while updating the statusline configuration
- Shows an error if a different statusline is already configured (to prevent conflicts)

### Manual Installation

If you prefer to configure manually, add this to `~/.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "cstatus",
    "padding": 0
  }
}
```

## Development Commands

```bash
# Install the application
go install github.com/CS-5/cstatus@latest

# Install as Claude Code statusline
cstatus install

# Run the application locally (for testing)
echo '{"model":{"display_name":"Test"}}' | go run main.go

# Build the application  
go build -o cstatus main.go

# Run all tests
go test ./... -v

# Run tests with coverage
go test ./... -cover
```

## Architecture

The project follows a clean, layered architecture with explicit widget composition:

### Package Structure

- **main.go**: CLI entry point with installation support - explicitly configures widgets and renders statusline using function-based approach
- **widgets.go**: Widget functions for project, git, model, session, context, and block timer display
- **claude/**: Claude Code JSON parsing, context creation, and JSONL transcript processing
- **util/**: StatuslineBuilder for composing widget functions, UI segments, ANSI rendering, and powerline separators

### Core Flow

```
JSON Input → claude.NewContextFromReader() → util.NewStatusLineBuilder().Append(...) → widget functions → ANSI Output
```

### Key Design Principles

1. **Function-Based Widgets**: Simple functions that take `*claude.Context` and return `*ui.Segment`
2. **Explicit Configuration**: No "default widgets" - main.go explicitly declares what widgets to use
3. **Type Safety**: Compile-time verification of widget composition via function signatures
4. **Simplified Rendering**: Helper functions reduce boilerplate in widgets

### Core Components

**claude/context.go**: Context management and data parsing
```go
// Creates context from Claude Code JSON input with error handling
func NewContextFromReader(r io.Reader) (*Context, error)

// Rich context with pre-computed shared data
type Context struct {
    Code         *ClaudeCode          // Parsed Claude Code JSON
    TokenMetrics *ClaudeTokenMetrics  // Parsed token usage from transcript
    BlockMetrics *ClaudeBlockMetrics  // Block timing for usage allocation
    WorkingDir   string               // Resolved working directory
    ProjectName  string               // Derived project name
}
```

**claude/claude.go**: JSONL transcript parsing for token metrics
```go
// Parses JSONL transcript file for token usage
func parseTokenMetrics(transcriptPath string) (*ClaudeTokenMetrics, error)

// Calculates session duration from timestamps
func GetSessionDuration(transcriptPath string) string
```

**util/util.go**: UI primitives and rendering helpers
```go
// Rendered output segment with powerline separators
type Segment struct {
    icon  string  // Icon to display
    text  string  // Text content
    bgHex string  // Background color hex
    fgHex string  // Foreground color hex
}

// Segment creation with icon and content
func NewSegment(icon, text, fgColor, bgColor string) *Segment

// Powerline separator rendering
func (s *Segment) Sep(next *Segment) string

// Cost and token formatting utilities
func FormatCost(cost float64) string
func FormatTokens(cost float64) string
```

**util/util.go**: Function-based widget composition and UI utilities
```go
// StatuslineBuilder for composing widget functions
type StatuslineBuilder struct {
    claudeContext *claude.Context
    segments      []*ui.Segment
}

// Append widget functions that return segments
func (b *StatuslineBuilder) Append(render func(claudeContext *claude.Context) *ui.Segment) *StatuslineBuilder
```

### Performance Optimizations

1. **Pre-computed Context**: All expensive operations (git commands, file parsing) done once in `NewContextFromReader()`
2. **Function-Based**: Simple function calls eliminate interface overhead
3. **Simplified Rendering**: Reduced ANSI escape sequence complexity with built-in powerline separators
4. **JSONL Streaming**: Efficient transcript parsing
5. **Conditional Rendering**: Widget functions return `nil` when they shouldn't display

## Adding New Widgets

The project uses a function-based widget approach for simplicity and compactness. Adding a widget is straightforward:

1. **Create a widget function** in `widgets.go`:

```go
func myWidget(claudeContext *claude.Context) *ui.Segment {
    // Return nil if widget shouldn't display
    if someCondition {
        return nil
    }
    
    // Use helper for consistent formatting
    return util.NewSegment(myIcon, content, fgColor, bgColor)
}
```

2. **Add constants** (if needed):

```go
// Icons and colors can be defined as constants
const (
    MyIcon = "🔧"
    MyBgColor = "#123456"
    MyFgColor = "#ffffff"
)
```

3. **Add widget to main.go**:

```go
fmt.Println(util.NewStatusLineBuilder(claudeContext).
    Append(projectWidget).
    Append(myWidget). // Add your widget
    Append(gitStatusWidget).
    Append(sessionWidget).
    Append(contextWidget).
    Append(blockTimerWidget).
    Render(),
)
```

4. **Write tests** (if needed):

```go
func TestMyWidget(t *testing.T) {
    ctx := &claude.Context{/* test data */}
    
    segment := myWidget(ctx)
    // assertions...
}
```

## Widget Development Patterns

### Conditional Rendering

```go
func myWidget(claudeContext *claude.Context) *ui.Segment {
    if !shouldDisplay {
        return nil  // Widget won't appear in statusline
    }
    // ... render logic
}
```

### Using UI Helpers

```go
// Simple icon + content
return ui.NewSegment(ui.MyIcon, content, ui.ColorBg, ui.ColorFg)

// Complex content with formatting
content := fmt.Sprintf("%s (%s)", primary, secondary) 
return ui.NewSegment(ui.MyIcon, content, ui.ColorBg, ui.ColorFg)

// Manual text control
text := fmt.Sprintf(" %s Custom Format %s ", icon, value)
return ui.CreateSegment(text, ui.ColorBg, ui.ColorFg)
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

### Dynamic Styling

```go
func (w *GitWidget) Render(ctx *render.RenderContext) *ui.Segment {
    if ctx.GitBranch == "" {
        return nil
    }

    // Change colors based on git status
    if ctx.GitHasChanges {
        content := fmt.Sprintf("%s %s", ctx.GitBranch, ui.GitChangesIcon)
        return ui.NewSegment(ui.Branch, content, ui.ColorGitChangesBg, ui.ColorGitChangesFg)
    }
    
    return ui.NewSegment(ui.Branch, ctx.GitBranch, ui.ColorGitBg, ui.ColorGitFg)
}
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
        content := fmt.Sprintf("%d tokens", ctx.TokenMetrics.ContextLength)
        return ui.NewSegment(ui.ContextIcon, content, ui.ColorContextBg, ui.ColorContextFg)
    }
    return nil
}
```

## Configuration

Widget configuration is explicit and happens at compile-time in `main.go`:

```go
// Current widget configuration
fmt.Println(util.NewStatusLineBuilder(claudeContext).
    Append(projectWidget).
    Append(gitStatusWidget).
    Append(sessionWidget).
    Append(contextWidget).
    Append(blockTimerWidget).
    Render(),
)
```

Colors and icons are defined directly in each widget function in `widgets.go`.

### Available Widgets

- `projectWidget`: Shows project name from workspace
- `gitStatusWidget`: Shows current git branch  
- `sessionWidget`: Shows cost and estimated token usage
- `contextWidget`: Shows context length and percentage of window used
- `blockTimerWidget`: Shows elapsed time since session start
- `modelWidget`: Shows Claude model name (not currently used)
- `versionWidget`: Shows Claude Code version (not currently used)

Future: Runtime configuration could be added by reading from config files or environment variables.

## Testing Guidelines

- **Unit test every widget** with various input conditions
- **Test conditional rendering** (nil returns)
- **Test the builder** with different widget combinations
- **Test context creation** with various JSON inputs
- **Test JSONL parsing** with mock transcript files
- **Test error conditions** in render context creation

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
        {"error case", &render.RenderContext{...}, true, ""},
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

## Performance Guidelines

- **Avoid I/O in widgets**: Use pre-computed data from `RenderContext`
- **Cache expensive operations**: Add to `NewRenderContext()` if needed by multiple widgets
- **Return early**: Return `nil` immediately for widgets that shouldn't display
- **Use UI helpers**: `NewSegment()` reduces allocations and formatting overhead
- **JSONL streaming**: The transcript parser streams data efficiently

## Debugging

Test context creation:
```go
jsonData := `{"model":{"display_name":"Test"},"transcript_path":"/path/to/transcript.jsonl"}`
ctx, err := render.NewRenderContext(jsonData)
if err != nil {
    log.Fatal(err) // Better error messages now available
}
```

Test widgets in isolation:
```go
widget := widgets.NewMyWidget()
segment := widget.Render(ctx)
```

Test full statusline:
```go
statusline := builder.New().
    AppendWidget(widgets.NewProject()).
    AppendWidget(widgets.NewMyWidget()).
    Build(ctx)
```

## Error Handling

The refactored code includes improved error handling:

- **JSON Parsing**: Clear error messages for malformed Claude Code JSON
- **Empty Input**: Explicit handling of empty JSON input
- **Type Safety**: Compile-time verification eliminates string-based registry errors
- **Nil Safety**: Consistent nil checking in widgets and rendering

## Common Patterns

### Value Formatting
```go
func (w *SessionWidget) formatCost(cost float64) string {
    if cost < 0.01 {
        return fmt.Sprintf("%.1f¢", cost*100)
    }
    return fmt.Sprintf("$%.2f", cost)
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

1. **Custom widgets**: Add via `AppendWidget()` in main.go
2. **JSONL data**: Extend transcript parsing for additional metrics
3. **UI helpers**: Add new segment creation functions to ui package
4. **Dynamic configuration**: Could read widget selection from config files
5. **Plugin system**: Could be added via Go plugins

## Best Practices

- **Explicit over implicit**: No hidden default widgets or string registries
- **Type safety first**: Use factories and interfaces over string lookups
- **Keep widgets focused**: One responsibility per widget
- **Test thoroughly**: Include edge cases and error conditions  
- **Use UI helpers**: Consistent formatting via `NewSegment()` and `CreateSegment()`
- **Performance first**: Always consider the impact of changes
- **Package boundaries**: Keep UI, data, and business logic separate
- **Error handling**: Provide clear, actionable error messages