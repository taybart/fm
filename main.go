package main

import (
	"github.com/nsf/termbox-go"
	"github.com/taybart/log"
	"io/ioutil"
	"os"
)

type directoryTree []dir

type dir struct {
	active int
	name   string
}

var c config

func main() {
	if os.Getenv("ENV") == "production" {
		log.SetLevel(log.WARN)
		log.SetOutput("./gofm.log")
		log.UseColors = false
	}

	setupDisplay()
	defer termbox.Close()

	var err error
	c, err = loadConfig("config.json")
	if err != nil {
		log.Errorln(err)
	}

	start := getWd()
	dt := directoryTree{dir{active: 0, name: start}}

	cd := 0
	for {
		files, err := ioutil.ReadDir(".")
		if err != nil {
			log.Errorln(err)
		}
		files = pruneDir(files)

		drawDir(dt[cd].active, files)

		render()
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Ch {
			case 'h':
				os.Chdir("../")
				cd--
				if cd < 0 {
					cd = 0
					dn := getWd()
					dt = dt.unshift(dir{active: 0, name: dn})
				}
			case 'l':
				if files[dt[cd].active].IsDir() {
					dn := files[dt[cd].active].Name()
					cd++
					if cd >= len(dt) {
						dt = dt.push(dir{active: 0, name: dn})
					}
					os.Chdir(dn)
				}
			case 'j':
				dt[cd].active++
				dt[cd].active %= len(files)
			case 'k':
				dt[cd].active--
				if dt[cd].active < 0 {
					dt[cd].active = 0
				}
			case 'q':
				termbox.Close()
				os.Exit(0)
			}
		}
	}
}

func (dt directoryTree) unshift(d dir) directoryTree {
	return append([]dir{d}, dt...)
}

func (dt directoryTree) push(d dir) directoryTree {
	return append(dt, d)
}

func pruneDir(dir []os.FileInfo) []os.FileInfo {
	pruned := []os.FileInfo{}
	for _, f := range dir {
		if rune(f.Name()[0]) == '.' {
			if c.ShowHidden {
				pruned = append(pruned, f)
			}
		} else {
			pruned = append(pruned, f)
		}
	}
	return pruned
}

func getWd() string {
	cd, err := os.Getwd()
	if err != nil {
		log.Errorln(err)
	}
	return cd
}
