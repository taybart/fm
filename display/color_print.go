package display

import (
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

func drawFilePreview(offset, width int, fname string) error {
	log.Info("Preview", fname)
	source, err := getFileContents(fname)
	// Determine lexer.
	l := lexers.Match(fname)
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	// Determine style.
	style := styles.Get("solarized-dark256")
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
				s.Bold(true)
			}
			if entry.Underline == chroma.Yes {
				s.Underline(true)
			}
			if entry.Colour.IsSet() {
				fg := tcell.NewHexColor(int32(entry.Colour))
				s = s.Foreground(fg)
			}
			/* if entry.Background.IsSet() {
				bg := tcell.NewHexColor(int32(entry.Background))
				s = s.Background(bg)
			} */
		}
		// log.Infof("%#v\n", token.Value)
		newline := regexp.MustCompile("^\n+")
		if newline.Match([]byte(token.Value)) {
			line++
			// log.Verbose("NewLine", line, len(token.Value))
			col = offset
			tab := regexp.MustCompile("\t+")
			if tab.Match([]byte(token.Value)) {
				log.Infof("%#v\n", token.Value)
				col += 2
			}
		}
		str := strings.ReplaceAll(token.Value, "\n", " ")
		puts(col, line, width, str, false, s)
		col += len(token.Value)
	}

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