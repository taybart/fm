package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"io"
	"os"
	"os/exec"
	"strings"
)

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
		deleteFileWithoutTrash(s)
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
	case "q", "quit":
		finalize()
	}
	// check that we are done and clear
	if s.mode == normal {
		s.cmd = ""
	}
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
