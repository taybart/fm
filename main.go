package main

import (
	// "fmt"
	"github.com/nsf/termbox-go"
	"github.com/taybart/log"
	"io/ioutil"
	"os"
)

func main() {

	setupDisplay()
	defer termbox.Close()

	dirIndex := 0
	cd, _ := os.Getwd()
	files, err := ioutil.ReadDir(cd)
	if err != nil {
		log.Errorln(err)
	}
	drawFileCol(dirIndex, files)
mainloop:
	for {
		cd, _ := os.Getwd()
		files, err := ioutil.ReadDir(cd)
		if err != nil {
			log.Errorln(err)
		}
		drawFileCol(dirIndex, files)

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Ch {
			case 'h':
				os.Chdir("../")
			case 'l':
				if files[dirIndex].IsDir() {
					os.Chdir(files[dirIndex].Name())
					dirIndex = 0
				}
			case 'j':
				dirIndex++
				dirIndex %= len(files)
			case 'k':
				dirIndex--
				if dirIndex < 0 {
					dirIndex = 0
				}
			case 'q':
				break mainloop
			}
		}
	}
}

func drawFileCol(dirIndex int, files []os.FileInfo) {
	for i, f := range files {
		active := false
		if i == dirIndex {
			active = true
		}
		printLine(0, i, f.Name(), active)
	}
	render()
}

func printLine(x, y int, s string, active bool) {
	fg := termbox.ColorDefault
	bg := termbox.ColorDefault
	if active {
		bg = termbox.ColorBlue
	}
	for _, c := range s {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func setupDisplay() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
	termbox.SetOutputMode(termbox.OutputNormal)
}

func render() {
	termbox.Flush()
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
}
