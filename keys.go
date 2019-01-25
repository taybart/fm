package main

import (
	"github.com/nsf/termbox-go"
	"os"
	"strings"
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
		go s.parseCommmandMode(ev)
	case confirm:
		s.parseConfirmMode(ev)
	case normal:
		s.parseNormalMode(ev)
	}
	s.lastInput = ev.Ch
}

func (s *fmState) parseNormalMode(ev termbox.Event) {
	switch ev.Ch {
	/* Movement */
	case 'h':
		if s.cd != "/" {
			if len(navtree) > 0 {
				dn := navtree[len(navtree)-1]
				navtree = navtree[:len(navtree)-1]
				os.Chdir(dn)
			} else {
				os.Chdir("../")
			}
		}
	case 'l':
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
	case 'j':
		if len(s.dir) > 0 {
			s.dt[s.cd].active++
		}
	case 'k':
		if len(s.dir) > 0 {
			s.dt[s.cd].active--
		}
		/* Special */
	case 'd':
		switch s.lastInput {
		case 'c':
			s.mode = command
			s.cmd = ":cd "
		case 'd':
			copyFile(s)
			s.moveFile = true
		}
	case 'y':
		// yy
		if s.lastInput == 'y' {
			copyFile(s)
		}
	case 'p':
		// pp
		if s.lastInput == 'p' {
			if s.moveFile {
				moveFile(s)
				s.moveFile = false
			} else {
				pasteFile(s)
			}
		}
	case 'i':
		inspectFile(s.active)
	case 'e', 'z':
		s.mode = single
	case ':':
		s.cmd = ":"
		s.mode = command
	case 's':
		s.cmd = ":!"
		s.mode = command
	case 'S':
		newShell()
	case '/':
		fuzzyFind(s)
	case 'q':
		finalize()
	}
	switch ev.Key {
	case termbox.KeySpace:
		if _, ok := s.selectedFiles[s.active.name]; !ok {
			s.selectedFiles[s.active.name] = false
		}
		s.selectedFiles[s.active.name] = !s.selectedFiles[s.active.name]
		s.dt[s.cd].active++
	case termbox.KeyCtrlJ:
		if len(s.dir) > 0 && (s.dt[s.cd].active < len(s.dir)-1) {
			s.dt[s.cd].active += conf.JumpAmount
		}
	case termbox.KeyCtrlK:
		if len(s.dir) > 0 {
			s.dt[s.cd].active -= conf.JumpAmount
		}
	case termbox.MouseWheelUp:
		if len(s.dir) > 0 {
			s.dt[s.cd].active--
		}
	case termbox.MouseWheelDown:
		if len(s.dir) > 0 {
			s.dt[s.cd].active++
		}
	case termbox.MouseLeft:
		s.dt[s.cd].active = ev.MouseY - 1
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
		} else {
			s.cmd = ""
			s.mode = normal
		}
	case termbox.KeySpace:
		s.cmd += " "
	default:
		s.cmd += string(ev.Ch)
	}
}

func (s *fmState) RunLetterCommand() {
	a.lagmode = single
	switch s.cmd {
	case "d":
		deleteFile(s)
	case "D":
		deleteFileWithoutTrash(s)
	case "u":
		undeleteFile()
	case "e":
		editFile(s.active)
	case "h":
		conf.ShowHidden = !conf.ShowHidden
	case "y":
		copyFile(s)
	case "p":
		pasteFile(s)
	}
	// check that we are done and clear
	if s.mode == normal {
		s.cmd = ""
	}

}
func (s *fmState) RunFullCommand() {

	a.lagmode = command
	args := strings.Split(s.cmd, " ")
	if s.cmd[1] == '!' {
		cmd := strings.Split(args[0], "!")
		runThis(cmd[1], args[1:]...)
		draw(s)
	} else {
		cmd := args[0][1:]
		switch cmd {
		case "cd":
			s.changeDirectory(args[1])
		case "d", "delete":
			deleteFile(s)
		case "D":
			deleteFileWithoutTrash(s)
		case "ud", "undelete":
			undeleteFile()
		case "rn", "rename":
			renameFile(s.active, args[1])
		case "e", "edit":
			editFile(s.active)
		case "th":
			conf.ShowHidden = !conf.ShowHidden
		case "sh", "shell":
			newShell()
		case "y", "copy":
			copyFile(s)
		case "p", "paste":
			pasteFile(s)
		case "q", "quit":
			finalize()
		}
	}
	// check that we are done and clear
	if s.mode == normal {
		s.cmd = ""
	}
}
