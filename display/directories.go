package display

import (
	"os"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/taybart/fm/fs"
	"github.com/taybart/log"
)

// DrawDir render directory
func drawDir(dir fs.Directory, offset, width int) {
	tbw, tbh := scr.Size()
	if tbw <= 0 || tbh <= 0 {
		return
	}
	_, tbheight := scr.Size()
	viewbox := tbheight - 2
	oob := 0
	// are we off the edge of the display
	if dir.Active+tbheight/2 > viewbox {
		oob = (dir.Active + tbheight/2) - viewbox
		if len(dir.Files[oob:]) < viewbox {
			oob -= tbheight - 2 - len(dir.Files[oob:])
		}
		if oob < 0 {
			oob = 0
		}
		dir.Files = dir.Files[oob:]
	}
	for i, f := range dir.Files {
		if i+topOffset == tbheight-1 {
			break
		}
		str := f.Name
		if f.IsDir {
			str += "/"
		}
		if f.IsLink {
			if f.Link.Broken {
				str += " ~> " + f.Link.Location
			} else {
				str += " -> " + f.Link.Location
			}
		}
		if dir.Selected[f.FullPath] {
			str = " " + str
		}

		if len(str) > width-4 {
			str = str[:width-3] + ".."
		}
		for len(str) < width-1 {
			str += " "
		}

		a := (dir.Active == i+oob)
		// Append count to end if dir
		if f.IsDir && a {
			count := fs.CountChildren(f)
			c := strconv.Itoa(count)
			str = str[:len(str)-(len(c)+1)] + c + " "
		}
		if f.IsLink && a && f.Link.Location != "" {
			if cf, err := os.Stat(f.Link.Location); err == nil && cf.IsDir() {
				count := fs.CountChildren(f)
				c := strconv.Itoa(count)
				str = str[:len(str)-(len(c)+1)] + c + " "
			}
		}
		s := getColors(f, a, dir.Selected[f.FullPath])

		puts(offset, i+topOffset, width, str, true, s)
	}
}

func drawParentDir(dir fs.Directory) {
	tbwidth, _ := scr.Size()
	cr := conf.ColumnRatios
	cw := conf.ColumnWidth
	if cw < 0 {
		cw = tbwidth
	}
	// Draw parent dir in first column
	width := int(float64(cr[0]) / 10.0 * float64(cw))

	drawDir(dir, 0, width)
}

func drawChildDir(w Window, offset, width int) {
	if w.Child.Path != "" {
		drawDir(w.Child, offset, width)
	} else {
		err := drawFilePreview(offset, width, w.Current.ActiveFile.FullPath)
		if err != nil && err != errFileTooLarge {
			log.Error(err)
		}
	}
}
func getColors(f fs.Pseudofile, active, selected bool) tcell.Style {
	s := tcell.StyleDefault
	s = s.Foreground(fgDefault)
	if active {
		s = s.Background(colorHighlight)
		s = s.Foreground(colorFgActive)
	}

	if f.IsDir {
		s = s.Foreground(colorFolder)
		if active {
			s = s.Foreground(colorFgActive)
		}
		s = s.Bold(true)
	} else {

		if !f.IsReal {
			s = s.Foreground(fgDefault)
		} else if (f.F.Mode()&0111) != 0 && !f.IsLink {
			s = s.Foreground(colorExec).Bold(true)
		} else if f.IsLink && f.Link.Location != "" {
			if cf, err := os.Stat(f.Link.Location); err == nil && cf.IsDir() {
				s = s.Foreground(colorSymlinkGood).Bold(true)
			}
			if f.Link.Broken {
				s = s.Foreground(colorSymlinkBad).Bold(true)
			}
		}
	}
	if selected {
		s = s.Foreground(colorSelected).Bold(true)
	}
	return s
}
