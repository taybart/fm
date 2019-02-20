package main

import (
	"fmt"
	"github.com/nsf/termbox-go"
	"github.com/pkg/errors"
	"github.com/taybart/log"
	"io"
	"os"
	"os/exec"
	"strings"
)

func (s *fmState) getConfirmation(action string) {
	s.cmd = "Confirm " + action + " [Yy]: "
	s.mode = confirm
}

func newShell() {
	shell, exists := os.LookupEnv("SHELL")
	if !exists {
		shell = "sh"
	}
	runThis(shell)
}

func (s *fmState) changeDirectory(file string) {
	dn := s.cd + "/" + file
	if _, ok := s.dt[dn]; !ok {
		s.dt[dn] = &dir{active: 0}
	}
	os.Chdir(dn)
}

func inspectFile(file pseudofile) {
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		editor = "vi"
	}
	if editor == "vim" || editor == "nvim" {
		runThis(editor, "-u", conf.Folder+"/vimrc.preview", "-M", file.name)
	} else {
		runThis(editor, file.name)
	}

}

func editFile(file pseudofile) {
	editor, exists := os.LookupEnv("EDITOR")
	if !exists {
		editor = "vi"
	}
	runThis(editor, file.name)
}
func renameFile(file pseudofile, newName string) {
	err := os.Rename(file.name, newName)
	if err != nil {
		log.Errorln(err)
	}
}

func copyFile(s *fmState) {
	if len(s.selectedFiles) == 0 {
		s.copySource = s.active
	}
	s.copyBufReady = true
}

func moveFile(s *fmState) error {
	if !s.copyBufReady {
		return errors.New("No file in copy buffer")
	}
	if len(s.selectedFiles) == 0 {
		err := pasteSingleFile(s, s.copySource)
		if err != nil {
			return err
		}
		os.Remove(s.copySource.fullPath)
	}
	s.selectedFiles = map[string]pseudofile{} // clear selected files
	return nil
}
func pasteFile(s *fmState) error {
	if !s.copyBufReady {
		return errors.New("No file in copy buffer")
	}
	if len(s.selectedFiles) == 0 {
		return pasteSingleFile(s, s.copySource)
	}
	for _, f := range s.selectedFiles {
		pasteSingleFile(s, f)
	}

	s.selectedFiles = map[string]pseudofile{} // clear selected files
	return nil
}

func pasteSingleFile(s *fmState, file pseudofile) error {
	destName := s.cd + "/" + file.name
	if _, err := os.Stat(destName); err == nil {
		ext := strings.Split(file.name, ".")
		if len(ext) > 1 {
			destName = s.cd + "/" + ext[0] + "_copy." + ext[1]
		} else {
			destName += "_copy"
		}
	}
	if file.isDir {
		runThis("cp", "-a", file.fullPath, destName)
	} else {
		buf := make([]byte, file.f.Size())
		source, err := os.Open(file.fullPath)
		destination, err := os.Create(destName)
		if err != nil {
			return err
		}
		for {
			n, err := source.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			if n == 0 {
				break
			}

			if _, err := destination.Write(buf[:n]); err != nil {
				return err
			}
		}
	}
	return nil
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

func fzf(input func(in io.WriteCloser)) []string {
	termbox.Close()
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
	setupDisplay()
	return strings.Split(string(result), "\n")
}

func fuzzyFind(s *fmState) {
	filtered := fzf(func(in io.WriteCloser) {
		for _, f := range s.dir {
			fmt.Fprintln(in, f.name)
		}
	})

	for i, f := range s.dir {
		if filtered[0] == f.name {
			s.dt[s.cd].active = i
			if s.dir[i].isDir {
				dn := s.cd + "/" + s.dir[i].name
				if s.cd == "/" {
					dn = s.cd + s.dir[i].name
				}
				if _, ok := s.dt[dn]; !ok {
					s.dt[dn] = &dir{active: 0}
				}
				navtree = append(navtree, s.cd)
				os.Chdir(dn)
			}
		}
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

func runThis(toRun string, args ...string) error {
	termbox.Close()
	cmd := exec.Command(toRun, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Errorln(err)
		panic(err)
	}
	setupDisplay()
	return nil
}
