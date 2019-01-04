package main

import (
	"github.com/nsf/termbox-go"
	"github.com/taybart/log"
	"os"
)

var conf config

func main() {
	if os.Getenv("ENV") == "production" {
		log.SetLevel(log.WARN)
		log.SetOutput("./gofm.log")
		log.UseColors = false
	}

	setupDisplay()
	defer termbox.Close()

	var err error
	home := os.Getenv("HOME")
	conf, err = loadConfig(home + "/.config/gofm/config.json")
	if err != nil {
		log.Errorln(err)
	}

	cd := pwd()
	dt := make(directoryTree)
	dt[cd] = &dir{active: 0}

	for {
		cd = pwd()

		draw(dt, cd)

		files := readDir(".")
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventResize:
			draw(dt, cd)
		case termbox.EventKey:
			switch ev.Ch {
			case 'h':
				if cd != "/" {
					os.Chdir("../")
				}
			case 'l':
				if files[dt[cd].active].IsDir() {
					dn := cd + "/" + files[dt[cd].active].Name()
					if _, ok := dt[dn]; !ok {
						dt[dn] = &dir{active: 0}
					}
					os.Chdir(dn)
				}
			case 'j':
				(dt[cd]).active++
				dt[cd].active %= len(files)
			case 'k':
				dt[cd].active--
				if dt[cd].active < 0 {
					dt[cd].active = 0
				}
			case 'S':
				newShell()
			case 'q':
				termbox.Close()
				os.Exit(0)
			}
		}
	}
}
