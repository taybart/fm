package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"os"
	// "os/exec"
)

func debug(x, y int, format string, v ...interface{}) {
	fg := termbox.ColorDefault
	bg := termbox.ColorDefault
	s := fmt.Sprintf(format, v...)
	printString(x, y, s, fg, bg)
}

func drawDir(active int, dir []os.FileInfo, offset, width int) {
	for i, f := range dir {
		str := f.Name()
		if f.IsDir() {
			str += "/"
		}
		if len(str) > width-4 {
			str = str[:width-3] + "..."
		}
		for len(str) < width {
			str += " "
		}

		fg := termbox.ColorDefault
		bg := termbox.ColorDefault
		if active == i {
			bg = termbox.ColorBlue
		}
		printString(offset, i, str, fg, bg)
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
	cr := conf.ColumnRatios
	cw := conf.ColumnWidth

	if cw < 0 {
		cw, _ = termbox.Size()
	}

	parentPath := getParentPath(cd)
	parentFiles := readDir(parentPath)
	if _, ok := dt[parentPath]; !ok {
		dt[parentPath] = dt.newDirForParent(cd)
	}
	offset := 0
	width := int(float64(cr[0]) / 10.0 * float64(cw))
	drawDir(dt[parentPath].active, parentFiles, offset, width)

	files := readDir(".")

	offset = width
	width = int(float64(cr[1]) / 10.0 * float64(cw))
	drawDir(dt[cd].active, files, offset, width)

	if files[dt[cd].active].IsDir() {
		childPath := cd + "/" + files[dt[cd].active].Name()
		files := readDir(childPath)
		if _, ok := dt[childPath]; !ok {
			dt[childPath] = &dir{active: 0}
		}
		offset += width
		width = int(float64(cr[2]) / 10.0 * float64(cw))
		drawDir(dt[childPath].active, files, offset, width)
		/* } else {
		n := files[dt[cd].active].Name()
		cmd := exec.Command("strings", n)
		buf, _ := cmd.Output()
		printString(conf.ColumnRatios[2]*conf.ColumnWidth, 0,
			string(buf), termbox.ColorDefault, termbox.ColorDefault) */
	}
	render()
}
