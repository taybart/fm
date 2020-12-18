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
	Path       string
	Selected   map[string]bool
	Files      []Pseudofile
}

// NewDir get directory
func NewDir(path string) (dir *Directory, err error) {
	empty, err := isEmpty(path)
	if empty || err != nil {
		if err != nil {
			log.Error(err)
		}
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
			linkloc, err = os.Readlink(filepath.Join(path, f.Name()))
			if err != nil {
				log.Errorln(err)
			}
			loc := linkloc
			if linkloc[0] != '/' {
				loc = filepath.Join(path, loc)
			}
			if _, err := os.Stat(loc); err != nil {
				broken = true
			}
		}
		fullPath := filepath.Join(path, f.Name())
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

// NewParentDir get parent dir and select itself
func NewParentDir(cd string) (dir *Directory, err error) {
	parent := GetParentPath(cd)
	dir, err = NewDir(parent)
	if err != nil {
		return
	}
	pn, err := getParentName(cd)
	if err != nil {
		return
	}
	dir.SelectFileByName(pn)
	return
}

// SelectFile change active file
func (d *Directory) SelectFile(direction int) bool {
	if len(d.Files) > 0 {
		index := d.Active
		index += direction
		if index >= len(d.Files) {
			index = len(d.Files) - 1
		}
		if index <= 0 {
			index = 0
		}
		d.Active = index
		d.ActiveFile = d.Files[d.Active]
		return true
	}
	return false
}

// SelectFileByName change active file
func (d *Directory) SelectFileByName(name string) {
	a := 0
	var af Pseudofile
	for i, f := range d.Files {
		if f.Name == name {
			a = i
			af = f
			break
		}
	}
	d.Active = a
	d.ActiveFile = af
}

// isEmpty checks if directory is empty
func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	files, err := f.Readdir(0)
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
