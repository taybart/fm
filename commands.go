package main

import (
	"github.com/nsf/termbox-go"
	"os"
	// "strings"
)

type mode int

const (
	normal mode = iota
	input
)

func (s *goFMState) KeyParser(ev termbox.Event) {
	switch s.mode {
	case input:
		switch ev.Key {
		case termbox.KeyEsc:
			s.mode = normal
			s.cmd = ""
		case termbox.KeyEnter:
			s.mode = normal
			s.RunCommand()
			s.cmd = ""
		default:
			s.cmd += string(ev.Ch)
		}
	case normal:
		switch ev.Ch {
		case 'h':
			if s.cd != "/" {
				os.Chdir("../")
			}
		case 'l':
			if len(s.dir) > 0 {
				if s.active.IsDir() {
					dn := s.cd + "/" + s.active.Name()
					if _, ok := s.dt[dn]; !ok {
						s.dt[dn] = &dir{active: 0}
					}
					os.Chdir(dn)
				}
			}
		case 'j':
			if len(s.dir) > 0 && (s.dt[s.cd].active < len(s.dir)-1) {
				s.dt[s.cd].active++
			}
		case 'k':
			if len(s.dir) > 0 {
				s.dt[s.cd].active--
				if s.dt[s.cd].active < 0 {
					s.dt[s.cd].active = 0
				}
			}
		case ':':
			s.cmd = ":"
			s.mode = input
		case 'S':
			newShell()
		case 'q':
			termbox.Close()
			os.Exit(0)
		}
	}
}

func (s *goFMState) RunCommand() {
	// activeFile := s.dir[s.dt[s.cd].active]
	/* if s.cmd[0] == '!' {
		cmd := strings.Split(s.cmd, " ")
		render()
	} */
	switch s.cmd[1:] {
	case "sh":
		newShell()
	case "q":
		termbox.Close()
		os.Exit(0)
	}
}

func meetsExitCondition(k termbox.Key) bool {
	return (k == termbox.KeyEnter || k == termbox.KeyEsc)
}

func renameFile() {
}

func copyFile() {
}

func deleteFile() {
}
