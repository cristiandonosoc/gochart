package files

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Main API ----------------------------------------------------------------------------------------

// LoadFileFromPath attempts to load a file from a path and will store it in the global cache.
func LoadFileFromPath(path string) (*LoadedFile, error) {
	key, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("abs %q: %w", path, err)
	}
	return GlobalFileCache().LoadFromPath(key, path, false)
}

// LoadFileFromPathWithKey is a more advanced way of loading files that permit to insert it in an
// specific key, rather than using the abs path, as it is normally done.
// |overwrite| refers to whether we allow people to overwrite keys or not.
func LoadFileFromPathWithKey(key, path string, overwrite bool) (*LoadedFile, error) {
	return GlobalFileCache().LoadFromPath(key, path, overwrite)
}

// QueryKey checks to see if the key is in the global cache, and if so, returns the loaded file.
// For files loaded as path, the key would the absolute path as returned by |filepath.Abs|.
func QueryKey(key string) (bool, *LoadedFile) {
	return GlobalFileCache().QueryKey(key)
}

// NewFromData creates a in-memory loaded file from the given data and stores it in the cache with
// the given key. From that moment on, it works similarly to a loaded file read from a file.
// |overwrite| refers to whether we allow people to overwrite keys or not.
func NewFromData(key string, data []byte, overwrite bool) (*LoadedFile, error) {
	return GlobalFileCache().NewFromData(key, data, overwrite)
}

// Cache Implementation ----------------------------------------------------------------------------

var once sync.Once
var gFileCache *fileCache

// FileCache represents a view to files loaded in memory.
type fileCache struct {
	files map[string]*LoadedFile
	// DisallowKeyCollision determines whether the cache will look for key collisions.
	// Normally we disallow this for tests.
	DisallowKeyCollision bool
}

func GlobalFileCache() *fileCache {
	once.Do(func() {
		gFileCache = &fileCache{
			files:                map[string]*LoadedFile{},
			DisallowKeyCollision: false,
		}
	})

	return gFileCache
}

// QueryKey checks the cache to see if that key has already been loaded.
func (fc *fileCache) QueryKey(key string) (bool, *LoadedFile) {
	if file, ok := fc.files[key]; ok {
		return true, file
	}

	return false, nil
}

// LoadFromPath creates a new loaded file from a path.
// The key of the file will be the absolute path of the file.
func (fc *fileCache) LoadFromPath(key, path string, overwrite bool) (*LoadedFile, error) {
	// Check if the file is already read.
	if file, ok := fc.files[key]; ok {
		return file, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", path, err)
	}

	return fc.NewFromData(key, data, overwrite)
}

// NewFromData creates a new loadedFile with the provided key and content.
// The key must not be in use already.
// This is normally used for in-memory files, usually for testing purposes.
func (fc *fileCache) NewFromData(key string, data []byte, overwrite bool) (*LoadedFile, error) {
	// We should not have the key already.
	if !fc.DisallowKeyCollision && !overwrite {
		if _, ok := fc.files[key]; ok {
		}
		return nil, fmt.Errorf("key %q is already in use", key)
	}

	file := &LoadedFile{
		Key:  key,
		Data: data,
	}
	fc.files[key] = file

	return file, nil
}
