package vt

// Viewport is a scrollable sub-region of a Canvas.
// It maps a virtual content area onto a fixed rectangular region of the canvas.
type Viewport struct {
	canvas   *Canvas
	x, y     uint // position on canvas
	w, h     uint // size of viewport
	scrollX  int  // horizontal scroll offset into content
	scrollY  int  // vertical scroll offset into content
	contentW int  // total content width (0 = unbounded)
	contentH int  // total content height (0 = unbounded)
}

// NewViewport creates a viewport positioned at (x, y) with size (w, h) on the given canvas.
func NewViewport(canvas *Canvas, x, y, w, h uint) *Viewport {
	return &Viewport{
		canvas: canvas,
		x:      x,
		y:      y,
		w:      w,
		h:      h,
	}
}

// SetContentSize sets the total content dimensions for scroll clamping.
func (v *Viewport) SetContentSize(w, h int) {
	v.contentW = w
	v.contentH = h
}

// ScrollTo sets the scroll offset, clamping to valid range.
func (v *Viewport) ScrollTo(x, y int) {
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if v.contentW > 0 {
		maxX := max(v.contentW-int(v.w), 0)
		if x > maxX {
			x = maxX
		}
	}
	if v.contentH > 0 {
		maxY := max(v.contentH-int(v.h), 0)
		if y > maxY {
			y = maxY
		}
	}
	v.scrollX = x
	v.scrollY = y
}

// ScrollBy adjusts the scroll offset by a delta.
func (v *Viewport) ScrollBy(dx, dy int) {
	v.ScrollTo(v.scrollX+dx, v.scrollY+dy)
}

// ScrollOffset returns the current scroll position.
func (v *Viewport) ScrollOffset() (int, int) {
	return v.scrollX, v.scrollY
}

// Size returns the viewport dimensions.
func (v *Viewport) Size() (uint, uint) {
	return v.w, v.h
}

// Position returns the viewport's position on the canvas.
func (v *Viewport) Position() (uint, uint) {
	return v.x, v.y
}

// Resize changes the viewport dimensions.
func (v *Viewport) Resize(w, h uint) {
	v.w = w
	v.h = h
}

// Move changes the viewport position on the canvas.
func (v *Viewport) Move(x, y uint) {
	v.x = x
	v.y = y
}

// Clear fills the viewport region with spaces using the given background color.
func (v *Viewport) Clear(bg AttributeColor) {
	for row := uint(0); row < v.h; row++ {
		for col := uint(0); col < v.w; col++ {
			v.canvas.WriteRune(v.x+col, v.y+row, Default, bg, ' ')
		}
	}
}

// WriteAt writes a string at content coordinates (cx, cy), subject to scroll offset.
// Only the visible portion is rendered. Returns the number of columns consumed.
func (v *Viewport) WriteAt(cx, cy int, fg, bg AttributeColor, s string) int {
	// Apply scroll offset to get screen-relative position
	screenY := cy - v.scrollY
	if screenY < 0 || screenY >= int(v.h) {
		return StringWidth(s)
	}

	screenX := cx - v.scrollX
	totalWidth := 0

	for _, r := range s {
		rw := RuneWidth(r)
		if rw <= 0 {
			continue
		}
		if screenX >= 0 && screenX+rw <= int(v.w) {
			canvasX := v.x + uint(screenX)
			canvasY := v.y + uint(screenY)
			if rw == 2 {
				v.canvas.WriteWideRuneB(canvasX, canvasY, fg, bg, r)
			} else {
				v.canvas.WriteRuneB(canvasX, canvasY, fg, bg, r)
			}
		}
		screenX += rw
		totalWidth += rw
	}
	return totalWidth
}

// PlotAt plots a single rune at content coordinates, subject to scroll.
func (v *Viewport) PlotAt(cx, cy int, fg AttributeColor, r rune) {
	screenX := cx - v.scrollX
	screenY := cy - v.scrollY
	if screenX < 0 || screenY < 0 || screenX >= int(v.w) || screenY >= int(v.h) {
		return
	}
	v.canvas.PlotColor(v.x+uint(screenX), v.y+uint(screenY), fg, r)
}

// EnsureVisible adjusts scroll so that content position (cx, cy) is visible.
func (v *Viewport) EnsureVisible(cx, cy int) {
	// Vertical
	if cy < v.scrollY {
		v.scrollY = cy
	} else if cy >= v.scrollY+int(v.h) {
		v.scrollY = cy - int(v.h) + 1
	}
	// Horizontal
	if cx < v.scrollX {
		v.scrollX = cx
	} else if cx >= v.scrollX+int(v.w) {
		v.scrollX = cx - int(v.w) + 1
	}
	// Clamp
	if v.scrollX < 0 {
		v.scrollX = 0
	}
	if v.scrollY < 0 {
		v.scrollY = 0
	}
}

// DrawBorder draws a border around the viewport using the given colors.
// The border is drawn OUTSIDE the viewport (viewport position should account for it).
func (v *Viewport) DrawBorder(fg, bg AttributeColor, style BorderStyle) {
	tl, h, tr, vt, br, bl := style.Glyphs()
	x, y, w, ht := v.x-1, v.y-1, v.w+2, v.h+2

	// Top
	v.canvas.PlotColor(x, y, fg, tl)
	for i := uint(1); i < w-1; i++ {
		v.canvas.PlotColor(x+i, y, fg, h)
	}
	v.canvas.PlotColor(x+w-1, y, fg, tr)

	// Sides
	for i := uint(1); i < ht-1; i++ {
		v.canvas.PlotColor(x, y+i, fg, vt)
		v.canvas.PlotColor(x+w-1, y+i, fg, vt)
	}

	// Bottom
	v.canvas.PlotColor(x, y+ht-1, fg, bl)
	for i := uint(1); i < w-1; i++ {
		v.canvas.PlotColor(x+i, y+ht-1, fg, h)
	}
	v.canvas.PlotColor(x+w-1, y+ht-1, fg, br)

	_ = bg // bg available for background fill if needed
}

// BorderStyle defines a set of border glyphs.
type BorderStyle struct {
	TopLeft, Horizontal, TopRight rune
	Vertical                      rune
	BottomRight, BottomLeft       rune
}

// Glyphs returns the six border characters.
func (bs BorderStyle) Glyphs() (tl, h, tr, v, br, bl rune) {
	return bs.TopLeft, bs.Horizontal, bs.TopRight, bs.Vertical, bs.BottomRight, bs.BottomLeft
}

// Predefined border styles.
var (
	BorderRounded = BorderStyle{'╭', '─', '╮', '│', '╯', '╰'}
	BorderSquare  = BorderStyle{'┌', '─', '┐', '│', '┘', '└'}
	BorderDouble  = BorderStyle{'╔', '═', '╗', '║', '╝', '╚'}
	BorderHeavy   = BorderStyle{'┏', '━', '┓', '┃', '┛', '┗'}
)
