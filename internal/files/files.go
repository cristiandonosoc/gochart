// Package files provides minor support for dealing with files.
package files

import (
	"fmt"
	"io/fs"
	"os"
)

// DirExists check whether the directory exists and is a directory (not another type of file).
func DirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("stating path %q: %w", path, err)
		}

		return false, nil
	}

	return info.IsDir(), nil
}

// StatFile returns the file info of the file if it can be done.
// If the file does not exists, the returned file info will be nil.
func StatFile(path string) (fs.FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			stat = nil
		} else {
			return nil, err
		}
	}

	return stat, nil
}
