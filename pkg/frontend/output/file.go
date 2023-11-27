package output

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
)

// LoadedFile represents a single gochart source file loaded in memory.
type LoadedFile struct {
	Path  string
	Data  []byte
	Lines []string
}

// LoadFileFromPath reads the contents of a file from path and caches it in convenient ways so that
// the rest of the compiler can benefit from it.
func LoadFileFromPath(path string) (*LoadedFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file at %q: %w", path, err)
	}

	lf, err := LoadFile(data)
	if err != nil {
		return nil, fmt.Errorf("creating load file: %w", err)
	}

	lf.Path = path
	return lf, nil
}

func LoadFile(data []byte) (*LoadedFile, error) {
	scanner := bufio.NewScanner(bytes.NewReader(data))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning file line by line: %w", err)
	}

	return &LoadedFile{
		Data:  data,
		Lines: lines,
	}, nil
}
