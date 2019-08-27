package handlers

import (
	"github.com/taybart/fm/fs"
	"github.com/taybart/log"
)

// Single key handler
func singleBuilder(r rune, dt *fs.Tree, cd string) string {
	switch trailing {
	case 'c':
		switch r {
		case 'd':
			log.Info("cd")
		}
		trailing = '⌘'
	case 'd':
		switch r {
		case 'd':
			cmd.Set("cut")
		}
		trailing = '⌘'
	case 'e':
		switch r {
		case 'd':
			cmd.Set("delete")
		}
		trailing = '⌘'
	case 'p':
		switch r {
		case 'p':
			cmd.Set("paste")
		}
		trailing = '⌘'
	case 'y':
		switch r {
		case 'y':
			log.Info("yank")
			cmd.Set("yank")
		}
		trailing = '⌘'
	default:
		switch r {
		// case 'c':
		// cmd.Set("{ d: change dir }")
		// state = single
		case 'd':
			cmd.Set("{ d: delete }")
			state = single
		case 'e':
			cmd.Set("edit")
		case 'i':
			cmd.Set("inspect")
		case 'y':
			cmd.Set("{ y: yank }")
			state = single
		case '/':
			cmd.Set("fuzzy")
		}
		trailing = r
	}
	cmd.Run(dt, cd)
	return cd
}
