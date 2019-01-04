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
	termbox.Close()
	shell, exists := os.LookupEnv("SHELL")
	if !exists {
		panic("No $SHELL defined")
	}
	cmd := exec.Command(shell)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Run()
	setupDisplay()
}
