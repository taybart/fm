package handlers

import (
	"github.com/gdamore/tcell"
	"github.com/taybart/fm/fs"
)

// Command : holds command input
type Command struct {
	Index  int
	Input  string
	Active bool
}

var cmd Command

// Reset : reset command
func (c *Command) Reset() {
	c.Input = ""
	c.Index = 0
}

// Add : add rune to command
func (c *Command) Add(r rune) {
	cmd.Input += string(r)
	c.UpdateIndex(1)
}

// Del : add rune to command
func (c *Command) Del() {
	c.Input = c.Input[:c.Index-1] + c.Input[c.Index:]
	c.UpdateIndex(-1)
}

// UpdateIndex : add rune to command
func (c *Command) UpdateIndex(dir int) {
	index := c.Index
	index += dir
	if index >= len(c.Input) {
		index = len(c.Input) - 1
	}
	if index <= 0 {
		index = 0
	}
	cmd.Index = index
}

// cmd rune handler
func cmdRune(r rune, dt *fs.Tree, current string) {
	// log.Verbose(r)
	if r != 127 && r != 0 {
		cmd.Add(r)
	}
}

// cmd key handler
func cmdKeys(k tcell.Key, dt *fs.Tree, current string) {
	switch k {
	case tcell.KeyRight:
		cmd.UpdateIndex(1)
	case tcell.KeyLeft:
		cmd.UpdateIndex(-1)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		cmd.Del()
	case tcell.KeyEnter:
		state = run
		state = normal
		cmd.Active = false
	case tcell.KeyEsc:
		cmd.Reset()
		state = run
		state = normal
		cmd.Active = false
	}
}
