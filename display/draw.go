package display

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/taybart/fm/config"
	"github.com/taybart/fm/fs"
	"github.com/taybart/log"
)

var conf *config.Config

// Window holds just a window
type Window struct {
	Parent  fs.Directory
	Current fs.Directory
	Child   fs.Directory
}

// Scr screen
var scr tcell.Screen

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

// Init start screen
func Init(c *config.Config) {
	conf = c
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
	// scr.EnableMouse()
	scr.Clear()
}

// Close shut er down
func Close() {
	scr.Fini()
}

// Draw display current state
func Draw(w Window) {
	scr.Clear()

	tbw, tbh := scr.Size()
	if tbw <= 0 || tbh <= 0 {
		return
	}

	cr := conf.ColumnRatios
	cw := conf.ColumnWidth
	if cw < 0 {
		cw = tbw
	}

	offset := int(float64(cr[0]) / 10.0 * float64(cw))
	width := int(float64(cr[1]) / 10.0 * float64(cw))
	drawParentDir(w.Parent)
	drawDir(w.Current, offset, width)

	offset = int(float64(cr[0])/10.0*float64(cw)) +
		int(float64(cr[1])/10.0*float64(cw))
	width = int(float64(cr[2]) / 10.0 * float64(cw))
	drawChildDir(w.Child, offset, width)

	drawHeader(w.Current)
	// drawFooter(files)

	scr.Show()
	scr.Sync()
}

func drawHeader(dir fs.Directory) {
	tbwidth, _ := scr.Size()
	// Print user/cd at top
	un := os.Getenv("USER")
	hn, err := os.Hostname()
	if err != nil {
		log.Errorln(err)
	}
	ustr := un + "@" + hn
	puts(0, 0, tbwidth, ustr, true, tcell.StyleDefault.Foreground(tcell.ColorGreen))
	dn := dir.ActiveFile.FullPath
	cd := dir.Path
	oset := 0
	if cd != "/" {
		dn += "/"
		oset = 1
	}

	puts(len(ustr)+1, 0, tbwidth, dn, true, tcell.StyleDefault.Foreground(tcell.ColorGreen))
	f := dir.ActiveFile
	name := f.Name
	if f.IsDir {
		name += "/"
	}
	puts(len(ustr)+len(cd)+1+oset, 0, tbwidth, name,
		true, tcell.StyleDefault)
}

/* func drawFooter(files []pseudofile) {
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
} */

// PollEvents get tcell events
func PollEvents() tcell.Event {
	return scr.PollEvent()
	// log.Verbose(reflect.TypeOf(event).String())
}
