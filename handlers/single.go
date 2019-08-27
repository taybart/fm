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
		case 'e':
			cmd.Set("edit")
		case 'u':
			cmd.Set("undo")
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
			cmd.Set("yank")
		}
		trailing = '⌘'
	case 'z':
		switch r {
		case 'h':
			cmd.Set("toggleHidden")
		}
		trailing = '⌘'
	default:
		switch r {
		case 'd':
			cmd.Set("{ d: delete }")
			state = single
		case 'i':
			cmd.Set("inspect")
		case 'y':
			cmd.Set("{ y: yank }")
			state = single
		case 'z':
			cmd.Set("{ h: toggleHidden }")
			state = single
		case '/':
			cmd.Set("fuzzy")
		}
		trailing = r
	}
	cd = cmd.Run(dt, cd)
	return cd
}
