package main

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/taybart/log"
	"io"
	"sort"
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
	if len(tmp) == 1 && cd != "/" {
		return "/"
	}

	return strings.Join(tmp, "/")
}

func getParentName(cd string) (string, error) {
	if cd == "/" {
		return "", errors.New("Root directory")
	}
	dirs := strings.Split(cd, "/")
	/* if len(dirs) == 1 {
		return "", errors.New("")
	} */
	return dirs[len(dirs)-1], nil
}

func (dt directoryTree) newDirForParent(cd string) *dir {
	pp := getParentPath(cd)
	p, err := getParentName(cd)
	if err != nil {
		return &dir{active: 0}
		// panic(err) // @TODO: tmp
	}
	fs, err := readDir(pp)
	if err != nil {
		panic(err) // @TODO: tmp
	}

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

func readDir(name string) ([]os.FileInfo, error) {
	f, err := os.Open(name)
	if err != nil {
		if os.IsPermission(err) {
			return nil, nil
		}
		w := fmt.Sprintf("filename: %s", name)
		return nil, errors.Wrap(err, w)
	}
	defer f.Close()

	files, err := f.Readdir(0) // Or f.Readdir(1)
	if err == io.EOF {
		w := fmt.Sprintf("filename: %s", name)
		return nil, errors.Wrap(err, w)
	}
	files = pruneDirs(files)
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	return files, nil
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
