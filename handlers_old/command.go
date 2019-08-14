package handler

import (
	"github.com/gdamore/tcell"
	"github.com/taybart/log"
)

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
	log.Verbose("parseCommmandMode done")
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
		readDir(s.cd)
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
