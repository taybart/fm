package handlers

import (
	"regexp"

	"github.com/gdamore/tcell"
	"github.com/taybart/fm/display"
	"github.com/taybart/fm/fs"
	"github.com/taybart/log"
)

// Command : holds command input
type Command struct {
	Index  int
	Input  string
	Active bool
}

var cmd Command

var trailing rune
var runCommands []string

// Reset : command
func (c *Command) Reset() {
	c.Input = ""
	c.Index = 0
	c.Active = false
}

// Set : command
func (c *Command) Set(input string) {
	c.Input = input
	c.Index = len(input)
	c.Active = true
}

// Add : add rune to command
func (c *Command) Add(r rune) {
	c.Input = c.Input[:c.Index] + string(r) + c.Input[c.Index:]
	c.UpdateIndex(1)
}

// Del : add rune to command
func (c *Command) Del() {
	if len(c.Input) == 0 {
		c.Reset()
		state = normal
		return
	}
	if c.Index >= len(c.Input) {
		c.Input = c.Input[:c.Index-1]
	} else {
		c.Input = c.Input[:c.Index-1] + c.Input[c.Index:]
	}
	c.UpdateIndex(-1)
}

// UpdateIndex : add rune to command
func (c *Command) UpdateIndex(dir int) {
	index := c.Index
	index += dir
	if index >= len(c.Input) {
		index = len(c.Input)
	}
	if index < 0 {
		index = 0
	}
	c.Index = index
}

// Run : command
func (c *Command) Run(dt *fs.Tree, cd string) {
	isShell := regexp.MustCompile(`^\!`)
	switch {
	case isShell.MatchString(c.Input):
		log.Info("command")
	}
	switch c.Input {
	case "delete":
		err := deletef(dt, cd)
		if err != nil {
			log.Error("deletef", err)
		}
	case "yank", "cut":
		err := yank(dt, cd)
		if err != nil {
			log.Error(c.Input, err)
		}
	case "paste":
		err := paste(dt, cd)
		if err != nil {
			log.Error("paste", err)
		}
	case "q", "quit":
		Close()
	}
	runCommands = append(runCommands, c.Input)
	c.Reset()
	state = normal
}

func prompt(p string) string {
	display.Prompt(p)
	event := display.PollEvents()
	switch ev := event.(type) {
	case *tcell.EventKey:
		return string(ev.Rune())
	}
	return ""
}

// cmd rune handler
func cmdRune(r rune, dt *fs.Tree, current string) {
	switch r {
	default:
		valid := regexp.MustCompile(`[^[:cntrl:]]`)
		if valid.Match([]byte(string(r))) {
			cmd.Add(r)
		}
	}
}

// cmd key handler
func cmdKeys(k tcell.Key, dt *fs.Tree, cd string) {
	switch k {
	case tcell.KeyRight:
		cmd.UpdateIndex(1)
	case tcell.KeyLeft:
		cmd.UpdateIndex(-1)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		cmd.Del()
	case tcell.KeyEnter:
		cmd.Run(dt, cd)
	case tcell.KeyEsc:
		cmd.Reset()
		state = normal
	}
}
