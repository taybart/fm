package main

import (
	"github.com/taybart/log"
	"io"
	// "io/ioutil"
	"os"
	"strings"
)

type dir struct {
	active int
	name   string
}
type directoryTree map[string]*dir

func getParentPath(cd string) string {
	dirs := strings.Split(cd, "/")
	tmp := dirs[:len(dirs)-1]

	return strings.Join(tmp, "/")
}

func getParentName(cd string) string {
	dirs := strings.Split(cd, "/")
	return dirs[len(dirs)-1]
}

func (dt directoryTree) newDirForParent(cd string) *dir {
	pp := getParentPath(cd)
	p := getParentName(cd)
	fs := readDir(pp)

	a := 0
	for _, f := range fs {
		if f.Name() == p {
			break
		}
		a++
	}
	return &dir{active: a}
}

func pruneDirs(dir []os.FileInfo) []os.FileInfo {
	pruned := []os.FileInfo{}
	for _, f := range dir {
		if rune(f.Name()[0]) == '.' {
			if conf.ShowHidden {
				pruned = append(pruned, f)
			}
		} else {
			pruned = append(pruned, f)
		}
	}
	return pruned
}

func readDir(name string) []os.FileInfo {
	f, err := os.Open(name)
	if err != nil {
		// return false, err
		log.Errorln(err)
	}
	defer f.Close()

	files, err := f.Readdir(0) // Or f.Readdir(1)
	if err == io.EOF {
		// return files, nil
		log.Errorln(err)
	}
	files = pruneDirs(files)
	return files
}
func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func pwd() string {
	cd, err := os.Getwd()
	if err != nil {
		log.Errorln(err)
	}
	return cd
}
