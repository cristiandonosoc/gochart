// Package files provides minor support for dealing with files.
package files

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ToUnixPath standardizes the path to be Unix-like. This is useful for making paths work
// uniformly between Windows and Linux.
func ToUnixPath(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

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
func StatFile(path string) (fs.FileInfo, bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		} else {
			return nil, false, err
		}
	}

	return stat, true, nil
}

// CopyFile
func CopyFile(src, dst string) error {
	return CopyFileAdvanced(src, dst, nil)
}

type CopyFileAdvancedOptions struct {
	// DstCreateDir determines whether to try to create the owning directory of the dst file.
	DstCreateDir         bool
	DstCreateDirFileMode fs.FileMode

	// Sync ensures any buffered data is sent immediatelly.
	Sync bool
}

var (
	GDefaultCopyFileAdvancedOptions = CopyFileAdvancedOptions{
		DstCreateDir:         false,
		DstCreateDirFileMode: 0755,
		Sync:                 false,
	}
)

func CopyFileAdvanced(src, dst string, options *CopyFileAdvancedOptions) error {
	if options == nil {
		options = &GDefaultCopyFileAdvancedOptions
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening %q: %w", src, err)
	}
	defer srcFile.Close()

	if options.DstCreateDir {
		dir := filepath.Dir(dst)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdirall %q: %w", dir, err)
		}
	}

	// Create (or truncate) the destination file.
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("opening %q: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copying data from %q to %q: %w", src, dst, err)
	}

	if options.Sync {
		if err := dstFile.Sync(); err != nil {
			return fmt.Errorf("calling sync on %q: %w", dst, err)
		}
	}

	return nil
}

// CopyDirRecursive copies all the content of a directory into another path.
func CopyDirRecursive(from, to string) error {
	// TODO(cdc): Make this use errgroup.
	from = filepath.Clean(from)

	var files []string
	err := filepath.WalkDir(from, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Remove the prefix from the path.
		path = ToUnixPath(strings.TrimPrefix(path, from))
		path = strings.TrimPrefix(path, "/")
		files = append(files, path)
		return nil
	})

	if err != nil {
		return fmt.Errorf("walking %q: %w", from, err)
	}

	to = filepath.Clean(to)

	for _, file := range files {
		src := filepath.Join(from, file)
		dst := filepath.Join(to, file)

		options := CopyFileAdvancedOptions{
			DstCreateDir:         true,
		}
		if err := CopyFileAdvanced(src, dst, &options); err != nil {
			return fmt.Errorf("copying %q -> %q: %w", src, dst, err)
		}
	}

	return nil
}
