package files

import (
	"bufio"
	"bytes"
	"fmt"
	"io/fs"
)

type LoadedFile struct {
	Key   string
	Data  []byte
	lines []string

	FromFile bool
	Stat     fs.FileInfo
}

// LoadedFilePosition represents a single position (character) within a loaded file.
type LoadedFilePosition struct {
	File *LoadedFile
	Line int
	Char int
}

func NewLoadedFilePosition(lf *LoadedFile, line, char int) *LoadedFilePosition {
	return &LoadedFilePosition{
		File: lf,
		Line: line,
		Char: char,
	}
}

// Path returns the Key as a path if the file was loaded from file rather than a buffer.
// Returns empty otherwise.
func (lf *LoadedFile) Path() string {
	if lf.FromFile {
		return lf.Key
	}

	return ""
}

// Lines lazily parses the content of the file into lines.
func (lf *LoadedFile) Lines() ([]string, error) {
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
