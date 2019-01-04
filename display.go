package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
)

func debug(x, y int, format string, v ...interface{}) {
	fg := termbox.ColorDefault
	bg := termbox.ColorDefault
	s := fmt.Sprintf(format, v...)
	printString(x, y, s, fg, bg)
}

func drawDir(active int, dir []os.FileInfo, offset int) {
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
		for len(str) < conf.ColumnWidth {
			str += " "
		}
		printString(5+offset, i, str, fg, bg)
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

func draw(dt directoryTree, cd string) {
	parentPath := getParentPath(cd)
	parentFiles := readDir(parentPath)
	if _, ok := dt[parentPath]; !ok {
		dt[parentPath] = dt.newDirForParent(cd)
	}
	drawDir(dt[parentPath].active, parentFiles, 0)

	files := readDir(".")
	drawDir(dt[cd].active, files, conf.ColumnWidth)

	if files[dt[cd].active].IsDir() {
		childPath := cd + "/" + files[dt[cd].active].Name()
		files := readDir(childPath)
		if _, ok := dt[childPath]; !ok {
			dt[childPath] = &dir{active: 0}
		}
		drawDir(dt[childPath].active, files, conf.ColumnWidth*2)
	}
	render()
}
