//go:build windows

package vt

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleMode             = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode             = kernel32.NewProc("SetConsoleMode")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
	procWaitForSingleObject        = kernel32.NewProc("WaitForSingleObject")
)

const (
	enableProcessedInput       = 0x0001
	enableLineInput            = 0x0002
	enableEchoInput            = 0x0004
	enableVirtualTerminalInput = 0x0200

	enableProcessedOutput           = 0x0001
	enableWrapAtEolOutput           = 0x0002
	enableVirtualTerminalProcessing = 0x0004

	waitObject0 = 0
)

type consoleScreenBufferInfo struct {
	Size        coord
	Cursor      coord
	Attributes  uint16
	Window      smallRect
	MaximumSize coord
}

type coord struct {
	X int16
	Y int16
}

type smallRect struct {
	Left   int16
	Top    int16
	Right  int16
	Bottom int16
}

type windowsTermios struct {
	inMode  uint32
	outMode uint32
}

func (e *Editor) setupSignalHandler() {
	// Windows doesn't support SIGWINCH
}

func (e *Editor) enableRawMode() error {
	if e.rawmode {
		return nil
	}

	var inMode uint32
	var outMode uint32

	hIn := syscall.Stdin
	hOut := syscall.Stdout

	if err := getConsoleMode(hIn, &inMode); err != nil {
		return err
	}
	if err := getConsoleMode(hOut, &outMode); err != nil {
		return err
	}

	e.origTermios = windowsTermios{inMode: inMode, outMode: outMode}

	// Input modes
	newInMode := inMode
	newInMode &^= (enableEchoInput | enableLineInput | enableProcessedInput)
	newInMode |= enableVirtualTerminalInput

	// Output modes
	newOutMode := outMode
	newOutMode |= enableVirtualTerminalProcessing

	if err := setConsoleMode(hIn, newInMode); err != nil {
		return err
	}
	if err := setConsoleMode(hOut, newOutMode); err != nil {
		return err
	}

	e.rawmode = true
	return nil
}

// DisableRawMode restores the terminal to its original mode.
func (e *Editor) DisableRawMode() {
	if e.rawmode {
		orig := e.origTermios.(windowsTermios)
		setConsoleMode(syscall.Stdin, orig.inMode)
		setConsoleMode(syscall.Stdout, orig.outMode)
		e.rawmode = false
	}
}

func (e *Editor) readKey() int {
	var buf [1]byte
	for {
		// Use os.Stdin.Read which wraps syscall.Read
		n, err := os.Stdin.Read(buf[:])
		if n == 1 {
			break
		}
		if err != nil {
			return -1
		}
	}
	c := int(buf[0])
	if c == keyEsc {
		var seq [3]byte

		if !waitForInput(100) {
			return keyEsc
		}
		n, _ := os.Stdin.Read(seq[0:1])
		if n == 0 {
			return keyEsc
		}

		if !waitForInput(100) {
			return keyEsc
		}
		n, _ = os.Stdin.Read(seq[1:2])
		if n == 0 {
			return keyEsc
		}

		if seq[0] == '[' {
			if seq[1] >= '0' && seq[1] <= '9' {
				if !waitForInput(100) {
					return keyEsc
				}
				n, _ = os.Stdin.Read(seq[2:3])
				if n == 0 {
					return keyEsc
				}
				if seq[2] == '~' {
					switch seq[1] {
					case '3':
						return delKey
					case '5':
						return pageUp
					case '6':
						return pageDown
					}
				}
			} else {
				switch seq[1] {
				case 'A':
					return arrowUp
				case 'B':
					return arrowDown
				case 'C':
					return arrowRight
				case 'D':
					return arrowLeft
				case 'H':
					return homeKey
				case 'F':
					return endKey
				}
			}
		} else if seq[0] == 'O' {
			switch seq[1] {
			case 'H':
				return homeKey
			case 'F':
				return endKey
			}
		}
		return keyEsc
	}
	return c
}

func getWindowSize() (int, int, error) {
	var info consoleScreenBufferInfo
	if err := getConsoleScreenBufferInfo(syscall.Stdout, &info); err != nil {
		return 24, 80, nil
	}
	// Window.Right/Bottom are inclusive 0-based coordinates.
	// Width = Right - Left + 1
	// Height = Bottom - Top + 1
	cols := int(info.Window.Right - info.Window.Left + 1)
	rows := int(info.Window.Bottom - info.Window.Top + 1)
	return rows, cols, nil
}

func (e *Editor) updateWindowSize() error {
	rows, cols, err := getWindowSize()
	if err != nil {
		return err
	}
	e.screenrows = rows - 2
	e.screencols = cols
	return nil
}

func (e *Editor) handleSigWinCh() {
	// Not called on Windows
}

// Helpers

func getConsoleMode(handle syscall.Handle, mode *uint32) error {
	r1, _, err := procGetConsoleMode.Call(uintptr(handle), uintptr(unsafe.Pointer(mode)))
	if r1 == 0 {
		return err
	}
	return nil
}

func setConsoleMode(handle syscall.Handle, mode uint32) error {
	r1, _, err := procSetConsoleMode.Call(uintptr(handle), uintptr(mode))
	if r1 == 0 {
		return err
	}
	return nil
}

func getConsoleScreenBufferInfo(handle syscall.Handle, info *consoleScreenBufferInfo) error {
	r1, _, err := procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(info)))
	if r1 == 0 {
		return err
	}
	return nil
}

func waitForInput(timeoutMs uint32) bool {
	r1, _, _ := procWaitForSingleObject.Call(uintptr(syscall.Stdin), uintptr(timeoutMs))
	return r1 == uintptr(waitObject0)
}
