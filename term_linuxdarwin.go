//go:build linux || darwin

package vt

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

var (
	origTermios *unix.Termios
	termMutex   sync.Mutex
)

func (e *Editor) setupSignalHandler() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for sig := range ch {
			switch sig {
			case syscall.SIGWINCH:
				e.handleSigWinCh()
			case syscall.SIGTERM, syscall.SIGINT:
				e.DisableRawMode()
				os.Exit(0)
			}
		}
	}()
}

func (e *Editor) enableRawMode() error {
	if e.rawmode {
		return nil
	}
	fd := int(os.Stdin.Fd())
	if !isatty(fd) {
		return fmt.Errorf("not a tty")
	}

	termios, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return err
	}

	termMutex.Lock()
	if origTermios == nil {
		// Save original copy
		orig := *termios
		origTermios = &orig
		e.origTermios = &orig
	}
	termMutex.Unlock()

	raw := *termios
	// Input modes
	raw.Iflag &^= (unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON)
	raw.Iflag |= unix.IUTF8 // Ensure UTF-8 input

	// Output modes
	raw.Oflag &^= unix.OPOST

	// Control modes
	raw.Cflag &^= (unix.CSIZE | unix.PARENB)
	raw.Cflag |= unix.CS8

	// Local modes
	raw.Lflag &^= (unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN)

	// Blocking read
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, &raw); err != nil {
		return err
	}
	e.rawmode = true

	// Enable mouse tracking (SGR 1006) and bracketed paste
	fmt.Print("\x1b[?1006;1000h\x1b[?2004h")

	// Start input reader
	e.startInputReader()

	return nil
}

// DisableRawMode restores the terminal to its original mode.
func (e *Editor) DisableRawMode() {
	termMutex.Lock()
	defer termMutex.Unlock()

	// Disable mouse tracking and bracketed paste
	fmt.Print("\x1b[?1006;1000l\x1b[?2004l")

	if e.rawmode && origTermios != nil {
		unix.IoctlSetTermios(int(os.Stdin.Fd()), ioctlWriteTermios, origTermios)
		e.rawmode = false
	}
}

func (e *Editor) startInputReader() {
	if e.inputChan != nil {
		return
	}
	e.inputChan = make(chan KeyEvent, 128)

	go func() {
		buf := make([]byte, 1024)
		reader := bufio.NewReader(os.Stdin)

		for {
			n, err := reader.Read(buf)
			if err != nil {
				close(e.inputChan)
				return
			}
			data := buf[:n]

			// Process buffer
			for i := 0; i < len(data); {
				// Handle ESC sequences
				if data[i] == '\x1b' {
					// Simple lookahead for ESC sequence vs separate ESC
					// If we only have ESC and nothing else in this read, we need to be careful.
					// But we are in blocking mode VMIN=1.
					// If user typed ESC, we get 1 byte.
					// If user typed Arrow, terminal sends 3 bytes quickly.
					// They usually come in one Read, or very close.
					// Ideally we use a timeout here if we only have ESC.

					if i+1 >= len(data) {
						// Only ESC left in buffer. Wait briefly to see if more comes.
						// This is a naive implementation of timeout.
						// A robust one would use select with timeout on reading.
						// But we already did a blocking Read.
						// We can check Buffered() but that only tells if we have buffer.
						// We can try to Read with deadline? os.Stdin doesn't support it easily.
						// Let's assume standalone ESC if not immediately followed.
						e.inputChan <- KeyEvent{Key: keyEsc}
						i++
						continue
					}

					// Parse sequence
					consumed, event := parseSequence(data[i:])
					if consumed > 0 {
						e.inputChan <- event
						i += consumed
						continue
					}
				}

				// Regular character
				// Handle UTF-8?
				// For now simple ascii/byte mapping
				k := int(data[i])
				if k == 127 {
					e.inputChan <- KeyEvent{Key: keyBackspace}
				} else if k < 32 {
					// Ctrl chars
					switch k {
					case 13:
						e.inputChan <- KeyEvent{Key: keyEnter}
					case 9:
						e.inputChan <- KeyEvent{Key: keyTab}
					default:
						// Map to ctrl constant
						e.inputChan <- KeyEvent{Key: k}
					}
				} else {
					e.inputChan <- KeyEvent{Key: k, Rune: rune(k)}
				}
				i++
			}
		}
	}()
}

func parseSequence(seq []byte) (int, KeyEvent) {
	if len(seq) < 2 || seq[0] != '\x1b' {
		return 0, KeyEvent{}
	}

	if seq[1] == '[' || seq[1] == 'O' {
		if len(seq) < 3 {
			return 0, KeyEvent{}
		}

		// Handle simple CSI/SS3
		switch seq[2] {
		case 'A':
			return 3, KeyEvent{Key: arrowUp}
		case 'B':
			return 3, KeyEvent{Key: arrowDown}
		case 'C':
			return 3, KeyEvent{Key: arrowRight}
		case 'D':
			return 3, KeyEvent{Key: arrowLeft}
		case 'H':
			return 3, KeyEvent{Key: homeKey}
		case 'F':
			return 3, KeyEvent{Key: endKey}
		}

		// PageUp/Down, Home/End, Insert/Delete: [1~, [2~ etc
		if seq[1] == '[' && len(seq) >= 4 && seq[3] == '~' {
			switch seq[2] {
			case '1':
				return 4, KeyEvent{Key: homeKey}
			case '2':
				return 4, KeyEvent{Key: 0} // Insert
			case '3':
				return 4, KeyEvent{Key: delKey}
			case '4':
				return 4, KeyEvent{Key: endKey}
			case '5':
				return 4, KeyEvent{Key: pageUp}
			case '6':
				return 4, KeyEvent{Key: pageDown}
			}
		}

		// TERM=linux F-keys: [[A .. [[E for F1..F5
		if seq[1] == '[' && seq[2] == '[' {
			if len(seq) >= 4 {
				switch seq[3] {
				case 'A':
					return 4, KeyEvent{Key: keyF1}
				case 'B':
					return 4, KeyEvent{Key: keyF2}
				case 'C':
					return 4, KeyEvent{Key: keyF3}
				case 'D':
					return 4, KeyEvent{Key: keyF4}
				case 'E':
					return 4, KeyEvent{Key: keyF5}
				}
			}
		}

		// Standard F-keys: [11~ .. [24~
		if seq[1] == '[' && len(seq) >= 5 && seq[4] == '~' {
			// [11~ is F1, [12~ is F2 ...
			// Simple mapping
			num := string(seq[2:4])
			switch num {
			case "11":
				return 5, KeyEvent{Key: keyF1}
			case "12":
				return 5, KeyEvent{Key: keyF2}
			case "13":
				return 5, KeyEvent{Key: keyF3}
			case "14":
				return 5, KeyEvent{Key: keyF4}
			case "15":
				return 5, KeyEvent{Key: keyF5}
			case "17":
				return 5, KeyEvent{Key: keyF6}
			case "18":
				return 5, KeyEvent{Key: keyF7}
			case "19":
				return 5, KeyEvent{Key: keyF8}
			case "20":
				return 5, KeyEvent{Key: keyF9}
			case "21":
				return 5, KeyEvent{Key: keyF10}
			case "23":
				return 5, KeyEvent{Key: keyF11}
			case "24":
				return 5, KeyEvent{Key: keyF12}
			}
		}
	}

	// Fallback: Just return ESC if we can't parse, or skip?
	// For robust parser, we might consume just ESC.
	return 0, KeyEvent{}
}

func (e *Editor) readKey() int {
	if e.inputChan == nil {
		e.startInputReader()
	}
	event := <-e.inputChan
	return event.Key
}

func getWindowSize() (int, int, error) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 24, 80, nil
	}
	return int(ws.Row), int(ws.Col), nil
}

func (e *Editor) updateWindowSize() error {
	rows, cols, err := getWindowSize()
	if err != nil {
		return err
	}
	e.screenrows = rows - 2 // room for status bar
	e.screencols = cols
	return nil
}

func (e *Editor) handleSigWinCh() {
	e.updateWindowSize()
	if e.cy > e.screenrows {
		e.cy = e.screenrows - 1
	}
	if e.cx > e.screencols {
		e.cx = e.screencols - 1
	}
	e.refreshScreen()
}

// ---------- termios helpers (POSIX) ----------

func isatty(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	return err == nil
}
