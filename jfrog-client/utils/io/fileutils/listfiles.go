package fileutils

import (
	"errors"
	"os"
	"path/filepath"
	"sort"
)

var SkipDir = errors.New("skip this directory")

type WalkFunc func(path string, info os.FileInfo, err error) error
type Stat func(path string) (info os.FileInfo, err error)

var stat = os.Stat
var lStat = os.Lstat

func walk(path string, info os.FileInfo, walkFn WalkFunc, visited map[string]bool, walkIntoDirSymlink bool) error {
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		realPath = path
	}
	isRealPathDir, err := IsDir(realPath)
	if err != nil {
		return err
	}
	if isRealPathDir {
		visited[realPath] = true
	}
	err = walkFn(path, info, nil)
	if err != nil {
		if info.IsDir() && err == SkipDir {
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return nil
	}

	names, err := readDirNames(path)
	if err != nil {
		return walkFn(path, info, err)
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		realPath, err = filepath.EvalSymlinks(filename)
		if err != nil {
			realPath = filename
		}

		if walkIntoDirSymlink && visited[realPath] {
			continue
		}
		var fileHandler Stat
		if walkIntoDirSymlink {
			fileHandler = stat
		} else {
			fileHandler = lStat
		}
		fileInfo, err := fileHandler(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != SkipDir {
				return err
			}
		} else {
			err = walk(filename, fileInfo, walkFn, visited, walkIntoDirSymlink)
			if err != nil {
				if !fileInfo.IsDir() || err != SkipDir {
					return err
				}
			}
		}
	}
	return nil
}

// The same as filepath.Walk the only difference is that we can walk into symlink.
// Avoiding infinite loops by saving the real paths we already visited.
func Walk(root string, walkFn WalkFunc, walkIntoDirSymlink bool) error {
	info, err := stat(root)
	visited := make(map[string]bool)
	if err != nil {
		return walkFn(root, nil, err)
	}
	return walk(root, info, walkFn, visited, walkIntoDirSymlink)
}

// readDirNames reads the directory named by dirname and returns
// a sorted list of directory entries.
// The same as path/filepath readDirNames function
func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}