package main

import (
	"fmt"
	// "github.com/nsf/termbox-go"
	"github.com/gdamore/tcell"
	"github.com/taybart/log"
	"os"
)

type fmState struct {
	cd            string
	dir           []pseudofile
	dt            directoryTree
	cmd           string
	active        pseudofile
	mode          mode
	copySource    pseudofile
	copyBufReady  bool
	moveFile      bool
	lastInput     rune
	selectedFiles map[string]pseudofile
}

var conf config

var scr tcell.Screen

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

	sf := make(map[string]pseudofile)
	s := &fmState{cmd: "", mode: normal, selectedFiles: sf}

	setupDisplay()
	defer scr.Fini()

	s.cd = pwd()
	s.dt = make(directoryTree)
	s.dt[s.cd] = &dir{active: 0}

	for {

		s.cd = pwd()
		s.dir, _, err = readDir(s.cd)
		if err != nil {
			panic(err)
		}

		// Bounds check
		if s.dt[s.cd].active > len(s.dir)-1 {
			s.dt[s.cd].active = len(s.dir) - 1
		}
		if s.dt[s.cd].active < 0 {
			s.dt[s.cd].active = 0
		}
		s.active = s.dir[s.dt[s.cd].active]

		draw(s)
		ev := scr.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventResize:
			draw(s)
		// case *tcell.EventKey, *tcell.EventMouse:
		case *tcell.EventKey:
			s.ParseKeyEvent(ev)
		}
	}
}

func setupLog() {
	var err error
	home := os.Getenv("HOME")
	log.UseColors(false)
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
