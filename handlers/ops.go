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

var deletedFiles []fs.Pseudofile

var selectedFiles []fs.Pseudofile

func paste(dt *fs.Tree, cd string) error {
	lastClipboard := "cut"
	for _, c := range runCommands {
		if c == "cut" || c == "yank" {
			lastClipboard = c
		}
	}
	switch lastClipboard {
	case "cut":
		if len(selectedFiles) == 0 {
			activeFile.Move(dt, cd)
		} else {
			for _, f := range selectedFiles {
				f.Move(dt, cd)
			}
		}
	case "yank":
		if len(selectedFiles) == 0 {
			activeFile.Copy(dt, cd)
		} else {
			for _, f := range selectedFiles {
				f.Copy(dt, cd)
			}
		}
	}
	selectedFiles = []fs.Pseudofile{}
	return nil
}

func yank(dt *fs.Tree, cd string) error {
	if len(selectedFiles) == 0 {
		activeFile = (*dt)[cd].ActiveFile
	}
	return nil
}

func deletef(dt *fs.Tree, cd string) error {
	if len(selectedFiles) == 0 {
		ans := prompt(fmt.Sprintf("Delete %s? [Y/n]", (*dt)[cd].ActiveFile.Name))
		if ans != "n" {
			moveToTrash((*dt)[cd].ActiveFile)
			err := dt.SelectFile(-1, cd)
			if err != nil {
				log.Error(err)
			}
		}
	} else {
		files := "[ "
		for i, f := range selectedFiles {
			files += f.Name
			if i < len(selectedFiles)-1 {
				files += ", "
			}
		}
		files += " ]"

		ans := prompt(fmt.Sprintf("Delete %v? [Y/n]", files))
		if ans != "n" && ans != "N" {
			for _, f := range selectedFiles {
				moveToTrash(f)
			}
		}
	}
	selectedFiles = []fs.Pseudofile{}
	return dt.Update(cd)
}

func inspect(dt *fs.Tree, cd string) error {
	file := (*dt)[cd].ActiveFile
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		editor = "vi"
	}
	if editor == "vim" || editor == "nvim" {
		return runThis(editor, "-u", conf.Folder+"/vimrc", "-M", file.Name)
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

func fuzzyCD(dt *fs.Tree, cd string) (string, error) {
	dir := (*dt)[cd]
	filtered := fzf(func(in io.WriteCloser) {
		for _, f := range dir.Files {
			fmt.Fprintln(in, f.Name)
		}
	})
	selected := filtered[0]
	fp := path.Join(cd, selected)
	isdir, err := fs.IsDir(fp)
	if err != nil {
		log.Error(err)
	}
	if isdir {
		(*dt)[cd].SelectFileByName(selected)
		err := dt.Update(cd)
		if err != nil {
			log.Error(err)
		}
		err = dt.ChangeDirectory(fp)
		if err != nil {
			log.Error(err)
		}
		cd = fp
	} else {
		(*dt)[cd].SelectFileByName(selected)
		err := dt.Update(cd)
		if err != nil {
			log.Error(err)
		}
	}
	return cd, nil
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

func digOutOfTrash() error {
	a := deletedFiles
	f, a := a[len(a)-1], a[:len(a)-1]
	deletedFiles = a
	home, _ := os.LookupEnv("HOME")
	trash := path.Join(home, "/.tmp/fm_trash/")
	if exists, err := fs.FileExists(trash); !exists {
		if err != nil {
			log.Errorln(err)
		}
		err = os.MkdirAll(trash, os.ModeDir|0755)
		if err != nil {
			log.Errorln(err)
		}
	}
	trashName := strings.ReplaceAll(f.FullPath, "/", "_")
	err := os.Rename(path.Join(trash, trashName), f.FullPath)
	if err != nil {
		log.Errorln(err)
	}
	return err
}

func moveToTrash(f fs.Pseudofile) {
	home, _ := os.LookupEnv("HOME")
	trash := path.Join(home, "/.tmp/fm_trash/")
	if exists, err := fs.FileExists(trash); !exists {
		if err != nil {
			log.Errorln(err)
		}
		err = os.MkdirAll(trash, os.ModeDir|0755)
		if err != nil {
			log.Errorln(err)
		}
	}
	trashName := strings.ReplaceAll(f.FullPath, "/", "_")
	err := os.Rename(f.FullPath, path.Join(trash, trashName))
	if err != nil {
		log.Errorln(err)
	}
	deletedFiles = append(deletedFiles, f)
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
