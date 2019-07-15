package main

import (
	// "github.com/nsf/tcell-go"
	"github.com/gdamore/tcell"
	// "github.com/taybart/log"
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

func (s *fmState) ParseKeyEvent(ev *tcell.EventKey) {
	switch s.mode {
	case single:
		s.cmd = string(ev.Rune())
		s.mode = normal
		s.RunLetterCommand()
	case command:
		go s.parseCommmandMode(ev)
	case confirm:
		s.parseConfirmMode(ev)
	case normal:
		s.parseNormalMode(ev)
	}
	s.lastInput = ev.Rune()
}

func (s *fmState) parseNormalMode(ev *tcell.EventKey) {
	switch ev.Rune() {
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
		if s.active.isLink && !s.active.link.broken {
			if f, err := os.Stat(s.active.link.location); f.IsDir() && err == nil {
				dn = s.active.link.location
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
		s.cmdIndex = 1
		s.mode = command
	case 's':
		s.cmd = ":!"
		s.cmdIndex = 2
		s.mode = command
	case 'S':
		newShell()
	case '/':
		fuzzyFind(s)
	case 'q':
		finalize()
	case ' ':
		if _, exist := s.selectedFiles[s.active.fullPath]; !exist {
			s.selectedFiles[s.active.fullPath] = s.active
		} else {
			delete(s.selectedFiles, s.active.fullPath)
		}
		s.dt[s.cd].active++
	}
	switch ev.Key() {
	case tcell.KeyCtrlJ:
		if len(s.dir) > 0 && (s.dt[s.cd].active < len(s.dir)-1) {
			s.dt[s.cd].active += conf.JumpAmount
		}
	case tcell.KeyCtrlK:
		if len(s.dir) > 0 {
			s.dt[s.cd].active -= conf.JumpAmount
		}
	/* case tcell.MouseWheelUp:
		if len(s.dir) > 0 {
			s.dt[s.cd].active--
		}
	case tcell.MouseWheelDown:
		if len(s.dir) > 0 {
			s.dt[s.cd].active++
		}
	case tcell.MouseLeft:
		s.dt[s.cd].active = ev.MouseY - 1 */
	case tcell.KeyEsc:
		s.selectedFiles = make(map[string]pseudofile) // clear selected files
	}
}

func (s *fmState) parseConfirmMode(ev *tcell.EventKey) {
	s.mode = normal
	if ev.Rune() == 'Y' || ev.Rune() == 'y' {
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

func (s *fmState) parseCommmandMode(ev *tcell.EventKey) {
	switch ev.Key() {
	case tcell.KeyEsc:
		s.mode = normal
		s.cmd = ""
		s.cmdIndex = 0
	case tcell.KeyEnter:
		s.mode = normal
		if len(s.cmd) > 1 {
			s.RunFullCommand()
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if len(s.cmd) > 1 && s.cmdIndex > 1 {
			s.cmdIndex--
			s.cmd = string(append([]rune(s.cmd[:s.cmdIndex]), []rune(s.cmd[s.cmdIndex+1:])...))
		} else if len(s.cmd) == 1 {
			s.cmd = ""
			s.cmdIndex = 0
			s.mode = normal
		}
	case tcell.KeyRight:
		if s.cmd != "" && s.cmdIndex < len(s.cmd) {
			s.cmdIndex++
		}
	case tcell.KeyLeft:
		if s.cmd != "" && s.cmdIndex > 1 {
			s.cmdIndex--
		}
	case tcell.KeyTab:
		if s.cmd[:3] == ":rn" {
			s.cmd = ":rn " + s.active.name
			s.cmdIndex += len(s.active.name) + 1
		}
	default:
		s.cmd = s.cmd[:s.cmdIndex] + string(ev.Rune()) + s.cmd[s.cmdIndex:]
		s.cmdIndex++
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
			renameFile(s.active, strings.Join(args[1:], " "))
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
