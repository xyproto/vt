package vt

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestNewTTYFromReader_ReadsPrintableKeys(t *testing.T) {
	tty := NewTTYFromReader(strings.NewReader("abc"))
	var got []string
	for i := range 3 {
		k := tty.ReadKey()
		if k == "" {
			t.Fatalf("ReadKey returned empty at i=%d", i)
		}
		got = append(got, k)
	}
	want := []string{"a", "b", "c"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("key %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestNewTTYFromReader_ArrowKey(t *testing.T) {
	// ESC [ A -> Up arrow
	tty := NewTTYFromReader(bytes.NewReader([]byte{27, '[', 'A'}))
	if k := tty.ReadKey(); k != "↑" {
		t.Errorf("expected ↑, got %q", k)
	}
}

func TestNewTTYFromReader_NoOpsDoNotPanic(t *testing.T) {
	tty := NewTTYFromReader(strings.NewReader(""))
	tty.RawMode()
	tty.NoBlock()
	tty.Restore()
	tty.RestoreNoFlush()
	tty.Flush()
	if _, err := tty.SetTimeout(0); err != nil {
		t.Errorf("SetTimeout: %v", err)
	}
	if err := tty.SetTimeoutNoSave(0); err != nil {
		t.Errorf("SetTimeoutNoSave: %v", err)
	}
	ok, err := tty.Poll(0)
	if err != nil || !ok {
		t.Errorf("Poll: got (%v,%v), want (true,nil)", ok, err)
	}
	tty.Close() // must not close fd 0
}

type closingReader struct {
	io.Reader
	closed bool
}

func (c *closingReader) Close() error { c.closed = true; return nil }

func TestNewTTYFromReader_CloseClosesReader(t *testing.T) {
	cr := &closingReader{Reader: strings.NewReader("")}
	tty := NewTTYFromReader(cr)
	tty.Close()
	if !cr.closed {
		t.Error("Close did not close the reader")
	}
}

func TestNewCanvasWithSize(t *testing.T) {
	c := NewCanvasWithSize(10, 3)
	if w, h := c.Size(); w != 10 || h != 3 {
		t.Errorf("size: got (%d,%d), want (10,3)", w, h)
	}
	// Zero dimensions are clamped to 1.
	c2 := NewCanvasWithSize(0, 0)
	if w, h := c2.Size(); w != 1 || h != 1 {
		t.Errorf("zero-size: got (%d,%d), want (1,1)", w, h)
	}
}

func TestCanvasSnapshot_Empty(t *testing.T) {
	c := NewCanvasWithSize(4, 2)
	var buf bytes.Buffer
	if err := c.Snapshot(&buf); err != nil {
		t.Fatal(err)
	}
	want := "vt-snapshot 1 w=4 h=2\n    \n    \n"
	if got := buf.String(); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestCanvasSnapshot_PlottedText(t *testing.T) {
	c := NewCanvasWithSize(5, 2)
	c.PlotColor(0, 0, Default, 'h')
	c.PlotColor(1, 0, Default, 'i')
	var buf bytes.Buffer
	if err := c.Snapshot(&buf); err != nil {
		t.Fatal(err)
	}
	want := "vt-snapshot 1 w=5 h=2\nhi   \n     \n"
	if got := buf.String(); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}

func TestCanvasSnapshot_NonPrintableRendersAsQuestionMark(t *testing.T) {
	c := NewCanvasWithSize(2, 1)
	c.PlotColor(0, 0, Default, '\x01')
	c.PlotColor(1, 0, Default, 'x')
	var buf bytes.Buffer
	if err := c.Snapshot(&buf); err != nil {
		t.Fatal(err)
	}
	want := "vt-snapshot 1 w=2 h=1\n?x\n"
	if got := buf.String(); got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}
