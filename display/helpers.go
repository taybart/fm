package display

import (
	"github.com/gdamore/tcell/v2"
)

func puts(x, y, maxWidth int, s string, wrap bool, style tcell.Style) {
	xstart := x
	for _, c := range s {
		if c == '\n' {
			x = xstart
			y++
		} else if c == '\r' {
			x = xstart
		} else {
			scr.SetContent(x, y, c, nil, style)
			x++
			if x > xstart+maxWidth {
				if !wrap {
					break
				}
				x = xstart
				y++
			}
		}
	}
}

/* func printPrompt(s string) {
	tbwidth, tbheight := scr.Size()
	puts(tbwidth/4, tbheight/2, tbwidth,
		s, true, tcell.StyleDefault)
	scr.Show()
} */
