package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/taybart/log"
	"io"
	"os"
	"os/exec"
	"strings"
)

type mode int

const (
	normal mode = iota
	command
	single
	confirm
)

var navtree = []string{}

var a = struct { // action
	cmd       string
	confirmed bool
	lagmode   mode
}{confirmed: false}

var deletedFiles = []string{}

func (s *fmState) KeyParser(ev termbox.Event) {
	ch := ev.Ch
	key := ev.Key

	switch s.mode {
	case single:
		s.cmd = string(ev.Ch)
		s.mode = normal
		s.RunLetterCommand()
	case command:
		switch key {
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
	case confirm:
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
	case normal:
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
		case ch == 'j':
			if len(s.dir) > 0 {
				s.dt[s.cd].active++
			}
		case key == termbox.KeyCtrlK:
			if len(s.dir) > 0 {
				s.dt[s.cd].active -= conf.JumpAmount
			}
		case ch == 'k':
			if len(s.dir) > 0 {
				s.dt[s.cd].active--
			}
		/* Special */
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
		default:
			// push input onto stack
		}
	}
}

func fzf(input func(in io.WriteCloser)) []string {
	termbox.Close()
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "sh"
	}
	cmd := exec.Command(shell, "-c", "fzf", "-m")
	cmd.Stderr = os.Stderr
	in, _ := cmd.StdinPipe()
	go func() {
		input(in)
		in.Close()
	}()
	result, _ := cmd.Output()
	setupDisplay()
	return strings.Split(string(result), "\n")
}

func fuzzyFind(s *fmState) {
	filtered := fzf(func(in io.WriteCloser) {
		for _, f := range s.dir {
			fmt.Fprintln(in, f.name)
		}
	})

	for i, f := range s.dir {
		if filtered[0] == f.name {
			s.dt[s.cd].active = i
		}
	}
}

func (s *fmState) changeDirectory(file string) {
	dn := s.cd + "/" + s.active.name
	if _, ok := s.dt[dn]; !ok {
		s.dt[dn] = &dir{active: 0}
	}
	os.Chdir(dn)
}

func (s *fmState) RunLetterCommand() {
	a.lagmode = single
	switch s.cmd {
	case "d":
		deleteFile(s)
	case "D":
		deleteFileFull(s)
	case "u":
		undeleteFile()
	case "e":
		editFile(s.active)
	case "h":
		conf.ShowHidden = !conf.ShowHidden
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
		render()
	}
	cmd := args[0][1:]
	switch cmd {
	case "cd":
		os.Chdir(args[1])
	case "d", "delete":
		deleteFile(s)
	case "D":
		deleteFileFull(s)
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
	case "q", "quit":
		finalize()
	}
	// check that we are done and clear
	if s.mode == normal {
		s.cmd = ""
	}
}

func (s *fmState) getConfirmation(action string) {
	s.cmd = "Confirm " + action + " [Yy]: "
	s.mode = confirm
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

func editFile(file pseudofile) {
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		panic("No $EDITOR defined")
	}
	runThis(editor, file.name)
}
func renameFile(file pseudofile, newName string) {
	err := os.Rename(file.name, newName)
	if err != nil {
		panic(err)
	}
}

func copyFile() {
}

func deleteFileFull(s *fmState) {
	if a.confirmed {
		os.Remove(s.active.name)
	} else {
		a.cmd = s.cmd
		s.getConfirmation("deletion")
	}
}
func deleteFile(s *fmState) {
	if a.confirmed {
		moveToTrash(s.active.name)
	} else {
		a.cmd = s.cmd
		s.getConfirmation("deletion")
	}
}
func undeleteFile() {
	if len(deletedFiles) != 0 {
		home, _ := os.LookupEnv("HOME")
		t := deletedFiles
		last := t[len(t)-1]
		deletedFiles = t[:len(t)-1] // pop
		tf := home + "/.tmp/gofm_trash/" + last
		os.Rename(tf, last)
	}
}

func moveToTrash(fn string) {
	home, _ := os.LookupEnv("HOME")
	if exists, err := fileExists(home + "/.tmp/gofm_trash/"); !exists {
		if err != nil {
			log.Errorln(err)
		}
		err = os.MkdirAll(home+"/.tmp/gofm_trash/", os.ModeDir|0755)
		if err != nil {
			log.Errorln(err)
		}
	}
	os.Rename(fn, home+"/.tmp/gofm_trash/"+fn)
	deletedFiles = append(deletedFiles, fn)
}

func takeOutTrash() {
	home, _ := os.LookupEnv("HOME")
	os.RemoveAll(home + "/.tmp/gofm_trash/")
	os.MkdirAll(home+"/.tmp/gofm_trash/", os.ModeDir|0755)
}

func finalize() {
	termbox.Close()
	takeOutTrash()
	os.Exit(0)
}
