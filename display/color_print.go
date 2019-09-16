package display

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/gdamore/tcell"
	"github.com/taybart/log"
)

var errFileTooLarge = errors.New("File too large for preview")

func drawFilePreview(offset, width int, fname string) error {
	log.Verbose("Preview", fname)
	source, err := getFileContents(fname)
	if err != nil {
		return err
	}
	l := lexers.Match(fname)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		// return fmt.Errorf("No lexer")
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	// Determine style.
	style := styles.Get("dracula")
	if style == nil {
		style = styles.Fallback
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		return err
	}

	line := 1
	col := offset
	for token := it(); token != chroma.EOF; token = it() {
		s := tcell.StyleDefault
		entry := style.Get(token.Type)
		if !entry.IsZero() {
			if entry.Bold == chroma.Yes {
				s = s.Bold(true)
			}
			if entry.Underline == chroma.Yes {
				s = s.Underline(true)
			}
			if entry.Colour.IsSet() {
				fg := tcell.NewHexColor(int32(entry.Colour))
				s = s.Foreground(fg)
			}
		}
		newline := regexp.MustCompile("^\n+")
		if newline.Match([]byte(token.Value)) {
			line++
			col = offset
			tab := regexp.MustCompile("\t+")
			if tab.Match([]byte(token.Value)) {
				col += 2
			}
		}
		str := strings.ReplaceAll(token.Value, "\n", " ")
		puts(col, line, width, str, false, s)
		col += len(token.Value)
	}
	scr.Sync()
	return nil
}

func getFileContents(fname string) (contents string, err error) {
	file, err := os.Open(fname)
	if err != nil {
		return
	}

	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return
	}

	if fi.Size() > 3*1024*1024 {
		log.Verbose("File is larger than 3MB, no preview")
		err = errFileTooLarge
		return
	}

	if fi.Mode().IsDir() {
		err = fmt.Errorf("%s is a directory", file.Name())
		return
	}
	s, err := ioutil.ReadFile(fname)
	if err != nil {
		return
	}
	return string(s), nil
}
