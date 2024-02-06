package files

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var once sync.Once
var gFileCache *fileCache

// FileCache represents a view to files loaded in memory.
type fileCache struct {
	files map[string]*loadedFile
}

func GlobalFileCache() *fileCache {
	once.Do(func() {
		gFileCache = &fileCache{}
	})

	return gFileCache
}

type loadedFile struct {
	Key   string
	Data  []byte
	lines []string

	FromFile bool
}

// Path returns the Key as a path if the file was loaded from file rather than a buffer.
// Returns empty otherwise.
func (lf *loadedFile) Path() string {
	if lf.FromFile {
		return lf.Key
	}

	return ""
}

// Lines lazily parses the content of the file into lines.
func (lf *loadedFile) Lines() ([]string, error) {
	// Check if the lines have already been loaded.
	if lf.lines != nil {
		return lf.lines, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(lf.Data))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file line by line: %w", err)
	}

	lf.lines = lines
	return lf.lines, nil
}

// QueryKey checks the cache to see if that key has already been loaded.
func (fc *fileCache) QueryKey(key string) (bool, *loadedFile) {
	if file, ok := fc.files[key]; ok {
		return true, file
	}

	return false, nil
}

// LoadFromPath creates a new loaded file from a path.
// The key of the file will be the absolute path of the file.
func (fc *fileCache) LoadFromPath(path string) (*loadedFile, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("abs %q: %w", path, err)
	}

	// Check if the file is already read.
	if file, ok := fc.files[abs]; ok {
		return file, nil
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", abs, err)
	}

	return fc.NewFromData(abs, data)
}

// NewFromData creates a new loadedFile with the provided key and content.
// The key must not be in use already.
// This is normally used for in-memory files, usually for testing purposes.
func (fc *fileCache) NewFromData(key string, data []byte) (*loadedFile, error) {
	// We should not have the key already.
	if _, ok := fc.files[key]; ok {
		return nil, fmt.Errorf("key %q is already in use", key)
	}

	file := &loadedFile{
		Key:  key,
		Data: data,
	}
	fc.files[key] = file

	return file, nil
}
