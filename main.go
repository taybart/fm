package main

import (
	"github.com/nsf/termbox-go"
	"github.com/taybart/log"
	"os"
)

type goFMState struct {
	cd     string
	dir    []os.FileInfo
	dt     directoryTree
	cmd    string
	active os.FileInfo
	mode   mode
}

var conf config

func main() {
	if os.Getenv("ENV") == "production" {
		log.SetLevel(log.WARN)
		log.SetOutput("./gofm.log")
		log.UseColors = false
	}

	s := &goFMState{cmd: "", mode: normal}

	setupDisplay()
	defer termbox.Close()

	var err error
	home := os.Getenv("HOME")
	conf, err = loadConfig(home + "/.config/gofm/config.json")
	if err != nil {
		log.Errorln(err)
	}

	s.cd = pwd()
	s.dt = make(directoryTree)
	s.dt[s.cd] = &dir{active: 0}

	for {
		s.cd = pwd()
		s.dir = readDir(".")
		s.active = s.dir[s.dt[s.cd].active]

		draw(s.dt, s.cd, s.cmd)

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventResize:
			draw(s.dt, s.cd, s.cmd)
		case termbox.EventKey:
			s.KeyParser(ev)
		}
	}
}
