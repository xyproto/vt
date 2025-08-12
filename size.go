package vt

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/xyproto/env/v2"
	"golang.org/x/term"
)

// MustTermSize returns the terminal size
func MustTermSize() (uint, uint) {
	// Try platform-specific detection first
	if w, h, ok := getPlatformTermSize(); ok {
		return w, h
	}

	// Try tmux detection first
	if w, h, ok := getTmuxSize(); ok {
		return w, h
	}

	// Try standard detection
	fd := int(os.Stdout.Fd())
	if term.IsTerminal(fd) {
		width, height, err := term.GetSize(fd)
		if err == nil {
			return uint(width), uint(height)
		}
	}

	// Try alternative file descriptors
	if w, h, ok := getAlternativeFdSize(); ok {
		return w, h
	}

	// Fallback to environment variables
	var w uint = 79
	if cols := env.Int("COLS", 0); cols > 0 {
		w = uint(cols)
	} else if cols := env.Int("COLUMNS", 0); cols > 0 {
		w = uint(cols)
	}
	return w, uint(env.Int("LINES", 25))
}

// getTmuxSize gets terminal size from tmux
func getTmuxSize() (uint, uint, bool) {
	// Check for tmux session
	if env.Str("TMUX") == "" {
		return 0, 0, false
	}

	// Try multiple tmux size detection methods
	// Method 1: Current pane size
	cmd := exec.Command("tmux", "display-message", "-p", "#{pane_width} #{pane_height}")
	output, err := cmd.Output()
	if err == nil {
		if w, h, ok := parseTmuxSize(output); ok {
			return w, h, true
		}
	}

	// Method 2: Window size (fallback)
	cmd = exec.Command("tmux", "display-message", "-p", "#{window_width} #{window_height}")
	output, err = cmd.Output()
	if err != nil {
		return 0, 0, false
	}

	return parseTmuxSize(output)
}

// parseTmuxSize parses tmux size output
func parseTmuxSize(output []byte) (uint, uint, bool) {
	parts := strings.Fields(strings.TrimSpace(string(output)))
	if len(parts) != 2 {
		return 0, 0, false
	}

	width, err1 := strconv.Atoi(parts[0])
	height, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}

	// Validate reasonable terminal sizes
	if width < 1 || height < 1 || width > 1000 || height > 1000 {
		return 0, 0, false
	}

	return uint(width), uint(height), true
}

// getAlternativeFdSize tries different file descriptors
func getAlternativeFdSize() (uint, uint, bool) {
	// Try stderr if stdout doesn't work
	if term.IsTerminal(int(os.Stderr.Fd())) {
		width, height, err := term.GetSize(int(os.Stderr.Fd()))
		if err == nil {
			return uint(width), uint(height), true
		}
	}

	// Try stdin
	if term.IsTerminal(int(os.Stdin.Fd())) {
		width, height, err := term.GetSize(int(os.Stdin.Fd()))
		if err == nil {
			return uint(width), uint(height), true
		}
	}

	return 0, 0, false
}
