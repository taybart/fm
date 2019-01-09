package main

import (
	"github.com/nsf/termbox-go"
	"os"
	"os/exec"
	"strings"
)

type mode int

const (
	normal mode = iota
	input
	confirm
)

func (s *goFMState) KeyParser(ev termbox.Event) {
	ch := ev.Ch
	key := ev.Key
	switch s.mode {
	case input:
		switch key {
		case termbox.KeyEsc:
			s.mode = normal
			s.cmd = ""
		case termbox.KeyEnter:
			if len(s.cmd) > 1 {
				s.RunCommand()
			}
			s.mode = normal
			s.cmd = ""
		case termbox.KeyBackspace, termbox.KeyBackspace2:
			if len(s.cmd) > 1 {
				s.cmd = s.cmd[:len(s.cmd)-1]
			}
		case termbox.KeySpace:
			s.cmd += " "
		default:
			s.cmd += string(ev.Ch)
		}
	case normal:
		switch {
		/* Movement */
		case ch == 'h':
			if s.cd != "/" {
				os.Chdir("../")
			}
		case ch == 'l':
			if len(s.dir) > 0 {
				if s.active.IsDir() {
					dn := s.cd + "/" + s.active.Name()
					if _, ok := s.dt[dn]; !ok {
						s.dt[dn] = &dir{active: 0}
					}
					os.Chdir(dn)
				}
			}
		case key == termbox.KeyCtrlJ:
			if len(s.dir) > 0 && (s.dt[s.cd].active < len(s.dir)-1) {
				s.dt[s.cd].active += conf.JumpAmount
				if s.dt[s.cd].active > len(s.dir)-1 {
					s.dt[s.cd].active = len(s.dir) - 1
				}
			}
		case ch == 'j':
			if len(s.dir) > 0 && (s.dt[s.cd].active < len(s.dir)-1) {
				s.dt[s.cd].active++
			}
		case key == termbox.KeyCtrlK:
			if len(s.dir) > 0 {
				s.dt[s.cd].active -= conf.JumpAmount
				if s.dt[s.cd].active < 0 {
					s.dt[s.cd].active = 0
				}
			}
		case ch == 'k':
			if len(s.dir) > 0 {
				s.dt[s.cd].active--
				if s.dt[s.cd].active < 0 {
					s.dt[s.cd].active = 0
				}
			}
		/* Special */
		case ch == ':':
			s.cmd = ":"
			s.mode = input
		case ch == 'S':
			newShell()
		case ch == 'q':
			termbox.Close()
			os.Exit(0)
		}
	}
}

func (s *goFMState) RunCommand() {
	args := strings.Split(s.cmd, " ")
	if s.cmd[1] == '!' {
		cmd := strings.Split(args[0], "!")
		runThis(cmd[1], args[1:]...)
		render()
	}
	cmd := args[0][1:]
	switch cmd {
	case "d", "delete":
		deleteFile(s.active)
	case "rn", "rename":
		renameFile(s.active, args[1])
	case "e", "edit":
		editFile(s.active)
	case "sh", "shell":
		newShell()
	case "q", "quit":
		finalize()
	}
}

func meetsExitCondition(k termbox.Key) bool {
	return (k == termbox.KeyEnter || k == termbox.KeyEsc)
}

func runThis(toRun string, args ...string) error {
	termbox.Close()
	cmd := exec.Command(toRun, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		panic(err)
	}
	setupDisplay()
	return nil
}

func newShell() {
	shell, exists := os.LookupEnv("SHELL")
	if !exists {
		panic("No $SHELL defined")
	}
	runThis(shell)
}

func editFile(file os.FileInfo) {
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		panic("No $EDITOR defined")
	}
	runThis(editor, file.Name())
}
func renameFile(file os.FileInfo, newName string) {
	err := os.Rename(file.Name(), newName)
	if err != nil {
		panic(err)
	}
}

func copyFile() {
}

func deleteFile(file os.FileInfo) {
	if getConfirmation("deletion") {
		moveToTrash(file)
	}
}

func moveToTrash(file os.FileInfo) {
}

func getConfirmation(action string) bool {
	printPrompt("Confirm " + action + " [Yy]")
	switch ev := termbox.PollEvent(); ev.Type {
	case termbox.EventKey:
		switch ev.Ch {
		case 'y', 'Y':
			return true
		}
	}
	return false
}

func takeOutTrash() {

}

func finalize() {
	termbox.Close()
	takeOutTrash()
	os.Exit(0)
}
