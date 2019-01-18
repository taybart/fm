package main

import (
	"github.com/nsf/termbox-go"
	"os"
)

// Mode used to determine what to do
type mode int

const (
	normal mode = iota
	command
	single
	confirm
)

// navtree is used for corner cases when following symlinks
var navtree = []string{}

// action placeholder, used in cases like confirmation
var a = struct {
	cmd       string
	confirmed bool
	lagmode   mode
}{confirmed: false}

var deletedFiles = []string{}

func (s *fmState) ParseKeyEvent(ev termbox.Event) {
	switch s.mode {
	case single:
		s.cmd = string(ev.Ch)
		s.mode = normal
		s.RunLetterCommand()
	case command:
		s.parseCommmandMode(ev)
	case confirm:
		s.parseConfirmMode(ev)
	case normal:
		s.parseNormalMode(ev)
	}
}

func (s *fmState) parseNormalMode(ev termbox.Event) {
	ch := ev.Ch
	key := ev.Key
	switch {
	/* Movement */
	case ch == 'h':
		if s.cd != "/" {
			if len(navtree) > 0 {
				dn := navtree[len(navtree)-1]
				navtree = navtree[:len(navtree)-1]
				os.Chdir(dn)
			} else {
				os.Chdir("../")
			}
		}
	case ch == 'l':
		if len(s.dir) == 0 {
			break
		}

		dn := ""
		if s.active.isDir {
			dn = s.cd + "/" + s.active.name
		}
		if s.active.isSymL {
			if f, err := os.Stat(s.active.symName); f.IsDir() && err == nil {
				dn = s.active.symName
			}
		}
		// if a new directory name exists go there
		if dn != "" {
			if s.cd == "/" {
				dn = s.cd + s.active.name
			}
			if _, ok := s.dt[dn]; !ok {
				s.dt[dn] = &dir{active: 0}
			}
			navtree = append(navtree, s.cd)
			os.Chdir(dn)
		}
	case key == termbox.KeyCtrlJ:
		if len(s.dir) > 0 && (s.dt[s.cd].active < len(s.dir)-1) {
			s.dt[s.cd].active += conf.JumpAmount
		}
	case ch == 'j', key == termbox.MouseWheelDown:
		if len(s.dir) > 0 {
			s.dt[s.cd].active++
		}
	case key == termbox.KeyCtrlK:
		if len(s.dir) > 0 {
			s.dt[s.cd].active -= conf.JumpAmount
		}
	case ch == 'k', key == termbox.MouseWheelUp:
		if len(s.dir) > 0 {
			s.dt[s.cd].active--
		}
	/* Special */
	case key == termbox.MouseLeft:
		s.dt[s.cd].active = ev.MouseY - 1
	case ch == 'i':
		inspectFile(s.active)
	case ch == 'e', ch == 'z':
		s.mode = single
	case ch == ':':
		s.cmd = ":"
		s.mode = command
	case ch == 'S':
		newShell()
	case ch == '/':
		fuzzyFind(s)
	case ch == 'q':
		finalize()
	}
}

func (s *fmState) parseConfirmMode(ev termbox.Event) {
	s.mode = normal
	if ev.Ch == 'Y' || ev.Ch == 'y' {
		s.cmd = a.cmd
		a.confirmed = true
		if a.lagmode == command {
			s.RunFullCommand()
		}
		if a.lagmode == single {
			s.RunLetterCommand()
		}
		a.confirmed = false
	}
	s.cmd = ""
}

func (s *fmState) parseCommmandMode(ev termbox.Event) {
	switch ev.Key {
	case termbox.KeyEsc:
		s.mode = normal
		s.cmd = ""
	case termbox.KeyEnter:
		s.mode = normal
		if len(s.cmd) > 1 {
			s.RunFullCommand()
		}
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if len(s.cmd) > 1 {
			s.cmd = s.cmd[:len(s.cmd)-1]
		}
	case termbox.KeySpace:
		s.cmd += " "
	default:
		s.cmd += string(ev.Ch)
	}
}
