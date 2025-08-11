package main

import (
	"github.com/xyproto/vt"
)

// Structs and themes that can be used when drawing widgets

type Theme struct {
	Text, Background, Title,
	BoxLight, BoxDark, BoxBackground,
	ButtonFocus, ButtonText,
	ListFocus, ListText, ListBackground vt.AttributeColor
	TL, TR, BL, BR, VL, VR, HT, HB rune
}

func NewTheme() *Theme {
	return &Theme{
		Text:           vt.Black,
		Background:     vt.BackgroundBlue,
		Title:          vt.LightCyan,
		BoxLight:       vt.White,
		BoxDark:        vt.Black,
		BoxBackground:  vt.BackgroundGray,
		ButtonFocus:    vt.LightYellow,
		ButtonText:     vt.White,
		ListFocus:      vt.Red,
		ListText:       vt.Black,
		ListBackground: vt.BackgroundGray,
		TL:             '╭', // top left
		TR:             '╮', // top right
		BL:             '╰', // bottom left
		BR:             '╯', // bottom right
		VL:             '│', // vertical line, left side
		VR:             '│', // vertical line, right side
		HT:             '─', // horizontal line
		HB:             '─', // horizontal bottom line
	}
}

// Output text at the given coordinates, with the configured theme
func (t *Theme) Say(c *vt.Canvas, x, y int, text string) {
	c.Write(uint(x), uint(y), t.Text, t.Background, text)
}

// Set the text color
func (t *Theme) SetTextColor(c vt.AttributeColor) {
	t.Text = c
}

// Set the background color
func (t *Theme) SetBackgroundColor(c vt.AttributeColor) {
	t.Background = c.Background()
}
