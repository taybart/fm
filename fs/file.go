package fs

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/taybart/log"
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

// Move file to new location
func (f *Pseudofile) Move(dt *Tree, directory string) {
	// Move
	log.Verbose("Moving", f.FullPath, path.Join(directory, f.Name))
	os.Rename(f.FullPath, path.Join(directory, f.Name))
	parent := GetParentPath(f.FullPath)
	log.Verbose("base", parent)
	for i, fi := range (*dt)[parent].Files {
		if f.FullPath == fi.FullPath {
			dir := (*dt)[parent]
			dir.Files = append(dir.Files[:i], dir.Files[i:]...)
		}
	}
	dt.Update(parent)
	dt.Update(directory)
}

// Copy file to new location
func (f *Pseudofile) Copy(dt *Tree, directory string) {
	if f.IsDir {
		copyDir(f.FullPath, path.Join(directory, f.Name))
	} else {
		name := path.Join(directory, f.Name)
		if exists, err := FileExists(name); exists {
			if err != nil {
				log.Error(err)
			}
			ext := strings.Split(f.Name, ".")
			if len(ext) > 1 {
				name = directory + "/" + ext[0] + "_copy." + ext[1]
			} else {
				name += "_copy"
			}
		}
		log.Info(f.FullPath, name)
		copyFile(f.FullPath, name)
	}
	parent := GetParentPath(f.FullPath)
	dt.Update(parent)
	dt.Update(directory)
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	var err error
	var srcfd *os.File
	var dstfd *os.File
	var srcinfo os.FileInfo

	if srcfd, err = os.Open(src); err != nil {
		return err
	}
	defer srcfd.Close()

	if dstfd, err = os.Create(dst); err != nil {
		return err
	}
	defer dstfd.Close()

	if _, err = io.Copy(dstfd, srcfd); err != nil {
		return err
	}
	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}
	return os.Chmod(dst, srcinfo.Mode())
}

// copyDir copies a whole directory recursively
func copyDir(src string, dst string) error {
	var err error
	var fds []os.FileInfo
	var srcinfo os.FileInfo

	if srcinfo, err = os.Stat(src); err != nil {
		return err
	}

	if err = os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	if fds, err = ioutil.ReadDir(src); err != nil {
		return err
	}
	for _, fd := range fds {
		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			if err = copyDir(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		} else {
			if err = copyFile(srcfp, dstfp); err != nil {
				fmt.Println(err)
			}
		}
	}
	return nil
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
