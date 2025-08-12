package vt

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/xyproto/env/v2"
)

// TerminalInfo holds information about the current terminal environment
type TerminalInfo struct {
	Type              string
	InTmux            bool
	InScreen          bool
	InSSH             bool
	SupportsMouse     bool
	SupportsColor     bool
	IsPutty           bool
	IsWindowsTerminal bool
	IsConHost         bool
	TmuxSession       string
	TmuxPane          string
}

// GetTerminalInfo detects and returns information about the current terminal
func GetTerminalInfo() *TerminalInfo {
	termType := env.Str("TERM")
	termProgram := env.Str("TERM_PROGRAM")

	info := &TerminalInfo{
		Type: termType,
	}

	// Detect tmux
	tmuxEnv := env.Str("TMUX")
	info.InTmux = tmuxEnv != "" ||
		strings.HasPrefix(termType, "tmux") ||
		strings.HasPrefix(termType, "screen")

	// Extract tmux session info
	if tmuxEnv != "" {
		parts := strings.Split(tmuxEnv, ",")
		if len(parts) >= 2 {
			info.TmuxSession = parts[0]
			info.TmuxPane = parts[1]
		}
	}

	// Detect screen
	info.InScreen = strings.Contains(termType, "screen")

	// Detect SSH
	info.InSSH = env.Str("SSH_TTY") != "" || env.Str("SSH_CLIENT") != ""

	// Detect putty (common putty terminal types)
	info.IsPutty = strings.Contains(termType, "putty") ||
		termProgram == "PuTTY" ||
		env.Str("PUTTY_VER") != ""

	// Detect Windows Terminal
	wtSession := env.Str("WT_SESSION")
	wtProfile := env.Str("WT_PROFILE_ID")
	info.IsWindowsTerminal = wtSession != "" || wtProfile != "" ||
		termProgram == "Windows Terminal"

	// Detect Windows ConHost
	info.IsConHost = termType == "" && !info.IsWindowsTerminal &&
		(env.Str("OS") == "Windows_NT" || env.Str("SESSIONNAME") != "")

	// Color support detection
	info.SupportsColor = termType != "dumb" &&
		(strings.Contains(termType, "color") ||
			strings.Contains(termType, "256") ||
			env.Str("COLORTERM") != "")

	// Mouse support (basic heuristic)
	info.SupportsMouse = !info.InScreen && termType != "dumb"

	return info
}

// SetupResizeHandler sets up SIGWINCH signal handling for terminal resize events
func SetupResizeHandler(canvas *Canvas) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGWINCH)

	go func() {
		for range c {
			handleResize(canvas)
		}
	}()
}

// handleResize handles terminal resize events
func handleResize(canvas *Canvas) {
	if canvas == nil {
		return
	}

	// Get new terminal size
	newW, newH := MustTermSize()

	// Check if size actually changed
	oldW, oldH := canvas.Size()
	if newW == oldW && newH == oldH {
		return
	}

	// Resize the canvas
	if resized := canvas.Resized(); resized != nil {
		// Resized() method handles the internal resize
		_ = resized
	}
}

// IsTerminalCompatible checks if the current terminal supports advanced features
func IsTerminalCompatible() bool {
	info := GetTerminalInfo()

	// Basic compatibility check
	if info.Type == "dumb" {
		return false
	}

	// Very basic terminals
	if strings.Contains(info.Type, "vt52") {
		return false
	}

	return true
}

// GetOptimalTimeout returns optimal timeout for key reading
func GetOptimalTimeout() int {
	info := GetTerminalInfo()

	// Slower timeout for SSH connections
	if info.InSSH {
		return 50 // 50ms
	}

	// Faster timeout for tmux
	if info.InTmux {
		return 5 // 5ms
	}

	// Default timeout
	return 10 // 10ms
}
