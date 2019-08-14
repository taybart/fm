package main

import (
	"fmt"
	"os"
	// "time"
	// "reflect"

	"github.com/gdamore/tcell"
	"github.com/taybart/fm/config"
	"github.com/taybart/fm/display"
	"github.com/taybart/fm/fs"
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
	dt, err := fs.Init(conf, cd)
	if err != nil {
		fmt.Println(err)
		log.Fatal(err)
	}

	display.Init(conf)
	defer display.Close()

	parent := fs.GetParentPath(cd)
	for {
		w := display.Window{Parent: *dt[parent], Current: *dt[cd]}
		if _, ok := dt[(dt)[cd].ActiveFile.FullPath]; ok {
			log.Info(dt[cd].ActiveFile.FullPath)
			w.Child = *dt[((dt)[cd]).ActiveFile.FullPath]
		}
		display.Draw(w) //, Child: *dt[cd+dt[cd].ActiveFile.FullPath]})
		event := display.PollEvents()
		switch ev := event.(type) {
		case *tcell.EventKey:
			if ev.Rune() == 'q' {
				display.Close()
				os.Exit(0)
			}
			if ev.Rune() == 'j' {
				dt[cd].Active++
				dt[cd].ActiveFile = dt[cd].Files[dt[cd].Active]
				err := dt.CD(cd)
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
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
	log.SetOutput(conf.Folder + "/fm.log")
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
