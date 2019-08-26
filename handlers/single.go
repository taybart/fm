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
			cmd.Run(dt, cd)
		}
		trailing = '⌘'
	case 'd':
		switch r {
		case 'd':
			log.Info("cut")
			cmd.Set("cut")
			cmd.Run(dt, cd)
		}
		trailing = '⌘'
	case 'e':
		switch r {
		case 'd':
			log.Info("delete")
			cmd.Set("delete")
			cmd.Run(dt, cd)
		}
		trailing = '⌘'
	case 'p':
		switch r {
		case 'p':
			log.Info("paste")
			cmd.Set("paste")
			cmd.Run(dt, cd)
		}
		trailing = '⌘'
	case 'y':
		switch r {
		case 'y':
			log.Info("yank")
			cmd.Set("yank")
			cmd.Run(dt, cd)
		}
		trailing = '⌘'
	default:
		switch r {
		case 'd':
			cmd.Set("{ d: delete }")
			state = single
		case 'c':
			cmd.Set("{ d: change dir }")
			state = single
		case 'y':
			cmd.Set("{ y: yank }")
			state = single
		case '/':
			fuzzyFind((*dt)[cd])
		}
		trailing = r
	}
	return cd
}
