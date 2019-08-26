package handlers

import (
	"github.com/gdamore/tcell"
	"github.com/taybart/fm/config"
	"github.com/taybart/fm/fs"
	"github.com/taybart/log"
)

// HandlerReturn : at the end returns
type HandlerReturn struct {
	CD  string
	Cmd Command
}

var conf *config.Config

var end chan bool

// State that the handlers are in
var state int

const (
	normal = iota + 1
	command
	run
)

// Init the handlers
func Init(c *config.Config, done chan bool) {
	conf = c
	end = done
	state = normal
	cmd = Command{Input: "", Index: 0, Active: false}
}

// Close end
func Close() {
	end <- true
}

// Keys handler
func Keys(ev *tcell.EventKey, dt *fs.Tree, cd string) HandlerReturn {
	switch state {
	case normal:
		cd = runes(ev.Rune(), dt, cd)
		cd = keys(ev.Key(), dt, cd)
	case command:
		cmdRune(ev.Rune(), dt, cd)
		cmdKeys(ev.Key(), dt, cd)
	}
	return HandlerReturn{CD: cd, Cmd: cmd}
}

// runes handler
func runes(r rune, dt *fs.Tree, current string) string {
	cd := current
	switch r {
	/* Movement */
	case 'h':
		parent := fs.GetParentPath(cd)
		err := dt.ChangeDirectory(parent)
		cd = parent
		if err != nil {
			log.Error(err)
			cd = current
		}
	case 'l':
		// down cd activefile
		child := (*dt)[cd].ActiveFile.FullPath
		err := dt.ChangeDirectory(child)
		cd = child
		if err != nil {
			cd = current
			log.Error(err)
		}
	case 'j':
		err := dt.SelectFile(1, cd)
		if err != nil {
			log.Error(err)
		}
	case 'k':
		err := dt.SelectFile(-1, cd)
		if err != nil {
			log.Error(err)
		}
	case ':':
		cmd.Reset()
		cmd.Active = true
		state = command
	// case ' ':
	// dt.
	case 'q':
		Close()
	}
	return cd
}

// runes handler
func keys(k tcell.Key, dt *fs.Tree, current string) string {
	cd := current
	switch k {
	case tcell.KeyCtrlJ:
		err := dt.SelectFile(conf.JumpAmount, cd)
		if err != nil {
			log.Error(err)
		}
	case tcell.KeyCtrlK:
		err := dt.SelectFile(-1*conf.JumpAmount, cd)
		if err != nil {
			log.Error(err)
		}
	case tcell.KeyEsc:
		// s.selectedFiles = make(map[string]pseudofile) // clear selected files
	}
	return cd
}

/* func parseNormalMode(ev *tcell.EventKey, dt fs.Tree, cd string) {
	switch ev.Rune() {
	[>Movement<]
		[>Special<]
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
	case tcell.KeyEsc:
		s.selectedFiles = make(map[string]pseudofile) // clear selected files
	}
} */
