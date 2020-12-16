package display

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/gdamore/tcell/v2/encoding"
	"github.com/taybart/fm/config"
	"github.com/taybart/fm/fs"
	"github.com/taybart/log"
)

var conf *config.Config

// Command : cmd
type Command struct {
	Index  int
	Input  string
	Active bool
}

// Window holds just a window
type Window struct {
	Cmd     Command
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
	drawChildDir(w, offset, width)

	drawHeader(w.Current)
	drawFooter(w)

	// scr.Sync()
	scr.Show()
}

// Prompt user
func Prompt(p string) {
	tbwidth, tbheight := scr.Size()
	// scr.Clear()
	for i := 0; i < tbwidth-len(p); i++ {
		p += " "
	}
	puts(0, tbheight-1, tbwidth, p, true, tcell.StyleDefault)
	scr.Show()
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

func drawFooter(w Window) {
	tbwidth, tbheight := scr.Size()
	if w.Cmd.Active {
		scr.SetContent(0, tbheight-1, ':', nil, tcell.StyleDefault)
		puts(1, tbheight-1, tbwidth, w.Cmd.Input, true, tcell.StyleDefault)

		c := ' '
		if w.Cmd.Index < len(w.Cmd.Input) {
			c = rune(w.Cmd.Input[w.Cmd.Index])
		}

		style := tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorWhite)
		scr.SetContent(w.Cmd.Index+1, tbheight-1, c, nil, style)
	} else {
		f := w.Current.ActiveFile
		if f.IsReal {
			s := fmt.Sprintf("%s %d %s %s",
				f.F.Mode(), f.F.Size(),
				f.F.ModTime().Format("Jan 2 15:04"), f.Name)
			puts(0, tbheight-1, tbwidth, s, true, tcell.StyleDefault)
		}
	}
}

// PollEvents get tcell events allows scr to not be exported
func PollEvents() tcell.Event {
	return scr.PollEvent()
}
