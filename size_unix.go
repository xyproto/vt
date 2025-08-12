//go:build !windows

package vt

// getPlatformTermSize gets terminal size using Unix methods
func getPlatformTermSize() (uint, uint, bool) {
	// Unix systems use standard terminal size detection
	// No special platform methods needed beyond what's in size.go
	return 0, 0, false
}
