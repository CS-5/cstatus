package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/CS-5/cstatus/claude"
	"github.com/CS-5/cstatus/util"
)

func main() {
	if runtime.GOOS == "windows" {
		fmt.Fprintf(os.Stderr, "Error: cstatus is not supported on Windows.\n")
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == "install" {
		if err := handleInstall(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Check if there's piped input; if not, show usage
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		fmt.Fprintf(os.Stderr, "Error: No input received. Either pipe JSON input or use 'install' command.\n")
		fmt.Fprintf(os.Stderr, "\nUsage:\n")
		fmt.Fprintf(os.Stderr, "  echo '{\"model\":{\"display_name\":\"Claude\"}}' | %s\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s install\n", os.Args[0])
		os.Exit(1)
	}

	claudeContext, err := claude.NewContextFromReader(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Claude context: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(util.NewStatusLineBuilder(claudeContext).
		Append(projectWidget).
		Append(gitStatusWidget).
		Append(sessionWidget).
		Append(contextWidget).
		Append(blockTimerWidget).
		Render(),
	)
}

func handleInstall() error {
	fmt.Println("Installing cstatus as Claude Code statusline...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}

	claudeDir := filepath.Join(homeDir, ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")

	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("could not create .claude directory: %w", err)
	}

	// Read existing settings or create empty object
	var settings map[string]any
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			fmt.Printf("Warning: Could not parse existing settings.json, creating new one\n")
			settings = make(map[string]any)
		}
	} else {
		settings = make(map[string]any)
	}

	// Check if statusline is already configured
	if existingStatusline, ok := settings["statusLine"]; ok {
		if statuslineMap, ok := existingStatusline.(map[string]any); ok {
			if command, ok := statuslineMap["command"].(string); ok {
				if strings.Contains(command, "cstatus") {
					fmt.Println("✓ cstatus is already installed as statusline")
					return nil
				}

				return fmt.Errorf("existing statusline configuration found; please remove it manually from %s before installing", settingsPath)
			}
		}
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine executable path: %w", err)
	}

	var commandPath string
	if _, err := exec.LookPath("cstatus"); err == nil {
		commandPath = "cstatus"
	} else {
		commandPath = executable
	}

	// Create statusLine configuration
	settings["statusLine"] = map[string]any{
		"type":    "command",
		"command": commandPath,
		"padding": 0,
	}

	settingsJSON, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("could not serialize settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, settingsJSON, 0644); err != nil {
		return fmt.Errorf("could not write settings file: %w", err)
	}

	fmt.Printf("✓ Successfully installed cstatus as Claude Code statusline\n")
	fmt.Printf("✓ Updated %s\n", settingsPath)
	fmt.Printf("\nThe statusline will now appear in Claude Code when you start a new session.\n")

	return nil
}
