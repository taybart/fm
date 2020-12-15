package handlers

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

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
func (c *Command) Run(dt *fs.Tree, cd string) string {
	didSomething := true
	cmdInput := strings.Split(c.Input, " ")
	switch cmdInput[0] {
	case "cd":
		name := strings.Join(cmdInput[1:], " ")
		if name[0] == '~' {
			name = path.Join(os.Getenv("HOME"), name[1:])
		}
		path, err := filepath.Abs(name)
		if err != nil {
			didSomething = false
			log.Error(err)
			break
		}
		err = dt.ChangeDirectory(path)
		if err != nil {
			log.Error(err)
		}
		cd = path
		log.Info(cd)
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
	case "inspect":
		err := inspect(dt, cd)
		if err != nil {
			log.Error("inspect", err)
		}
	case "edit":
		err := edit(dt, cd)
		if err != nil {
			log.Error("edit", err)
		}
	case "fuzzy":
		var err error
		cd, err = fuzzyCD(dt, cd)
		if err != nil {
			log.Error("fuzzy", err)
		}
	case "toggleHidden":
		toggleHidden()
		err := dt.Update(cd)
		if err != nil {
			log.Error("toggleHidden", err)
		}
	case "undo":
		err := digOutOfTrash()
		if err != nil {
			log.Error("undo", err)
		}
		err = dt.Update(cd)
		if err != nil {
			log.Error("undo", err)
		}
	case "s", "shell":
		err := runThis(os.Getenv("SHELL"))
		if err != nil {
			log.Error("shell", err)
		}
	case "refresh":
		dt.Update(cd)
	case "rename", "rn":
		name := strings.Join(cmdInput[1:], " ")
		fp := (*dt)[cd].ActiveFile.FullPath
		newPath := path.Join(fs.GetParentPath(fp), name)
		log.Verbose("Renaming", (*dt)[cd].ActiveFile.Name, name)
		err := os.Rename(fp, newPath)
		if err != nil {
			log.Error(err)
			break
		}
		err = dt.Update(cd)
		if err != nil {
			log.Error(err)
		}
		(*dt)[cd].SelectFileByName(name)
	case "q", "quit":
		Close()
	default:
		didSomething = false
	}
	if didSomething {
		runCommands = append(runCommands, c.Input)
		c.Reset()
		state = normal
	}
	return cd
}

// Complete : command
func (c *Command) Complete(dt *fs.Tree, cd string) {

	cmdInput := strings.Split(c.Input, " ")
	switch cmdInput[0] {
	case "rename", "rn":
		c.Input = fmt.Sprintf("%s %s", cmdInput[0], (*dt)[cd].ActiveFile.Name)
		c.Index = len(c.Input)
	}
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
func cmdRune(r rune, dt *fs.Tree, current string) string {
	switch r {
	default:
		valid := regexp.MustCompile(`[^[:cntrl:]]`)
		if valid.Match([]byte(string(r))) {
			cmd.Add(r)
			log.Verbose(cmd)
		}
	}
	return current
}

// cmd key handler
func cmdKeys(k tcell.Key, dt *fs.Tree, cd string) string {
	switch k {
	case tcell.KeyRight:
		cmd.UpdateIndex(1)
	case tcell.KeyLeft:
		cmd.UpdateIndex(-1)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		cmd.Del()

	case tcell.KeyTab:
		cmd.Complete(dt, cd)
	case tcell.KeyEnter:
		isShell := regexp.MustCompile(`^!`)
		if isShell.MatchString(cmd.Input) {
			c := strings.Split(cmd.Input[1:], " ")
			err := runThis(c[0], c[1:]...)
			if err != nil {
				log.Error(err)
			}
			err = dt.Update(cd)
			if err != nil {
				log.Error(err)
			}
			cmd.Reset()
			state = normal
		} else {
			cd = cmd.Run(dt, cd)
		}
	case tcell.KeyEsc:
		cmd.Reset()
		state = normal
	}
	return cd
}
