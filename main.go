package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/taybart/log"
	"os"
)

type fmState struct {
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
	_, sessionActive := os.LookupEnv("FM_SESSION_ACTIVE")
	if sessionActive {
		fmt.Println("Nesting sessions is not a wise idea")
		os.Exit(0)
	}
	os.Setenv("FM_SESSION_ACTIVE", "true")
	defer os.Unsetenv("FM_SESSION_ACTIVE")

	setupLog()

	s := &fmState{cmd: "", mode: normal}

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
		case termbox.EventKey, termbox.EventMouse:
			s.ParseKeyEvent(ev)
		}
	}
}

func setupLog() {
	var err error
	home := os.Getenv("HOME")
	log.UseColors = false
	conf, err = loadConfig(home + "/.config/fm/config.json")
	log.SetOutput(conf.Folder + "/fm.log")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	llevel := log.WARN
	if os.Getenv("ENV") != "production" {
		llevel = log.DEBUG
	}
	log.SetLevel(llevel)
}
