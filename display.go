package main

// Display files are complicated...sorry its so silly

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
	"os/exec"
	"strconv"
)

const (
	topOffset = 1
	fgDefault = termbox.Attribute(0xe0)
)

func drawDir(active int, count int, dir []pseudofile, offset, width int) {
	_, tbheight := termbox.Size()
	viewbox := tbheight - 2
	oob := 0
	// are we off the edge of the display
	if active+tbheight/2 > viewbox {
		oob = (active + tbheight/2) - viewbox
		if len(dir[oob:]) < viewbox {
			oob -= tbheight - 2 - len(dir[oob:])
		}
		if oob < 0 {
			oob = 0
		}
		dir = dir[oob:]
	}
	for i, f := range dir {
		if i+topOffset == tbheight-1 {
			break
		}
		str := f.name
		if f.isDir {
			str += "/"
		}
		if f.isSymL {
			str += " -> " + f.symName
		}

		if len(str) > width-4 {
			str = str[:width-3] + ".."
		}
		for len(str) < width-1 {
			str += " "
		}

		a := (active == i+oob)
		// Append count to end if dir
		if f.isDir && a {
			c := strconv.Itoa(count)
			str = str[:len(str)-(len(c)+1)] + c + " "
		}
		if f.isSymL && a && f.symName != "" {
			if cf, err := os.Stat(f.symName); err == nil && cf.IsDir() {
				c := strconv.Itoa(count)
				str = str[:len(str)-(len(c)+1)] + c + " "
			}
		}
		fg, bg := getColors(f, a)

		printString(offset, i+topOffset, width, str, true, fg, bg)
	}
}

func getColors(f pseudofile, selected bool) (termbox.Attribute, termbox.Attribute) {
	fg := fgDefault
	bg := termbox.ColorDefault
	if selected {
		bg = termbox.ColorBlue
	}

	if f.isDir {
		fg = termbox.ColorCyan
		if selected {
			fg = fgDefault
		}
		fg |= termbox.AttrBold
	} else {

		if !f.isReal {
			fg = fgDefault
		} else if (f.f.Mode()&0111) != 0 && !f.isSymL {
			fg = termbox.ColorYellow | termbox.AttrBold
		} else if f.isSymL && f.symName != "" {
			fg = termbox.ColorMagenta | termbox.AttrBold
			if cf, err := os.Stat(f.symName); err == nil && cf.IsDir() {
				fg = termbox.ColorBlue | termbox.AttrBold
			}
		}
	}
	return fg, bg
}

func drawParentDir(files []pseudofile, dt directoryTree, cd string, count int) {
	tbwidth, _ := termbox.Size()
	cr := conf.ColumnRatios
	cw := conf.ColumnWidth
	if cw < 0 {
		cw = tbwidth
	}
	parentPath := getParentPath(cd)
	parentFiles, _, _ := readDir(parentPath)
	if _, ok := dt[parentPath]; !ok {
		dt[parentPath] = dt.newDirForParent(cd)
	}
	// Draw parent dir in first column
	width := int(float64(cr[0]) / 10.0 * float64(cw))
	drawDir(dt[parentPath].active, count, parentFiles, 0, width)

}

func drawChildDir(parent pseudofile, dt *directoryTree, cd string, count *int) {
	tbwidth, tbheight := termbox.Size()
	cr := conf.ColumnRatios
	cw := conf.ColumnWidth
	if cw < 0 {
		cw = tbwidth
	}
	offset := int(float64(cr[0])/10.0*float64(cw)) +
		int(float64(cr[1])/10.0*float64(cw))
	width := int(float64(cr[2]) / 10.0 * float64(cw))
	// Draw child directory or preview file < 100KB in last column
	if parent.isDir {
		childPath := cd + "/" + parent.name
		if cd == "/" {
			childPath = cd + parent.name
		}
		files, c, err := readDir(childPath)

		if !os.IsPermission(err) {
			if files[0].isReal {
				*count = c
			}
			if _, ok := (*dt)[childPath]; !ok {
				(*dt)[childPath] = &dir{active: 0}
			}
			drawDir((*dt)[childPath].active, 0, files, offset, width)
		}
	} else if parent.isSymL && parent.symName != "" {
		if f, err := os.Stat(parent.symName); f.IsDir() && err == nil {
			childP := parent.symName
			files, c, err := readDir(childP)
			if !os.IsPermission(err) && len(files) > 0 {
				if files[0].isReal {
					*count = c
				}
				if _, ok := (*dt)[childP]; !ok {
					(*dt)[childP] = &dir{active: 0}
				}
				drawDir((*dt)[childP].active, 0, files, offset, width)
			}
		}
	} else if parent.isReal &&
		parent.f.Size() < 100*1024*1024 {

		n := parent.name
		cmd := exec.Command("cat", n)
		buf, _ := cmd.Output()
		if len(buf) > cw*tbheight {
			buf = buf[:cw*tbheight]
		}
		printString(offset, topOffset, width,
			string(buf), conf.WrapText, termbox.ColorDefault, termbox.ColorDefault)
	}
}

func drawHeader(userinput string, files []pseudofile, dt directoryTree, cd string) {
	tbwidth, _ := termbox.Size()
	// Print user/cd at top
	un := os.Getenv("USER")
	hn, err := os.Hostname()
	if err != nil {
		panic(err)
	}
	ustr := un + "@" + hn
	printString(0, 0, tbwidth, ustr, true, termbox.ColorGreen, termbox.ColorDefault)
	dn := cd
	oset := 0
	if cd != "/" {
		dn += "/"
		oset = 1
	}

	printString(len(ustr)+1, 0, tbwidth, dn, true, termbox.ColorBlue, termbox.ColorDefault)
	f := files[dt[cd].active]
	name := f.name
	if f.isDir {
		name += "/"
	}
	printString(len(ustr)+len(cd)+1+oset, 0, tbwidth, name,
		true, termbox.ColorDefault, termbox.ColorDefault)
}

func drawFooter(userinput string, files []pseudofile, dt directoryTree, cd string) {
	tbwidth, tbheight := termbox.Size()
	if len(userinput) > 0 {
		printString(0, tbheight-1, tbwidth,
			userinput, true, termbox.ColorDefault, termbox.ColorDefault)
	} else {
		f := files[dt[cd].active]
		if f.isReal {
			s := fmt.Sprintf("%s %d %s %s",
				f.f.Mode(), f.f.Size(),
				f.f.ModTime().Format("Jan 2 15:04"), f.name)
			printString(0, tbheight-1, tbwidth,
				s, true, termbox.ColorDefault, termbox.ColorDefault)
		}
	}
}

func draw(dt directoryTree, cd, userinput string) {

	files, amtFiles, err := readDir(".")
	if err != nil {
		panic(err) // @TODO: tmp
	}

	// draw parent
	drawParentDir(files, dt, cd, amtFiles)
	childCount := 0
	drawChildDir(files[dt[cd].active], &dt, cd, &childCount)

	{ // Draw current directory
		tbw, _ := termbox.Size()
		cr := conf.ColumnRatios
		cw := conf.ColumnWidth
		if cw < 0 {
			cw = tbw
		}
		offset := int(float64(cr[0]) / 10.0 * float64(cw))
		width := int(float64(cr[1]) / 10.0 * float64(cw))
		drawDir(dt[cd].active, childCount, files, offset, width)
	}

	drawHeader(userinput, files, dt, cd)

	// draw footer for frame
	drawFooter(userinput, files, dt, cd)
	render()
}

func setupDisplay() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	// termbox.SetOutputMode(termbox.OutputNormal)
	termbox.SetOutputMode(termbox.Output256)
}

func render() {
	termbox.Flush()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}

func printString(x, y, maxWidth int, s string, wrap bool, fg, bg termbox.Attribute) {
	xstart := x
	for _, c := range s {
		if c == '\n' {
			x = xstart
			y++
		} else if c == '\r' {
			x = xstart
		} else {
			termbox.SetCell(x, y, c, fg, bg)
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

func printPrompt(s string) {
	tbwidth, tbheight := termbox.Size()
	printString(tbwidth/4, tbheight/2, tbwidth,
		s, true, termbox.ColorDefault, termbox.ColorDefault)
	render()
}
