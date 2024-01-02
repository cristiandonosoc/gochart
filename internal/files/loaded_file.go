package files

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
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

	lf, err := LoadFileFromData(data)
	if err != nil {
		return nil, fmt.Errorf("creating load file: %w", err)
	}

	lf.Path = path
	return lf, nil
}

func LoadFileFromData(data []byte) (*LoadedFile, error) {
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

// LoadedFilePosition represents a character position within a loaded file.
type LoadedFilePosition struct {
	File *LoadedFile
	Line int
	Char int
}

// Render prints the position pointed, with an possible message associated with it.
// The printing is "clang error style".
// |contextSize| is how many lines around the position to print (in each direction).
func (lfp *LoadedFilePosition) Render(header, msg, footer string, contextSize int) string {
	if lfp.File == nil {
		return "<no loaded file>"
	}

	var output []string

	// Start with an empty line.
	output = append(output, "")
	if header != "" {
		output = append(output, header)
	}

	// Show a couple lines of context.
	begin := lfp.Line - contextSize
	end := lfp.Line + contextSize

	// Just in case the printing didn't get to print the msg.
	printedMsg := false

	for i := begin; i <= end; i++ {
		// NOTE: Line as 1-based.
		arrayIndex := i - 1
		if arrayIndex < 0 {
			continue
		}

		if arrayIndex >= len(lfp.File.Lines) {
			break
		}

		line := lfp.File.Lines[arrayIndex]

		// We print the line.
		output = append(output, fmt.Sprintf("%4d: %s", i, line))

		// if this is the error line, we also add a caret on the location and the actual error.
		if i == lfp.Line {
			// We need to create the padding taking into account tabs! Fucking tabs!
			padding := make([]rune, 0, lfp.Char)
			for _, runeValue := range line {
				if len(padding) >= lfp.Char-1 {
					break
				}

				// if it is a tab, we want to add it as is into the padding.
				// Otherwise we just add a space (we assume all other characters are 1 char wide.
				if runeValue == '\t' {
					padding = append(padding, '\t')
					continue
				}

				padding = append(padding, ' ')
			}

			// We merge the padding with the known padding we added to account for line numbers.
			output = append(output, fmt.Sprintf("%*s%s^---> %s", 6, " ", string(padding), msg))
			printedMsg = true
		}
	}

	if !printedMsg && msg != "" {
		output = append(output, msg)
	}

	if footer != "" {
		output = append(output, "", footer)
	}

	// Trim the spaces.
	for i := 0; i < len(output); i++ {
		output[i] = strings.TrimRight(output[i], " \r\n\t")
	}

	return strings.Join(output, "\n")
}
