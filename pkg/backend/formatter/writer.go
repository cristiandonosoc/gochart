package formatter

import (
	"fmt"
	"io"
	"strings"
)

type IndentedWriter struct {
	indent      int
	lines       [][]rune
	currentLine int
}

func NewIndentedWriter() *IndentedWriter {
	iw := &IndentedWriter{}
	// Initialize the first line.
	iw.lines = append(iw.lines, []rune{})
	return iw
}

var _ io.Writer = (*IndentedWriter)(nil)

func (iw *IndentedWriter) MustIncreaseIndent() {
	iw.indent++
}

func (iw *IndentedWriter) MustDecreateIndent() {
	iw.indent--
	if iw.indent < 0 {
		panic("Wrong scoped indent!")
	}
}

func (iw *IndentedWriter) Write(p []byte) (int, error) {
	data := strings.TrimSpace(string(p))
	for _, r := range data {
		if r == '\n' {
			iw.handleNewLine()
			continue
		}

		iw.appendRune(r)
	}

	return len(p), nil
}

func (iw *IndentedWriter) String() string {
	lines := make([]string, 0, len(iw.lines))

	for _, runeLine := range iw.lines {
		lines = append(lines, string(runeLine))
	}

	return strings.Join(lines, "\n")
}

func (iw *IndentedWriter) handleNewLine() {
	// Add a new line
	iw.lines = append(iw.lines, []rune{})
	iw.currentLine++

	// Ident the current line according to the indent level.
	for i := 0; i < iw.indent; i++ {
		iw.appendRune('\t')
	}
}

func (iw *IndentedWriter) appendRune(r rune) {
	iw.lines[iw.currentLine] = append(iw.lines[iw.currentLine], r)
}

func writeStringf(w io.Writer, format string, args ...interface{}) {
	w.Write([]byte(fmt.Sprintf(format, args...)))
}
