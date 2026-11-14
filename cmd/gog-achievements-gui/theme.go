package main

import (
	"image/color"

	"gioui.org/text"
	"gioui.org/widget/material"
)

// Theme colours mapped from tokens.css (GOG late-night neon).
// OKLCH sources live in /tokens.css — these NRGBA values are the Gio runtime map.
var (
	colPaper     = color.NRGBA{R: 0x1A, G: 0x14, B: 0x24, A: 0xFF} // oklch(14% 0.018 305)
	colPaper2    = color.NRGBA{R: 0x24, G: 0x1C, B: 0x32, A: 0xFF} // oklch(18% 0.020 305)
	colPaper3    = color.NRGBA{R: 0x2E, G: 0x24, B: 0x40, A: 0xFF} // oklch(22% 0.022 305)
	colInk       = color.NRGBA{R: 0xF2, G: 0xEE, B: 0xF8, A: 0xFF} // oklch(94% 0.010 305)
	colInk2      = color.NRGBA{R: 0xB0, G: 0xA6, B: 0xC4, A: 0xFF} // oklch(72% 0.012 305)
	colMuted     = color.NRGBA{R: 0x7A, G: 0x70, B: 0x90, A: 0xFF} // oklch(52% 0.014 305)
	colRule      = color.NRGBA{R: 0x3A, G: 0x30, B: 0x4E, A: 0xFF} // oklch(28% 0.020 305)
	colAccent    = color.NRGBA{R: 0xB8, G: 0x4D, B: 0xFF, A: 0xFF} // oklch(64% 0.20 305)
	colAccentInk = color.NRGBA{R: 0x1A, G: 0x14, B: 0x24, A: 0xFF}
	colFocus     = color.NRGBA{R: 0xC9, G: 0x6A, B: 0xFF, A: 0xFF} // oklch(68% 0.22 305)
	colSuccess   = color.NRGBA{R: 0x3D, G: 0xD6, B: 0x8C, A: 0xFF} // oklch(72% 0.14 155)
	colDanger    = color.NRGBA{R: 0xF0, G: 0x6A, B: 0x5A, A: 0xFF} // oklch(68% 0.16 25)
)

func newTheme() *material.Theme {
	th := material.NewTheme()
	th.Palette.Bg = colPaper
	th.Palette.Fg = colInk
	th.Palette.ContrastBg = colAccent
	th.Palette.ContrastFg = colAccentInk
	th.Shaper = text.NewShaper()
	return th
}
