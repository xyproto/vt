package vt

import (
	"testing"
	"time"
)

// stallingReader yields its data, then reports "nothing right now" forever,
// the way an idle file descriptor does.
type stallingReader struct {
	data  []byte
	delay time.Duration
}

func (r *stallingReader) Read(p []byte) (int, error) {
	if len(r.data) > 0 {
		n := copy(p, r.data)
		r.data = r.data[n:]
		return n, nil
	}
	if r.delay > 0 {
		time.Sleep(r.delay)
	}
	return 0, nil
}

// TestReadKeyNoInputDoesNotPanic checks that reads which return no bytes and
// no error leave ReadKey with an empty pending buffer to report, rather than
// indexing into it.
func TestReadKeyNoInputDoesNotPanic(t *testing.T) {
	tty := NewTTYFromReader(&stallingReader{data: []byte("ab")})
	defer tty.Close()

	for i, want := range []string{"a", "b"} {
		if got := tty.ReadKey(); got != want {
			t.Fatalf("key %d: got %q, want %q", i, got, want)
		}
	}
	if got := tty.ReadKey(); got != "" {
		t.Errorf("with no input available: got %q, want %q", got, "")
	}
}

// TestReadKeyTimeoutReturnsEmpty checks that ReadKeyTimeout gives up after the
// given duration instead of waiting for input that never arrives.
func TestReadKeyTimeoutReturnsEmpty(t *testing.T) {
	tty := NewTTYFromReader(&stallingReader{delay: 20 * time.Millisecond})
	defer tty.Close()

	start := time.Now()
	if got := tty.ReadKeyTimeout(50 * time.Millisecond); got != "" {
		t.Errorf("got %q, want %q", got, "")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("ReadKeyTimeout blocked for %v", elapsed)
	}
}

// TestReadKeyTimeoutReturnsPendingKeys checks that buffered keys are returned
// immediately, and that a burst arriving in one read is handed out key by key.
func TestReadKeyTimeoutReturnsPendingKeys(t *testing.T) {
	tty := NewTTYFromReader(&stallingReader{data: []byte("xyz")})
	defer tty.Close()

	for i, want := range []string{"x", "y", "z"} {
		if got := tty.ReadKeyTimeout(time.Second); got != want {
			t.Fatalf("key %d: got %q, want %q", i, got, want)
		}
	}
}

// TestReadKeyLongBurst checks that a burst larger than one read buffer is
// returned in full, in order.
func TestReadKeyLongBurst(t *testing.T) {
	const count = 20000
	payload := make([]byte, count)
	for i := range payload {
		payload[i] = byte('a' + i%26)
	}
	tty := NewTTYFromReader(&stallingReader{data: payload})
	defer tty.Close()

	for i := range count {
		want := string(rune('a' + i%26))
		if got := tty.ReadKey(); got != want {
			t.Fatalf("key %d of %d: got %q, want %q", i, count, got, want)
		}
	}
	if got := tty.ReadKey(); got != "" {
		t.Errorf("after the burst: got %q, want %q", got, "")
	}
}
