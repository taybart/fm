package fs

import (
	"github.com/pkg/errors"
	"github.com/taybart/fm/config"
	"github.com/taybart/log"
	"strings"
)

var conf *config.Config

// Init initalize
func Init(c *config.Config, cd string) (dt Tree, err error) {
	conf = c

	dt = Tree{}
	dt[cd], err = NewDir(cd)
	if err != nil {
		return
	}
	parent := GetParentPath(cd)
	dt[parent], err = NewParentDir(cd)
	if err != nil {
		return
	}

	child := cd + "/" + dt[cd].ActiveFile.Name

	if ok, staterr := IsDir(child); ok && staterr == nil {
		dt[child], err = NewDir(child)
		if err != nil {
			return
		}
	}
	return
}

// CD change directory
func (dt *Tree) ReadChild(cd string) (err error) {
	child := (*dt)[cd].ActiveFile.FullPath
	if ok, staterr := IsDir(child); ok && staterr == nil {
		var ch *Directory
		ch, err = NewDir(child)
		if err != nil {
			return
		}
		log.Info(*ch)
		(*dt)[child] = ch
	}
	return nil
}

// Tree holds multiple directories
type Tree map[string]*Directory

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
	a := 0
	var af Pseudofile
	for i, f := range dir.Files {
		if f.Name == pn {
			a = i
			af = f
			break
		}
	}
	dir.Active = a
	dir.ActiveFile = af
	return
}

// GetParentPath dwisott
func GetParentPath(cd string) string {
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
	return dirs[len(dirs)-1], nil
}
