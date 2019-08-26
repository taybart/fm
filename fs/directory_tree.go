package fs

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/taybart/fm/config"
	"github.com/taybart/log"
)

// Tree holds multiple directories
type Tree map[string]*Directory

var conf *config.Config

// Init initalize
func Init(c *config.Config, cd string) (dt *Tree, err error) {
	conf = c

	dt = &Tree{}
	(*dt)[cd], err = NewDir(cd)
	if err != nil {
		return
	}
	parent := GetParentPath(cd)
	(*dt)[parent], err = NewParentDir(cd)
	if err != nil {
		return
	}

	child := (*dt)[cd].ActiveFile.FullPath
	if ok, staterr := IsDir(child); ok && staterr == nil {
		(*dt)[child], err = NewDir(child)
		if err != nil {
			return
		}
	}
	return
}

// SelectFile change active file
func (dt *Tree) SelectFile(direction int, cd string) (err error) {
	if selected := (*dt)[cd].SelectFile(direction); selected {
		err := dt.ReadChild(cd)
		if err != nil {
			log.Error(err)
		}
	}
	return
}

// ChangeDirectory cd
func (dt *Tree) ChangeDirectory(dirname string) (err error) {
	if ok, staterr := IsDir(dirname); !ok || staterr != nil {
		err = errors.New("Not a directory")
		return
	}
	parent := GetParentPath(dirname)
	// log.Debug("parent", parent)

	if _, exists := (*dt)[parent]; !exists && parent != "" {
		(*dt)[parent], err = NewParentDir(dirname)
		if err != nil {
			log.Error(err)
			return
		}
	}

	log.Verbose("Changing to", dirname)
	if _, exists := (*dt)[dirname]; !exists {
		log.Debug(dirname, "Does not exist, adding dir")
		(*dt)[dirname], err = NewDir(dirname)
		if err != nil {
			log.Error(err)
			return
		}
	}

	child := (*dt)[dirname].ActiveFile.FullPath
	if ok, staterr := IsDir(child); ok || staterr != nil {
		if _, exists := (*dt)[child]; !exists {
			log.Debug(child, "Does not exist, adding child")
			(*dt)[child], err = NewDir(child)
			if err != nil {
				log.Error(err)
				return
			}
		}
	}
	return
}

// Update sync directory
func (dt *Tree) Update(cd string) (err error) {
	aName := (*dt)[cd].ActiveFile.Name
	dir, err := NewDir(cd)
	if err != nil {
		return
	}

	dir.SelectFileByName(aName)
	(*dt)[cd] = dir
	return
}

// ReadChild change directory
func (dt *Tree) ReadChild(cd string) (err error) {
	child := (*dt)[cd].ActiveFile.FullPath

	if _, exists := (*dt)[child]; exists {
		return
	}

	if ok, staterr := IsDir(child); ok && staterr == nil {
		var ch *Directory
		ch, err = NewDir(child)
		if err != nil {
			return
		}
		(*dt)[child] = ch
	}
	return nil
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
