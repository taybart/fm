package main

import (
	"github.com/nsf/termbox-go"
	"github.com/taybart/log"
	"os"
)

func (s *fmState) getConfirmation(action string) {
	s.cmd = "Confirm " + action + " [Yy]: "
	s.mode = confirm
}

func newShell() {
	shell, exists := os.LookupEnv("SHELL")
	if !exists {
		panic("No $SHELL defined")
	}
	runThis(shell)
}

func inspectFile(file pseudofile) {
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		panic("No $EDITOR defined")
	}
	if editor == "vim" || editor == "nvim" {
		runThis(editor, "-u", conf.Folder+"/vimrc.preview", file.name)
	} else {
		runThis(editor, file.name)
	}

}

func editFile(file pseudofile) {
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		panic("No $EDITOR defined")
	}
	runThis(editor, file.name)
}
func renameFile(file pseudofile, newName string) {
	err := os.Rename(file.name, newName)
	if err != nil {
		panic(err)
	}
}

func copyFile(file pseudofile) {
}

func pasteFile(file pseudofile) {
}

func deleteFileWithoutTrash(s *fmState) {
	if a.confirmed {
		os.Remove(s.active.name)
	} else {
		a.cmd = s.cmd
		s.getConfirmation("deletion")
	}
}
func deleteFile(s *fmState) {
	if a.confirmed {
		moveToTrash(s.active.name)
	} else {
		a.cmd = s.cmd
		s.getConfirmation("deletion")
	}
}
func undeleteFile() {
	if len(deletedFiles) != 0 {
		home, _ := os.LookupEnv("HOME")
		t := deletedFiles
		last := t[len(t)-1]
		deletedFiles = t[:len(t)-1] // pop
		tf := home + "/.tmp/gofm_trash/" + last
		os.Rename(tf, last)
	}
}

func moveToTrash(fn string) {
	home, _ := os.LookupEnv("HOME")
	if exists, err := fileExists(home + "/.tmp/gofm_trash/"); !exists {
		if err != nil {
			log.Errorln(err)
		}
		err = os.MkdirAll(home+"/.tmp/gofm_trash/", os.ModeDir|0755)
		if err != nil {
			log.Errorln(err)
		}
	}
	os.Rename(fn, home+"/.tmp/gofm_trash/"+fn)
	deletedFiles = append(deletedFiles, fn)
}

func takeOutTrash() {
	home, _ := os.LookupEnv("HOME")
	os.RemoveAll(home + "/.tmp/gofm_trash/")
	os.MkdirAll(home+"/.tmp/gofm_trash/", os.ModeDir|0755)
}

func finalize() {
	termbox.Close()
	takeOutTrash()
	os.Exit(0)
}
