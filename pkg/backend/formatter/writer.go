package formatter

import (
	"fmt"
	"io"
	"strings"
)

type IndentedWriter struct {
	indent int
	lines  [][]rune
	linei  int
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

	// If the current line is empty (or only has tabs), we add the indentation.
	line := strings.TrimSpace(string(iw.currentLine()))
	if line == "" {
		iw.clearLine(iw.linei)
		iw.appendIndentation()
	}
}

func (iw *IndentedWriter) MustDecreateIndent() {
	iw.indent--
	if iw.indent < 0 {
		panic("Wrong scoped indent!")
	}

	// If the current line is empty (or only has tabs), we add the indentation.
	line := strings.TrimSpace(string(iw.currentLine()))
	if line == "" {
		iw.clearLine(iw.linei)
		iw.appendIndentation()
	}
}

func (iw *IndentedWriter) Write(p []byte) (int, error) {
	return iw.write(p, true)
}

func (iw *IndentedWriter) WriteIgnoreIndent(p []byte) (int, error) {
	return iw.write(p, false)
}

func (iw *IndentedWriter) write(p []byte, indent bool) (int, error) {
	data := string(p)
	for _, r := range data {
		if indent {
			if r == '\n' {
				iw.handleNewLine()
				continue
			}
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

func (iw *IndentedWriter) currentLine() []rune {
	return iw.lines[iw.linei]
}

func (iw *IndentedWriter) clearLine(i int) {
	iw.lines[i] = []rune{}
}

func (iw *IndentedWriter) handleNewLine() {
	// Add a new line
	iw.lines = append(iw.lines, []rune{})
	iw.linei++

	iw.appendIndentation()
}

func (iw *IndentedWriter) appendIndentation() {
	// Ident the current line according to the indent level.
	for i := 0; i < iw.indent; i++ {
		iw.appendRune('\t')
	}
}

func (iw *IndentedWriter) appendRune(r rune) {
	iw.lines[iw.linei] = append(iw.lines[iw.linei], r)
}

func writeStringf(w io.Writer, format string, args ...interface{}) {
	w.Write([]byte(fmt.Sprintf(format, args...)))
}

func writeStringfNoIndent(w io.Writer, format string, args ...interface{}) {
	str := fmt.Sprintf(format, args...)
	if iw, ok := w.(*IndentedWriter); ok {
		iw.WriteIgnoreIndent([]byte(str))
	} else {
		w.Write([]byte(str))
	}
}
