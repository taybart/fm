package main

// Display files are complicated...sorry its so silly

import (
	"fmt"
	// "github.com/nsf/termbox-go"
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/taybart/log"
	"os"
	"os/exec"
	"strconv"
)

const (
	topOffset        = 1
	fgDefault        = tcell.Color223
	colorFolder      = tcell.Color66
	colorFgActive    = tcell.Color235
	colorHighlight   = tcell.Color109
	colorSymlinkGood = tcell.Color142
	colorSymlinkBad  = tcell.Color167
	colorExec        = tcell.Color124
	colorSelected    = tcell.Color214
)

func drawDir(active int, count int, selected map[string]bool, dir []pseudofile, offset, width int) {
	_, tbheight := scr.Size()
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
		if f.isLink {
			if f.link.broken {
				str += " ~> " + f.link.location
			} else {
				str += " -> " + f.link.location
			}
		}
		if selected[f.fullPath] {
			str = " " + str
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
		if f.isLink && a && f.link.location != "" {
			if cf, err := os.Stat(f.link.location); err == nil && cf.IsDir() {
				c := strconv.Itoa(count)
				str = str[:len(str)-(len(c)+1)] + c + " "
			}
		}
		s := getColors(f, a, selected[f.fullPath])

		// puts(offset, i+topOffset, width, str, true, fg, bg)
		puts(offset, i+topOffset, width, str, true, s)
	}
}

func getColors(f pseudofile, active, selected bool) tcell.Style {
	s := tcell.StyleDefault
	s = s.Foreground(fgDefault)
	if active {
		s = s.Background(colorHighlight)
		s = s.Foreground(colorFgActive)
	}

	if f.isDir {
		s = s.Foreground(colorFolder)
		if active {
			s = s.Foreground(colorFgActive)
		}
		s = s.Bold(true)
	} else {

		if !f.isReal {
			s = s.Foreground(fgDefault)
		} else if (f.f.Mode()&0111) != 0 && !f.isLink {
			s = s.Foreground(colorExec).Bold(true)
		} else if f.isLink && f.link.location != "" {
			if cf, err := os.Stat(f.link.location); err == nil && cf.IsDir() {
				s = s.Foreground(colorSymlinkGood).Bold(true)
			}
			if f.link.broken {
				s = s.Foreground(colorSymlinkBad).Bold(true)
			}
		}
	}
	if selected {
		s = s.Foreground(colorSelected).Bold(true)
	}
	return s
}

func drawParentDir(files []pseudofile, s *fmState, count int) {
	tbwidth, _ := scr.Size()
	cr := conf.ColumnRatios
	cw := conf.ColumnWidth
	if cw < 0 {
		cw = tbwidth
	}
	parentPath := getParentPath(s.cd)
	parentFiles, _, _ := readDir(parentPath)
	if _, ok := s.dt[parentPath]; !ok {
		s.dt[parentPath] = s.dt.newDirForParent(s.cd)
	}
	// Draw parent dir in first column
	width := int(float64(cr[0]) / 10.0 * float64(cw))

	// @TODO temp
	selectedFiles := make(map[string]bool)
	for f := range s.selectedFiles {
		selectedFiles[f] = true
	}

	drawDir(s.dt[parentPath].active, count, selectedFiles, parentFiles, 0, width)

}

func drawChildDir(parent pseudofile, s *fmState, count *int) {
	tbwidth, tbheight := scr.Size()
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
		childPath := s.cd + "/" + parent.name
		if s.cd == "/" {
			childPath = s.cd + parent.name
		}
		files, c, err := readDir(childPath)

		if !os.IsPermission(err) {
			if files[0].isReal {
				*count = c
			}
			if _, ok := s.dt[childPath]; !ok {
				s.dt[childPath] = &dir{active: 0}
			}

			// @TODO temp
			selectedFiles := make(map[string]bool)
			for f := range s.selectedFiles {
				selectedFiles[f] = true
			}
			drawDir(s.dt[childPath].active, 0, selectedFiles, files, offset, width)
		}
	} else if parent.isLink && parent.link.location != "" && !parent.link.broken {
		if f, err := os.Stat(parent.link.location); f.IsDir() && err == nil {
			childP := parent.link.location
			files, c, err := readDir(childP)
			if !os.IsPermission(err) && len(files) > 0 {
				if files[0].isReal {
					*count = c
				}
				if _, ok := s.dt[childP]; !ok {
					s.dt[childP] = &dir{active: 0}
				}
				// @TODO temp
				selectedFiles := make(map[string]bool)
				for f := range s.selectedFiles {
					selectedFiles[f] = true
				}
				drawDir(s.dt[childP].active, 0, selectedFiles, files, offset, width)
			}
		}
	} else if parent.isReal &&
		parent.f.Size() < 3*1024*1024 {

		n := parent.name
		cmd := exec.Command("strings", n)
		buf, _ := cmd.Output()
		if len(buf) > cw*tbheight-2 {
			buf = buf[:cw*tbheight-2]
		}
		puts(offset, topOffset, width,
			string(buf), conf.WrapText, tcell.StyleDefault)
	}
}

func drawHeader(userinput string, files []pseudofile, dt directoryTree, cd string) {
	tbwidth, _ := scr.Size()
	// Print user/cd at top
	un := os.Getenv("USER")
	hn, err := os.Hostname()
	if err != nil {
		log.Errorln(err)
	}
	ustr := un + "@" + hn
	puts(0, 0, tbwidth, ustr, true, tcell.StyleDefault.Foreground(tcell.ColorGreen))
	dn := cd
	oset := 0
	if cd != "/" {
		dn += "/"
		oset = 1
	}

	puts(len(ustr)+1, 0, tbwidth, dn, true, tcell.StyleDefault.Foreground(tcell.ColorGreen))
	f := files[dt[cd].active]
	name := f.name
	if f.isDir {
		name += "/"
	}
	puts(len(ustr)+len(cd)+1+oset, 0, tbwidth, name,
		true, tcell.StyleDefault)
}

func (s *fmState) drawFooter(files []pseudofile) {
	tbwidth, tbheight := scr.Size()
	if len(s.cmd) > 0 {
		puts(0, tbheight-1, tbwidth,
			s.cmd, true, tcell.StyleDefault)
		c := ' '
		if s.cmdIndex < len(s.cmd) {
			c = rune(s.cmd[s.cmdIndex])
		}
		style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
		scr.SetCell(s.cmdIndex, tbheight-1, style, c)
	} else {
		f := files[s.dt[s.cd].active]
		if f.isReal {
			s := fmt.Sprintf("%s %d %s %s",
				f.f.Mode(), f.f.Size(),
				f.f.ModTime().Format("Jan 2 15:04"), f.name)
			puts(0, tbheight-1, tbwidth,
				s, true, tcell.StyleDefault)
		}
	}
}

func draw(s *fmState) {
	scr.Clear()

	tbw, tbh := scr.Size()
	if tbw <= 0 || tbh <= 0 {
		return
	}
	files, amtFiles, err := readDir(".")
	if err != nil {
		log.Errorln(err)
	}

	// draw parent
	drawParentDir(files, s, amtFiles)
	childCount := 0
	drawChildDir(files[s.dt[s.cd].active], s, &childCount)

	{ // Draw current directory
		tbw, _ := scr.Size()
		cr := conf.ColumnRatios
		cw := conf.ColumnWidth
		if cw < 0 {
			cw = tbw
		}
		offset := int(float64(cr[0]) / 10.0 * float64(cw))
		width := int(float64(cr[1]) / 10.0 * float64(cw))
		// @TODO temp
		selectedFiles := make(map[string]bool)
		for f := range s.selectedFiles {
			selectedFiles[f] = true
		}
		drawDir(s.dt[s.cd].active, childCount, selectedFiles, files, offset, width)
	}

	drawHeader(s.cmd, files, s.dt, s.cd)

	// draw footer for frame
	s.drawFooter(files)
	scr.Show()
	// render()
}

func setupDisplay() {
	var err error
	scr, err = tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	encoding.Register()
	if err = scr.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	scr.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorDefault).
		Background(tcell.ColorDefault))
	scr.EnableMouse()
	scr.Clear()
}

func render() {
	// scr.Flush()
	scr.Clear()
}

func puts(x, y, maxWidth int, s string, wrap bool, style tcell.Style) {
	xstart := x
	for _, c := range s {
		if c == '\n' {
			x = xstart
			y++
		} else if c == '\r' {
			x = xstart
		} else {
			scr.SetCell(x, y, style, c)
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
	tbwidth, tbheight := scr.Size()
	puts(tbwidth/4, tbheight/2, tbwidth,
		s, true, tcell.StyleDefault)
	render()
}
