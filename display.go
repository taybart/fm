package main

import (
	"github.com/nsf/termbox-go"
	// "github.com/taybart/log"
	"fmt"
	"os"
)

func printLoc(x, y int, format string, v ...interface{}) {
	fg := termbox.ColorDefault
	bg := termbox.ColorDefault
	s := fmt.Sprintf(format, v...)
	printString(x, y, s, fg, bg)
}

func drawDir(active int, dir []os.FileInfo) {
	for i, f := range dir {
		str := f.Name()
		if f.IsDir() {
			str += "/"
		}
		fg := termbox.ColorDefault
		bg := termbox.ColorDefault
		if active == i {
			bg = termbox.ColorBlue
		}
		printString(5, i, str, fg, bg)
	}
}

func printString(x, y int, s string, fg, bg termbox.Attribute) {
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
