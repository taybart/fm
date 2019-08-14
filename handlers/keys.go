package handlers

import (
	// "fmt"

	"github.com/gdamore/tcell"
	"github.com/taybart/fm/config"
	"github.com/taybart/fm/fs"
	"github.com/taybart/log"
)

var conf *config.Config

var end chan bool

// Init the handlers
func Init(c *config.Config, done chan bool) {
	conf = c
	end = done
}

// Close end
func Close() {
	end <- true
}

// Keys handler
func Keys(ev *tcell.EventKey, dt *fs.Tree, cd string) string {
	cd = runes(ev.Rune(), dt, cd)
	return cd
}

// runes handler
func runes(r rune, dt *fs.Tree, current string) string {
	cd := current
	switch r {
	/* Movement */
	case 'h':
		parent := fs.GetParentPath(cd)
		err := dt.ChangeDirectory(parent)
		cd = parent
		if err != nil {
			log.Error(err)
			cd = current
		}
	case 'l':
		// down cd activefile
		child := (*dt)[cd].ActiveFile.FullPath
		err := dt.ChangeDirectory(child)
		cd = child
		if err != nil {
			cd = current
			log.Error(err)
		}
	case 'j':
		dt.SelectFile(1, cd)
	case 'k':
		dt.SelectFile(-1, cd)
	case 'q':
		Close()
	}
	return cd
}

/* func parseNormalMode(ev *tcell.EventKey, dt fs.Tree, cd string) {
	switch ev.Rune() {
	[>Movement<]
	case 'h':
		if cd != "/" {
			if len(navtree) > 0 {
				dn := navtree[len(navtree)-1]
				navtree = navtree[:len(navtree)-1]
				os.Chdir(dn)
			} else {
				os.Chdir("../")
			}
		}
	case 'l':
		if len(s.dir) == 0 {
			break
		}

		dn := ""
		if s.active.isDir {
			dn = s.cd + "/" + s.active.name
		}
		if s.active.isLink && !s.active.link.broken {
			if f, err := os.Stat(s.active.link.location); f.IsDir() && err == nil {
				dn = s.active.link.location
			}
		}
		// if a new directory name exists go there
		if dn != "" {
			if s.cd == "/" {
				dn = s.cd + s.active.name
			}
			if _, ok := s.dt[dn]; !ok {
				s.dt[dn] = &dir{active: 0}
			}
			navtree = append(navtree, s.cd)
			os.Chdir(dn)
		}
	case 'j':
		if len(s.dir) > 0 {
			s.dt[s.cd].active++
		}
	case 'k':
		if len(s.dir) > 0 {
			s.dt[s.cd].active--
		}
		[>Special<]
	case 'd':
		switch s.lastInput {
		case 'c':
			s.mode = command
			s.cmd = ":cd "
		case 'd':
			copyFile(s)
			s.moveFile = true
		}
	case 'y':
		// yy
		if s.lastInput == 'y' {
			copyFile(s)
		}
	case 'p':
		// pp
		if s.lastInput == 'p' {
			if s.moveFile {
				moveFile(s)
				s.moveFile = false
			} else {
				pasteFile(s)
			}
		}
	case 'i':
		inspectFile(s.active)
	case 'e', 'z':
		s.mode = single
	case ':':
		s.cmd = ":"
		s.cmdIndex = 1
		s.mode = command
	case 's':
		s.cmd = ":!"
		s.cmdIndex = 2
		s.mode = command
	case 'S':
		newShell()
	case '/':
		fuzzyFind(s)
	case 'q':
		finalize()
	case ' ':
		if _, exist := s.selectedFiles[s.active.fullPath]; !exist {
			s.selectedFiles[s.active.fullPath] = s.active
		} else {
			delete(s.selectedFiles, s.active.fullPath)
		}
		s.dt[s.cd].active++
	}
	switch ev.Key() {
	case tcell.KeyCtrlJ:
		if len(s.dir) > 0 && (s.dt[s.cd].active < len(s.dir)-1) {
			s.dt[s.cd].active += conf.JumpAmount
		}
	case tcell.KeyCtrlK:
		if len(s.dir) > 0 {
			s.dt[s.cd].active -= conf.JumpAmount
		}
	case tcell.KeyEsc:
		s.selectedFiles = make(map[string]pseudofile) // clear selected files
	}
} */
