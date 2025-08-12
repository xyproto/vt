//go:build windows

package vt

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/xyproto/env/v2"
)

// getPlatformTermSize gets terminal size using Windows methods
func getPlatformTermSize() (uint, uint, bool) {
	// Try Windows Terminal detection first
	if isWindowsTerminal() {
		if w, h, ok := getWindowsTerminalSize(); ok {
			return w, h, true
		}
	}

	// Try PowerShell/CMD commands
	if w, h, ok := getPowerShellTermSize(); ok {
		return w, h, true
	}

	return 0, 0, false
}

// isWindowsTerminal checks if running in Windows Terminal
func isWindowsTerminal() bool {
	return env.Str("WT_SESSION") != "" || env.Str("WT_PROFILE_ID") != ""
}

// getWindowsTerminalSize gets size using Windows Terminal methods
func getWindowsTerminalSize() (uint, uint, bool) {
	// Windows Terminal supports standard ANSI queries
	// Try environment variables
	if cols := env.Int("COLUMNS", 0); cols > 0 {
		if lines := env.Int("LINES", 0); lines > 0 {
			return uint(cols), uint(lines), true
		}
	}

	// Try wt.exe command
	cmd := exec.Command("wt", "-w", "0", "--", "echo", "%COLUMNS% %LINES%")
	output, err := cmd.Output()
	if err == nil {
		parts := strings.Fields(strings.TrimSpace(string(output)))
		if len(parts) == 2 {
			width, err1 := strconv.Atoi(parts[0])
			height, err2 := strconv.Atoi(parts[1])
			if err1 == nil && err2 == nil && width > 0 && height > 0 {
				return uint(width), uint(height), true
			}
		}
	}

	return 0, 0, false
}

// getPowerShellTermSize gets terminal size using PowerShell
func getPowerShellTermSize() (uint, uint, bool) {
	// Try PowerShell Get-Host
	cmd := exec.Command("powershell", "-Command", "(Get-Host).UI.RawUI.WindowSize.Width; (Get-Host).UI.RawUI.WindowSize.Height")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) == 2 {
			width, err1 := strconv.Atoi(strings.TrimSpace(lines[0]))
			height, err2 := strconv.Atoi(strings.TrimSpace(lines[1]))
			if err1 == nil && err2 == nil && width > 0 && height > 0 {
				return uint(width), uint(height), true
			}
		}
	}

	// Fallback to cmd mode con
	cmd = exec.Command("cmd", "/c", "mode", "con")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		var width, height int
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.Contains(line, "Columns:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					width, _ = strconv.Atoi(parts[1])
				}
			} else if strings.Contains(line, "Lines:") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					height, _ = strconv.Atoi(parts[1])
				}
			}
		}
		if width > 0 && height > 0 {
			return uint(width), uint(height), true
		}
	}

	return 0, 0, false
}
