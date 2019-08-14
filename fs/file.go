package fs

import (
	"io"
	"os"
)

// Pseudofile file info
type Pseudofile struct {
	Name     string
	FullPath string
	IsDir    bool
	IsReal   bool
	IsLink   bool
	Link     Link
	F        os.FileInfo
}

// Link symlink
type Link struct {
	Location string
	Broken   bool
}

// IsDir check if file is real
func IsDir(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

// FileExists check if file is real
func FileExists(fn string) (bool, error) {
	if _, err := os.Stat(fn); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// CountChildren get amount of children
func CountChildren(pf Pseudofile) int {
	f, err := os.Open(pf.FullPath)
	if err != nil {
		return -1
	}
	defer f.Close()

	files, err := f.Readdir(0) // Or f.Readdir(1)
	if err == io.EOF {
		return 0
	}
	files = prune(files)
	return len(files)
}
