package fs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/taybart/log"
)

// Directory holds representation of a directory
type Directory struct {
	Active     int
	ActiveFile Pseudofile
	Name       string
	Path       string
	Selected   map[string]bool
	Files      []Pseudofile
}

// NewDir get directory
func NewDir(path string) (dir *Directory, err error) {
	empty, err := isEmpty(path)
	if empty {
		fakeFile := Pseudofile{Name: "empty directory...", IsDir: false, IsReal: false}
		dir = &Directory{
			Active:     0,
			ActiveFile: fakeFile,
			Path:       path,
			Files:      []Pseudofile{fakeFile},
		}
		return
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsPermission(err) {
			return
		}
		w := fmt.Sprintf("filename: %s", path)
		err = errors.Wrap(err, w)
		return
	}
	defer f.Close()

	files, err := f.Readdir(0) // Or f.Readdir(1)
	if err == io.EOF {
		w := fmt.Sprintf("filename: %s", path)
		err = errors.Wrap(err, w)
		return
	}

	files = prune(files)
	sort.Slice(files, func(i, j int) bool { return files[i].Name() < files[j].Name() })
	pfs := make([]Pseudofile, len(files))
	for i, f := range files {
		isLink := false
		broken := false
		linkloc := ""
		if f.Mode()&os.ModeSymlink != 0 {
			isLink = true
			linkloc, err = os.Readlink(path + "/" + f.Name())
			if err != nil {
				log.Errorln(err)
			}
			if _, err := os.Stat(linkloc); err != nil {
				broken = true
			}
		}
		var fullPath string
		if path != "." {
			fullPath = path + "/" + f.Name()
		} else {
			var err error
			fullPath, err = filepath.Abs(f.Name())
			if err != nil {
				log.Errorln(err)
			}
		}
		pfs[i] = Pseudofile{
			Name: f.Name(), FullPath: fullPath,
			IsDir: f.IsDir(), IsReal: true, IsLink: isLink,
			Link: Link{Broken: broken, Location: linkloc},
			F:    f,
		}
	}
	dir = &Directory{
		Path:       path,
		Active:     0,
		ActiveFile: pfs[0],
		Selected:   make(map[string]bool),
		Files:      pfs,
	}
	return
}

// isEmpty checks if directory is empty
func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	files, err := f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}

	files = prune(files)
	return len(files) == 0, err // Either not empty or error, suits both cases
}

func prune(dir []os.FileInfo) []os.FileInfo {
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
