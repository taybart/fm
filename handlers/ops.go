package handlers

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/taybart/fm/display"
	"github.com/taybart/fm/fs"
	"github.com/taybart/log"
)

var activeFile fs.Pseudofile

func paste(dt *fs.Tree, cd string) error {
	lastClipboard := "cut"
	for _, c := range runCommands {
		if c == "cut" || c == "yank" {
			lastClipboard = c
		}
	}
	switch lastClipboard {
	case "cut":
		activeFile.Move(dt, cd)
	case "yank":
		activeFile.Copy(dt, cd)
	}
	return nil
}

func yank(dt *fs.Tree, cd string) error {
	activeFile = (*dt)[cd].ActiveFile
	return nil
}

func deletef(dt *fs.Tree, cd string) error {
	ans := prompt(fmt.Sprintf("Delete %s? [Y/n]", (*dt)[cd].ActiveFile.Name))
	if ans == "n" {
		return nil
	}
	moveToTrash((*dt)[cd].ActiveFile.FullPath)
	return dt.Update(cd)
}

func inspect(dt *fs.Tree, cd string) error {
	file := (*dt)[cd].ActiveFile
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		editor = "vi"
	}
	if editor == "vim" || editor == "nvim" {
		return runThis(editor, "-u", conf.Folder+"/vimrc.preview", "-M", file.Name)
	}
	return runThis(editor, file.Name)
}

func edit(dt *fs.Tree, cd string) error {
	file := (*dt)[cd].ActiveFile
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		editor = "vi"
	}
	return runThis(editor, file.Name)
}

func fuzzyFind(dir *fs.Directory) error {
	filtered := fzf(func(in io.WriteCloser) {
		for _, f := range dir.Files {
			fmt.Fprintln(in, f.Name)
		}
	})
	log.Info(filtered)
	return nil
}

func fzf(input func(in io.WriteCloser)) []string {
	display.Close()
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "sh"
	}
	cmd := exec.Command(shell, "-c", "fzf", "-m")
	cmd.Stderr = os.Stderr
	in, _ := cmd.StdinPipe()
	go func() {
		input(in)
		in.Close()
	}()
	result, _ := cmd.Output()
	display.Init(conf)
	return strings.Split(string(result), "\n")
}

// TODO add deletedFiles
func moveToTrash(fn string) {
	home, _ := os.LookupEnv("HOME")
	if exists, err := fs.FileExists(home + "/.tmp/fm_trash/"); !exists {
		if err != nil {
			log.Errorln(err)
		}
		err = os.MkdirAll(home+"/.tmp/fm_trash/", os.ModeDir|0755)
		if err != nil {
			log.Errorln(err)
		}
	}
	err := os.Rename(fn, home+"/.tmp/fm_trash/"+path.Base(fn))
	if err != nil {
		log.Errorln(err)
	}
	// deletedFiles = append(deletedFiles, fn)
}

func takeOutTrash() {
	home, _ := os.LookupEnv("HOME")
	os.RemoveAll(home + "/.tmp/fm_trash/")
	err := os.MkdirAll(home+"/.tmp/fm_trash/", os.ModeDir|0755)
	if err != nil {
		log.Errorln(err)
	}
}

func runThis(toRun string, args ...string) error {
	display.Close()
	cmd := exec.Command(toRun, args...)

	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	log.Verbose("Executing command", toRun)
	err := cmd.Run()
	if err != nil {
		log.Fatalln(err)
	}

	display.Init(conf)
	log.Verbose("Done")
	return nil
}
