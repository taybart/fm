package handler

import (
	// "github.com/nsf/tcell-go"
	"github.com/gdamore/tcell"
	"github.com/taybart/log"
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

func HandleKeyEvent(s *fm, ev *tcell.EventKey) {
	switch s.mode {
	case single:
		s.cmd = string(ev.Rune())
		s.mode = normal
		s.RunLetterCommand()
	case command:
		s.parseCommmandMode(ev)
	case confirm:
		s.parseConfirmMode(ev)
	case normal:
		s.parseNormalMode(ev)
	}
	s.lastInput = ev.Rune()

	log.Verbose("ParseKeyEvent done")
}
