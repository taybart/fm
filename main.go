package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
	"github.com/taybart/fm/config"
	"github.com/taybart/fm/display"
	"github.com/taybart/fm/fs"
	"github.com/taybart/fm/handlers"
	"github.com/taybart/log"
)

var conf *config.Config

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

	cd := pwd()
	cmd := handlers.Command{}
	dt, err := fs.Init(conf, cd)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	display.Init(conf)
	defer display.Close()

	quit := make(chan bool)
	handlers.Init(conf, quit)

	go func() {
		for {
			parent := fs.GetParentPath(cd)
			w := display.Window{Current: *(*dt)[cd], Cmd: display.Command(cmd)}
			if parentDir, ok := (*dt)[parent]; ok && parent != "" {
				w.Parent = *parentDir
			}
			if childDir, ok := (*dt)[(*dt)[cd].ActiveFile.FullPath]; ok {
				w.Child = *childDir
			}
			display.Draw(w)
			event := display.PollEvents()
			switch ev := event.(type) {
			case *tcell.EventKey:
				hr := handlers.Keys(ev, dt, cd)
				cd = hr.CD
				cmd = hr.Cmd
			}
		}
	}()
	<-quit
	display.Close()
}

func pwd() string {
	cd, err := os.Getwd()
	if err != nil {
		log.Errorln(err)
	}
	return cd
}

func setupLog() {
	var err error
	home := os.Getenv("HOME")
	log.UseColors(false)
	log.SetTimeFmt("2006-01-02 15:04:05.9999")
	conf, err = config.Load(home + "/.config/fm/config.json")
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	err = log.SetOutput(conf.Folder + "/fm.log")
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}
	llevel := log.WARN
	if os.Getenv("ENV") == "development" {
		llevel = log.DEBUG
	}
	log.SetLevel(llevel)
}
