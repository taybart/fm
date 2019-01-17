package main

import (
	"github.com/nsf/termbox-go"
	"github.com/taybart/log"
	"os"
)

type goFMState struct {
	cd     string
	dir    []pseudofile
	dt     directoryTree
	cmd    string
	active pseudofile
	mode   mode
}

var conf config

func main() {
	var err error
	home := os.Getenv("HOME")
	log.UseColors = false
	log.SetOutput(home + "/.config/gofm/gofm.log")
	if os.Getenv("ENV") == "production" {
		log.SetLevel(log.WARN)
		conf, err = loadConfig(home + "/.config/gofm/config.json")
		if err != nil {
			panic(err)
		}
	} else {
		log.SetLevel(log.DEBUG)
		conf, err = loadConfig(home + "/.config/gofm/config.json")
		if err != nil {
			log.Errorln(err)
			os.Exit(1)
		}
	}

	s := &goFMState{cmd: "", mode: normal}

	setupDisplay()
	defer termbox.Close()

	s.cd = pwd()
	s.dt = make(directoryTree)
	s.dt[s.cd] = &dir{active: 0}

	for {

		s.dir, _, err = readDir(".")
		if err != nil {
			panic(err)
		}

		s.cd = pwd()
		// Bounds check
		if s.dt[s.cd].active > len(s.dir)-1 {
			s.dt[s.cd].active = len(s.dir) - 1
		}
		if s.dt[s.cd].active < 0 {
			s.dt[s.cd].active = 0
		}
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
