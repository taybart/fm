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
	single
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
	takeOutTrash()
	end <- true
}

// Keys handler
func Keys(ev *tcell.EventKey, dt *fs.Tree, cd string) HandlerReturn {
	switch state {
	case normal:
		cd = keys(ev.Key(), dt, cd)
		cd = runes(ev.Rune(), dt, cd)
	case command:
		cmdKeys(ev.Key(), dt, cd)
		cmdRune(ev.Rune(), dt, cd)
	case single:
		singleBuilder(ev.Rune(), dt, cd)
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
	case 'q':
		Close()
	default:
		cd = singleBuilder(r, dt, cd)
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
		cmd.Reset()
		state = normal
		// s.selectedFiles = make(map[string]pseudofile) // clear selected files
	}
	return cd
}