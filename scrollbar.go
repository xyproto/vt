package vt

// Scrollbar renders a vertical or horizontal scrollbar on a Canvas.
type Scrollbar struct {
	canvas *Canvas
	fg     AttributeColor
	bg     AttributeColor
}

// NewScrollbar creates a scrollbar renderer for the given canvas.
func NewScrollbar(canvas *Canvas, fg, bg AttributeColor) *Scrollbar {
	return &Scrollbar{canvas: canvas, fg: fg, bg: bg}
}

// DrawVertical draws a vertical scrollbar at column x, spanning rows [y, y+height).
// position is 0.0–1.0 indicating the thumb position.
// thumbSize is 0.0–1.0 indicating the fraction of content visible.
func (sb *Scrollbar) DrawVertical(x, y, height uint, position, thumbSize float64) {
	if height == 0 {
		return
	}
	if thumbSize > 1 {
		thumbSize = 1
	}
	if thumbSize < 0 {
		thumbSize = 0
	}
	if position < 0 {
		position = 0
	}
	if position > 1 {
		position = 1
	}

	thumbH := max(int(float64(height)*thumbSize), 1)

	trackSpace := int(height) - thumbH
	thumbStart := int(float64(trackSpace) * position)

	for i := range height {
		r := '│'
		color := sb.bg
		if int(i) >= thumbStart && int(i) < thumbStart+thumbH {
			r = '┃'
			color = sb.fg
		}
		sb.canvas.PlotColor(x, y+i, color, r)
	}
}

// DrawHorizontal draws a horizontal scrollbar at row y, spanning columns [x, x+width).
func (sb *Scrollbar) DrawHorizontal(x, y, width uint, position, thumbSize float64) {
	if width == 0 {
		return
	}
	if thumbSize > 1 {
		thumbSize = 1
	}
	if thumbSize < 0 {
		thumbSize = 0
	}
	if position < 0 {
		position = 0
	}
	if position > 1 {
		position = 1
	}

	thumbW := max(int(float64(width)*thumbSize), 1)

	trackSpace := int(width) - thumbW
	thumbStart := int(float64(trackSpace) * position)

	for i := range width {
		r := '─'
		color := sb.bg
		if int(i) >= thumbStart && int(i) < thumbStart+thumbW {
			r = '━'
			color = sb.fg
		}
		sb.canvas.PlotColor(x+i, y, color, r)
	}
}
