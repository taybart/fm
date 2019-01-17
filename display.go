package main

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

func debug(x, y int, format string, v ...interface{}) {
	fg := termbox.ColorDefault
	bg := termbox.ColorDefault
	s := fmt.Sprintf(format, v...)
	printString(x, y, 10000, s, fg, bg)
}

func drawDir(active int, count int, dir []pseudofile, offset, width int) {
	_, tbheight := termbox.Size()
	viewbox := tbheight - 2
	oob := 0
	// are we off the edge
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
		if f.isSymL && a {
			if f, err := os.Stat(f.symName); f.IsDir() && err == nil {
				c := strconv.Itoa(count)
				str = str[:len(str)-(len(c)+1)] + c + " "
			}
		}
		fg, bg := getColors(f, a)

		printString(offset, i+topOffset, width, str, fg, bg)
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
			if f, err := os.Stat(f.symName); f.IsDir() && err == nil {
				fg = termbox.ColorBlue | termbox.AttrBold
			}
		}
	}
	return fg, bg
}

func printStringNoWrap(x, y, maxWidth int, s string, fg, bg termbox.Attribute) {
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
				break
			}
		}
	}
}
func printString(x, y, maxWidth int, s string, fg, bg termbox.Attribute) {
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
				x = xstart
				y++
			}
		}
	}
}

func printPrompt(s string) {
	tbwidth, tbheight := termbox.Size()
	printString(tbwidth/4, tbheight/2, tbwidth,
		s, termbox.ColorDefault, termbox.ColorDefault)
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

func draw(dt directoryTree, cd, userinput string) {
	cr := conf.ColumnRatios
	cw := conf.ColumnWidth

	tbwidth, tbheight := termbox.Size()
	if cw < 0 {
		cw = tbwidth
	}

	files, amtFiles, err := readDir(".")
	if err != nil {
		panic(err) // @TODO: tmp
	}

	parentPath := getParentPath(cd)
	parentFiles, _, err := readDir(parentPath)
	if _, ok := dt[parentPath]; !ok {
		dt[parentPath] = dt.newDirForParent(cd)
	}

	// Draw parent dir in first column
	offset := 0
	width := int(float64(cr[0]) / 10.0 * float64(cw))
	drawDir(dt[parentPath].active, amtFiles, parentFiles, offset, width)

	offset = int(float64(cr[0])/10.0*float64(cw)) +
		int(float64(cr[1])/10.0*float64(cw))
	width = int(float64(cr[2]) / 10.0 * float64(cw))
	count := 0
	if len(files) > 0 {
		// Draw child directory or preview file < 100KB in last column
		if files[dt[cd].active].isDir {
			childPath := cd + "/" + files[dt[cd].active].name
			if cd == "/" {
				childPath = cd + files[dt[cd].active].name
			}
			files, c, err := readDir(childPath)

			if !os.IsPermission(err) {
				if files[0].isReal {
					count = c
				}
				if _, ok := dt[childPath]; !ok {
					dt[childPath] = &dir{active: 0}
				}
				drawDir(dt[childPath].active, 0, files, offset, width)
			}
		} else if files[dt[cd].active].isSymL && files[dt[cd].active].symName != "" {
			if f, err := os.Stat(files[dt[cd].active].symName); f.IsDir() && err == nil {
				childP := files[dt[cd].active].symName
				files, c, err := readDir(childP)
				if !os.IsPermission(err) && len(files) > 0 {
					if files[0].isReal {
						count = c
					}
					if _, ok := dt[childP]; !ok {
						dt[childP] = &dir{active: 0}
					}
					drawDir(dt[childP].active, 0, files, offset, width)
				}
			}
		} else if files[dt[cd].active].isReal &&
			files[dt[cd].active].f.Size() < 100*1024*1024 {

			n := files[dt[cd].active].name
			// cmd := exec.Command("/Users/taylor/Downloads/vimpager/vimcat", "-o", "-", n)
			cmd := exec.Command("cat", n)
			buf, _ := cmd.Output()
			if len(buf) > cw*tbheight {
				buf = buf[:200]
			}
			if conf.WrapText {
				printString(offset, topOffset, width,
					string(buf), termbox.ColorDefault, termbox.ColorDefault)
			} else {
				printStringNoWrap(offset, topOffset, width,
					string(buf), termbox.ColorDefault, termbox.ColorDefault)
			}
		}
	}

	// Draw current dir in middle column
	offset = int(float64(cr[0]) / 10.0 * float64(cw))
	width = int(float64(cr[1]) / 10.0 * float64(cw))
	drawDir(dt[cd].active, count, files, offset, width)

	{
		// Print user/cd at top
		// cdFG := termbox.ColorBlue | termbox.AttrBold
		un := os.Getenv("USER")
		hn, err := os.Hostname()
		if err != nil {
			panic(err)
		}
		ustr := un + "@" + hn
		printString(0, 0, tbwidth, ustr, termbox.ColorGreen, termbox.ColorDefault)
		// printString(len(ustr)+1, 0, tbwidth, cd, cdFG, termbox.ColorDefault)
		dn := cd
		oset := 0
		if cd != "/" {
			dn += "/"
			oset = 1
		}

		printString(len(ustr)+1, 0, tbwidth, dn, termbox.ColorBlue, termbox.ColorDefault)
		f := files[dt[cd].active]
		name := f.name
		if f.isDir {
			name += "/"
		}
		printString(len(ustr)+len(cd)+1+oset, 0, tbwidth, name,
			termbox.ColorDefault, termbox.ColorDefault)
	}

	// Print user input or dir info at bottom
	if len(userinput) > 0 {
		printString(0, tbheight-1, tbwidth,
			userinput, termbox.ColorDefault, termbox.ColorDefault)
	} else {
		f := files[dt[cd].active]
		if f.isReal {
			s := fmt.Sprintf("%s %d %s %s",
				f.f.Mode(), f.f.Size(),
				f.f.ModTime().Format("Jan 2 15:04"), f.name)
			printString(0, tbheight-1, tbwidth,
				s, termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	render()
}
