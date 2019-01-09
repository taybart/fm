package main

import (
	"github.com/nsf/termbox-go"
	"os"
	"os/exec"
)

type command int

const (
	noop command = iota
	updown
	changedir
)

func newShell() {
	shell, exists := os.LookupEnv("SHELL")
	if !exists {
		panic("No $SHELL defined")
	}
	runThis(shell)
}

func runThis(toRun string, args ...string) error {
	termbox.Close()
	cmd := exec.Command(toRun, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	setupDisplay()
	return nil
}
